package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	repository *Repository
}

func NewHandler(repository *Repository) *Handler {
	return &Handler{repository: repository}
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request SignupRequest

	if err := decoder.Decode(&request); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"リクエストの形式が正しくありません",
		)
		return
	}

	name := strings.TrimSpace(request.Name)
	email := strings.ToLower(strings.TrimSpace(request.Email))

	if message := validateSignup(request, name, email); message != "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", message)
		return
	}

	birthday, err := time.Parse("2006-01-02", request.Birthday)
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_BIRTHDAY",
			"生年月日の形式が正しくありません",
		)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(request.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"サーバーエラーが発生しました",
		)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	createdUser, err := h.repository.Create(
		ctx,
		name,
		email,
		birthday,
		string(passwordHash),
	)
	if err != nil {
		if isUniqueViolation(err) {
			writeError(
				w,
				http.StatusConflict,
				"EMAIL_ALREADY_EXISTS",
				"このメールアドレスは既に登録されています",
			)
			return
		}

		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"サーバーエラーが発生しました",
		)
		return
	}

	response := SignupResponse{
		ID:        createdUser.ID,
		Name:      createdUser.Name,
		Email:     createdUser.Email,
		Birthday:  createdUser.Birthday.Format("2006-01-02"),
		CreatedAt: createdUser.CreatedAt.Format(time.RFC3339),
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request LoginRequest

	if err := decoder.Decode(&request); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"リクエストの形式が正しくありません",
		)
		return
	}

	identifier := normalizeLoginIdentifier(request)

	if message := validateLogin(identifier, request.Password); message != "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", message)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	foundUser, err := h.repository.FindByLoginIdentifier(ctx, identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeInvalidCredentials(w)
			return
		}

		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"サーバーエラーが発生しました",
		)
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(foundUser.PasswordHash),
		[]byte(request.Password),
	); err != nil {
		writeInvalidCredentials(w)
		return
	}

	response := LoginResponse{
		ID:        foundUser.ID,
		Name:      foundUser.Name,
		Email:     foundUser.Email,
		Birthday:  foundUser.Birthday.Format("2006-01-02"),
		CreatedAt: foundUser.CreatedAt.Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, response)
}

func validateSignup(request SignupRequest, name, email string) string {
	if name == "" {
		return "名前を入力してください"
	}

	if len([]rune(name)) > 50 {
		return "名前は50文字以内で入力してください"
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return "正しいメールアドレスを入力してください"
	}

	if len(request.Password) < 8 {
		return "パスワードは8文字以上で入力してください"
	}

	if request.Birthday == "" {
		return "生年月日を入力してください"
	}

	return ""
}

func normalizeLoginIdentifier(request LoginRequest) string {
	if strings.TrimSpace(request.Identifier) != "" {
		return strings.TrimSpace(request.Identifier)
	}

	if strings.TrimSpace(request.Email) != "" {
		return strings.ToLower(strings.TrimSpace(request.Email))
	}

	return strings.TrimSpace(request.Name)
}

func validateLogin(identifier, password string) string {
	if identifier == "" {
		return "ユーザー名またはメールアドレスを入力してください"
	}

	if password == "" {
		return "パスワードを入力してください"
	}

	return ""
}

func isUniqueViolation(err error) bool {
	var pgError *pgconn.PgError

	if errors.As(err, &pgError) {
		return pgError.Code == "23505"
	}

	return false
}

func writeInvalidCredentials(w http.ResponseWriter) {
	writeError(
		w,
		http.StatusUnauthorized,
		"INVALID_CREDENTIALS",
		"ユーザー名またはパスワードが違います",
	)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{
		Error: errorBody{
			Code:    code,
			Message: message,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}
