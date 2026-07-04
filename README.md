# Twitter Clone Backend

Twitter風アプリケーションのバックエンドAPIです。Goで構築し、ユーザー情報をPostgreSQLへ保存します。

現在は新規登録APIとヘルスチェックAPIを実装しています。パスワードはbcryptでハッシュ化してから保存します。

## 使用技術

- Go 1.25
- PostgreSQL 16
- Docker / Docker Compose

## Dockerで起動する

PostgreSQLとGo APIをまとめて起動します。

```bash
docker compose up --build
```

起動するサービスは以下の通りです。

| サービス | URL・ポート |
| --- | --- |
| Go API | `http://localhost:8080` |
| PostgreSQL | `localhost:5432` |

## データベース設定

Docker Composeの開発用設定は以下です。

| 項目 | 値 |
| --- | --- |
| Host | ローカルからは `localhost`、Docker内からは `postgres` |
| Port | `5432` |
| Database | `twitter` |
| User | `twitter` |
| Password | `twitter_password` |

接続URL：

```text
postgres://twitter:twitter_password@localhost:5432/twitter?sslmode=disable
```

初回起動時に `migrations/001_create_users.sql` が実行され、`users` テーブルが作成されます。

## 入力ルール

- 名前は必須、50文字以内
- メールアドレスは有効な形式であること
- メールアドレスは重複不可
- 生年月日は `YYYY-MM-DD` 形式
- パスワードは8文字以上

## 動作確認

```bash
curl -i -X POST http://localhost:8080/api/signup \
  -H "Content-Type: application/json" \
  -d '{
    "name": "テストユーザー",
    "email": "test@example.com",
    "birthday": "2000-01-01",
    "password": "password123"
  }'
```

## 主なディレクトリ構成

```text
.
├── cmd/api/main.go               # APIサーバーの起動とルーティング
├── internal/
│   ├── config/config.go          # 環境変数の読み込み
│   ├── database/database.go      # PostgreSQL接続
│   └── user/
│       ├── handler.go            # HTTPリクエストとレスポンス
│       ├── models.go             # API・ユーザーの型定義
│       └── repository.go         # ユーザーのDB操作
├── migrations/
│   └── 001_create_users.sql      # usersテーブル作成
├── Dockerfile
└── compose.yml
```