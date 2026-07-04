# Backend Article API

Microservice API artikel sederhana berbasis Go, MySQL, dan Goose migration.

## Stack
- Go
- `net/http`
- `github.com/go-chi/chi/v5`
- MySQL
- `github.com/jmoiron/sqlx`
- `github.com/pressly/goose/v3`
- `github.com/go-playground/validator/v10`

## Fitur
- Create article
- List article dengan `limit` dan `offset`
- Detail article by ID
- Update article dengan `PUT`
- Partial update article dengan `PATCH`
- Soft delete article dengan `DELETE` menjadi `status = "thrash"`
- Validasi request body
- Postman collection siap pakai

## Struktur Project
```text
cmd/api/main.go                # entrypoint server
internal/config/config.go      # load config dari env
internal/database/mysql.go     # koneksi MySQL
internal/httpx/response.go     # helper JSON response
internal/article/              # domain article
migrations/                    # file migration Goose
postman/                       # Postman collection
```

## Kebutuhan
- Go 1.26+
- MySQL 8+

## Environment
Copy file contoh env:

```bash
cp .env.example .env
```

Isi `.env`:

```env
APP_PORT=8080
APP_ENV=development
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=backend_article
DB_USER=root
DB_PASSWORD=
DB_PARAMS=parseTime=true&loc=UTC
```

## Install Dependency
```bash
go mod tidy
```

## Migration
### Validasi migration
```bash
go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations validate
```

### Jalankan migration ke MySQL
Set env Goose lalu jalankan `up`:

```bash
export GOOSE_DRIVER=mysql
export GOOSE_DBSTRING="root:@tcp(127.0.0.1:3306)/backend_article?parseTime=true&loc=UTC"
export GOOSE_MIGRATION_DIR=./migrations

go run github.com/pressly/goose/v3/cmd/goose@latest up
```

### Cek status migration
```bash
go run github.com/pressly/goose/v3/cmd/goose@latest status
```

## Menjalankan Aplikasi
```bash
go run ./cmd/api
```

Server jalan di:
```text
http://localhost:8080
```

## Validasi Request
### Create dan PUT
- `title`: wajib, minimal 20 karakter, maksimal 200 karakter
- `content`: wajib, minimal 200 karakter
- `category`: wajib, minimal 3 karakter, maksimal 100 karakter
- `status`: wajib, hanya `publish`, `draft`, `thrash`

### PATCH
- body tidak boleh kosong
- field yang dikirim tetap harus lolos rule validasi di atas

## API Endpoint

### 1. Create article
**POST** `/article`

Request:
```json
{
  "title": "This is an article title with enough length",
  "content": "This article content is intentionally made very long so it passes the minimum two hundred character validation rule. This article content is intentionally made very long so it passes the minimum two hundred character validation rule.",
  "category": "Technology",
  "status": "publish"
}
```

Success response:
```json
{
  "id": 1,
  "title": "This is an article title with enough length",
  "content": "This article content is intentionally made very long so it passes the minimum two hundred character validation rule. This article content is intentionally made very long so it passes the minimum two hundred character validation rule.",
  "category": "Technology",
  "status": "publish",
  "created_date": "2026-07-04T01:00:00Z",
  "updated_date": "2026-07-04T01:00:00Z"
}
```

### 2. List article
**GET** `/article/{limit}/{offset}`

Contoh:
```text
GET /article/10/0
```

Success response:
```json
[
  {
    "id": 1,
    "title": "This is an article title with enough length",
    "content": "This article content is intentionally made very long so it passes the minimum two hundred character validation rule. This article content is intentionally made very long so it passes the minimum two hundred character validation rule.",
    "category": "Technology",
    "status": "publish",
    "created_date": "2026-07-04T01:00:00Z",
    "updated_date": "2026-07-04T01:00:00Z"
  }
]
```

### 3. Detail article
**GET** `/article/{id}`

Contoh:
```text
GET /article/1
```

### 4. Replace article
**PUT** `/article/{id}`

Request:
```json
{
  "title": "This is updated article title text",
  "content": "This updated article content is intentionally made very long so it passes the minimum two hundred character validation rule. This updated article content is intentionally made very long so it passes the minimum two hundred character validation rule.",
  "category": "Business",
  "status": "draft"
}
```

### 5. Partial update article
**PATCH** `/article/{id}`

Request:
```json
{
  "title": "This patch title is valid"
}
```

Atau:
```json
{
  "status": "thrash"
}
```

### 6. Soft delete article
**DELETE** `/article/{id}`

Behavior:
- record tidak dihapus fisik dari database
- `status` akan diubah menjadi `thrash`

Success response:
```json
{
  "message": "article deleted"
}
```

## Format Error Response
```json
{
  "message": "validation failed",
  "errors": {
    "title": [
      "must be at least 20 characters"
    ]
  }
}
```

## Testing
Jalankan seluruh test:

```bash
go test ./...
```

## Format dan Build
```bash
gofmt -w cmd internal
go build ./cmd/api
```

## Postman
Import file berikut ke Postman:

```text
postman/article-api.postman_collection.json
```

Collection sudah berisi:
- variable `base_url`
- variable `article_id`
- variable `limit`
- variable `offset`
- contoh request body valid untuk create, put, dan patch
