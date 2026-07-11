package post

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"twitter_golang_backend/internal/auth"
)

type Handler struct {
	repository *Repository
	uploadDir  string
}

func NewHandler(repository *Repository, uploadDir string) *Handler {
	return &Handler{repository: repository, uploadDir: uploadDir}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 6<<20)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_FORM", "投稿データが不正、または画像が大きすぎます")
		return
	}
	doc := strings.TrimSpace(r.FormValue("doc"))
	if utf8.RuneCountInString(doc) > 140 {
		writeError(w, http.StatusBadRequest, "DOC_TOO_LONG", "投稿内容は140文字以内にしてください")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		writeError(w, http.StatusBadRequest, "INVALID_IMAGE", "画像を読み込めません")
		return
	}
	if file != nil {
		defer file.Close()
	}
	if doc == "" && file == nil {
		writeError(w, http.StatusBadRequest, "CONTENT_REQUIRED", "投稿内容または画像を指定してください")
		return
	}

	var imageURL *string
	if file != nil {
		url, err := h.saveImage(file, header)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_IMAGE", err.Error())
			return
		}
		imageURL = &url
	}

	userID, _ := auth.UserID(r.Context())
	ctx := r.Context()
	created, err := h.repository.Create(ctx, userID, doc, imageURL)
	if err != nil {
		if imageURL != nil {
			_ = os.Remove(filepath.Join(h.uploadDir, filepath.Base(*imageURL)))
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "投稿を保存できませんでした")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) saveImage(file multipart.File, header *multipart.FileHeader) (string, error) {
	if header.Size > 5<<20 {
		return "", fmt.Errorf("画像は5MB以下にしてください")
	}
	first := make([]byte, 512)
	n, err := file.Read(first)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("画像を読み込めません")
	}
	contentType := http.DetectContentType(first[:n])
	extensions := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/gif": ".gif", "image/webp": ".webp"}
	ext, ok := extensions[contentType]
	if !ok {
		return "", fmt.Errorf("JPEG、PNG、GIF、WebP画像を選択してください")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("画像を読み込めません")
	}
	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		return "", fmt.Errorf("画像を保存できません")
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("画像を保存できません")
	}
	name := hex.EncodeToString(random) + ext
	destination, err := os.OpenFile(filepath.Join(h.uploadDir, name), os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
	if err != nil {
		return "", fmt.Errorf("画像を保存できません")
	}
	defer destination.Close()
	if _, err := io.Copy(destination, file); err != nil {
		return "", fmt.Errorf("画像を保存できません")
	}
	return "/uploads/" + name, nil
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
