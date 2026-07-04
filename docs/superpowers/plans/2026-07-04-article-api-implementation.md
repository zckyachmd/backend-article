# Article API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a simple Go article API microservice backed by MySQL with Goose migrations, validation, soft delete, tests, and a Postman collection.

**Architecture:** Use a shallow handler → service → repository structure under `internal/article`, with `cmd/api/main.go` bootstrapping config, DB, router, and server startup. Keep SQL explicit in repository functions, validation near request DTO handling, and response helpers in a small `internal/httpx` package.

**Tech Stack:** Go, `net/http`, `github.com/go-chi/chi/v5`, `database/sql`, `github.com/jmoiron/sqlx`, `github.com/go-sql-driver/mysql`, `github.com/pressly/goose/v3`, `github.com/go-playground/validator/v10`, `github.com/joho/godotenv`, `github.com/stretchr/testify`

---

## File Map

### Files to create
- `cmd/api/main.go` — app entrypoint, config load, DB setup, router startup
- `internal/config/config.go` — environment-backed app config
- `internal/database/mysql.go` — SQLX MySQL connection helper
- `internal/httpx/response.go` — JSON success and error response helpers
- `internal/article/model.go` — persistent article model and status constants
- `internal/article/dto.go` — request/response DTOs and payload validation helpers
- `internal/article/repository.go` — SQL queries for create/list/detail/update/soft delete
- `internal/article/service.go` — business rules and repository orchestration
- `internal/article/handler.go` — HTTP handlers for all article endpoints
- `internal/article/routes.go` — route registration for article endpoints
- `internal/article/dto_test.go` — payload validation tests
- `internal/article/service_test.go` — service-level tests with repository stubs
- `internal/article/handler_test.go` — handler tests using `httptest`
- `migrations/202607040001_create_posts_table.sql` — initial Goose migration
- `postman/article-api.postman_collection.json` — Postman collection with variables
- `.env.example` — local config example

### Files to modify
- `go.mod` — add runtime and test dependencies
- `.gitignore` — ignore local `.env` and build artifacts if needed

## Task 1: Add module dependencies and local config example

**Files:**
- Modify: `go.mod`
- Modify: `.gitignore`
- Create: `.env.example`

- [ ] **Step 1: Add required dependencies to `go.mod`**

```go
module backend-article

go 1.26.4

require (
    github.com/go-chi/chi/v5 v5.2.3
    github.com/go-playground/validator/v10 v10.27.0
    github.com/go-sql-driver/mysql v1.9.3
    github.com/joho/godotenv v1.5.1
    github.com/jmoiron/sqlx v1.4.0
    github.com/pressly/goose/v3 v3.26.0
    github.com/stretchr/testify v1.10.0
)
```

- [ ] **Step 2: Download modules and record checksums**

Run: `go mod tidy`
Expected: `go.mod` and `go.sum` updated with no errors

- [ ] **Step 3: Ignore local environment file and binary output**

Add to `.gitignore` if missing:

```gitignore
.env
bin/
```

- [ ] **Step 4: Create local environment example**

Create `.env.example`:

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

- [ ] **Step 5: Verify module graph compiles at dependency level**

Run: `go test ./...`
Expected: PASS or `?[no test files]` for packages that exist so far

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum .gitignore .env.example
git commit -m "chore: add article api dependencies"
```

## Task 2: Create configuration and database bootstrap

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/database/mysql.go`
- Test: `go test ./...`

- [ ] **Step 1: Write minimal config loader code**

Create `internal/config/config.go`:

```go
package config

import (
    "fmt"
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    AppPort    string
    AppEnv     string
    DBHost     string
    DBPort     string
    DBName     string
    DBUser     string
    DBPassword string
    DBParams   string
}

func Load() (Config, error) {
    _ = godotenv.Load()

    cfg := Config{
        AppPort:    getEnv("APP_PORT", "8080"),
        AppEnv:     getEnv("APP_ENV", "development"),
        DBHost:     getEnv("DB_HOST", "127.0.0.1"),
        DBPort:     getEnv("DB_PORT", "3306"),
        DBName:     os.Getenv("DB_NAME"),
        DBUser:     os.Getenv("DB_USER"),
        DBPassword: os.Getenv("DB_PASSWORD"),
        DBParams:   getEnv("DB_PARAMS", "parseTime=true&loc=UTC"),
    }

    if cfg.DBName == "" || cfg.DBUser == "" {
        return Config{}, fmt.Errorf("DB_NAME and DB_USER are required")
    }

    return cfg, nil
}

func getEnv(key, fallback string) string {
    value := os.Getenv(key)
    if value == "" {
        return fallback
    }

    return value
}
```

- [ ] **Step 2: Write MySQL connection helper**

Create `internal/database/mysql.go`:

```go
package database

import (
    "fmt"
    "time"

    "backend-article/internal/config"

    _ "github.com/go-sql-driver/mysql"
    "github.com/jmoiron/sqlx"
)

func NewMySQL(cfg config.Config) (*sqlx.DB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBParams)

    db, err := sqlx.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }

    db.SetConnMaxLifetime(5 * time.Minute)
    db.SetMaxIdleConns(5)
    db.SetMaxOpenConns(10)

    if err := db.Ping(); err != nil {
        _ = db.Close()
        return nil, err
    }

    return db, nil
}
```

- [ ] **Step 3: Run package tests to catch compile errors**

Run: `go test ./...`
Expected: PASS or no test files

- [ ] **Step 4: Commit**

```bash
git add internal/config/config.go internal/database/mysql.go
git commit -m "feat: add config and mysql bootstrap"
```

## Task 3: Add migration and migration workflow note

**Files:**
- Create: `migrations/202607040001_create_posts_table.sql`
- Test: `migrations/202607040001_create_posts_table.sql`

- [ ] **Step 1: Create initial Goose migration**

Create `migrations/202607040001_create_posts_table.sql`:

```sql
-- +goose Up
CREATE TABLE posts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    category VARCHAR(100) NOT NULL,
    status VARCHAR(100) NOT NULL,
    created_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_posts_status (status),
    INDEX idx_posts_category (category)
);

-- +goose Down
DROP TABLE posts;
```

- [ ] **Step 2: Validate migration syntax with Goose**

Run: `go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations validate`
Expected: validation succeeds for the migration file

- [ ] **Step 3: Commit**

```bash
git add migrations/202607040001_create_posts_table.sql
git commit -m "feat: add posts migration"
```

## Task 4: Write failing validation tests for article DTOs

**Files:**
- Create: `internal/article/dto.go`
- Create: `internal/article/dto_test.go`
- Test: `internal/article/dto_test.go`

- [ ] **Step 1: Write validation tests first**

Create `internal/article/dto_test.go`:

```go
package article

import (
    "strings"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestCreateArticleRequestValidate_Success(t *testing.T) {
    req := CreateArticleRequest{
        Title:    strings.Repeat("T", 20),
        Content:  strings.Repeat("C", 200),
        Category: "Tech",
        Status:   StatusPublish,
    }

    errs := req.Validate()

    assert.Empty(t, errs)
}

func TestCreateArticleRequestValidate_Failure(t *testing.T) {
    req := CreateArticleRequest{
        Title:    "short",
        Content:  "tiny",
        Category: "ab",
        Status:   "unknown",
    }

    errs := req.Validate()

    assert.Contains(t, errs, "title")
    assert.Contains(t, errs, "content")
    assert.Contains(t, errs, "category")
    assert.Contains(t, errs, "status")
}

func TestPatchArticleRequestValidate_RejectEmptyBody(t *testing.T) {
    req := PatchArticleRequest{}

    errs := req.Validate()

    assert.Contains(t, errs, "body")
}

func TestPatchArticleRequestValidate_FieldRules(t *testing.T) {
    title := "too short"
    category := "ab"
    status := "deleted"
    content := "tiny"

    req := PatchArticleRequest{
        Title:    &title,
        Content:  &content,
        Category: &category,
        Status:   &status,
    }

    errs := req.Validate()

    assert.Contains(t, errs, "title")
    assert.Contains(t, errs, "content")
    assert.Contains(t, errs, "category")
    assert.Contains(t, errs, "status")
}
```

- [ ] **Step 2: Run DTO test to verify failure**

Run: `go test ./internal/article -run TestCreateArticleRequestValidate_Success -v`
Expected: FAIL because DTO types and methods do not exist yet

- [ ] **Step 3: Implement DTOs and validation**

Create `internal/article/dto.go`:

```go
package article

import (
    "strings"
)

type CreateArticleRequest struct {
    Title    string `json:"title"`
    Content  string `json:"content"`
    Category string `json:"category"`
    Status   string `json:"status"`
}

type UpdateArticleRequest = CreateArticleRequest

type PatchArticleRequest struct {
    Title    *string `json:"title"`
    Content  *string `json:"content"`
    Category *string `json:"category"`
    Status   *string `json:"status"`
}

type ArticleResponse struct {
    ID          int64  `json:"id"`
    Title       string `json:"title"`
    Content     string `json:"content"`
    Category    string `json:"category"`
    Status      string `json:"status"`
    CreatedDate string `json:"created_date"`
    UpdatedDate string `json:"updated_date"`
}

func (r CreateArticleRequest) Validate() map[string][]string {
    errs := map[string][]string{}

    if len(strings.TrimSpace(r.Title)) < 20 {
        errs["title"] = append(errs["title"], "must be at least 20 characters")
    }
    if len(strings.TrimSpace(r.Content)) < 200 {
        errs["content"] = append(errs["content"], "must be at least 200 characters")
    }
    if len(strings.TrimSpace(r.Category)) < 3 {
        errs["category"] = append(errs["category"], "must be at least 3 characters")
    }
    if !isAllowedStatus(r.Status) {
        errs["status"] = append(errs["status"], "must be one of publish, draft, thrash")
    }

    return errs
}

func (r PatchArticleRequest) Validate() map[string][]string {
    errs := map[string][]string{}

    if r.Title == nil && r.Content == nil && r.Category == nil && r.Status == nil {
        errs["body"] = append(errs["body"], "at least one field is required")
        return errs
    }

    if r.Title != nil && len(strings.TrimSpace(*r.Title)) < 20 {
        errs["title"] = append(errs["title"], "must be at least 20 characters")
    }
    if r.Content != nil && len(strings.TrimSpace(*r.Content)) < 200 {
        errs["content"] = append(errs["content"], "must be at least 200 characters")
    }
    if r.Category != nil && len(strings.TrimSpace(*r.Category)) < 3 {
        errs["category"] = append(errs["category"], "must be at least 3 characters")
    }
    if r.Status != nil && !isAllowedStatus(*r.Status) {
        errs["status"] = append(errs["status"], "must be one of publish, draft, thrash")
    }

    return errs
}

func isAllowedStatus(value string) bool {
    switch strings.ToLower(strings.TrimSpace(value)) {
    case StatusPublish, StatusDraft, StatusThrash:
        return true
    default:
        return false
    }
}
```

- [ ] **Step 4: Add status constants and model type needed by DTO package**

Create `internal/article/model.go`:

```go
package article

import "time"

const (
    StatusPublish = "publish"
    StatusDraft   = "draft"
    StatusThrash  = "thrash"
)

type Article struct {
    ID          int64     `db:"id"`
    Title       string    `db:"title"`
    Content     string    `db:"content"`
    Category    string    `db:"category"`
    Status      string    `db:"status"`
    CreatedDate time.Time `db:"created_date"`
    UpdatedDate time.Time `db:"updated_date"`
}
```

- [ ] **Step 5: Run DTO tests to verify they pass**

Run: `go test ./internal/article -run Validate -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/article/dto.go internal/article/dto_test.go internal/article/model.go
git commit -m "feat: add article dto validation"
```

## Task 5: Add HTTP JSON helpers

**Files:**
- Create: `internal/httpx/response.go`
- Test: `go test ./...`

- [ ] **Step 1: Create response helper package**

Create `internal/httpx/response.go`:

```go
package httpx

import (
    "encoding/json"
    "net/http"
)

type ErrorResponse struct {
    Message string              `json:"message"`
    Errors  map[string][]string `json:"errors,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, message string, errors map[string][]string) {
    WriteJSON(w, status, ErrorResponse{
        Message: message,
        Errors:  errors,
    })
}
```

- [ ] **Step 2: Run package tests to verify compile**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/httpx/response.go
git commit -m "feat: add http json helpers"
```

## Task 6: Write failing service tests for business rules

**Files:**
- Create: `internal/article/service.go`
- Create: `internal/article/service_test.go`
- Test: `internal/article/service_test.go`

- [ ] **Step 1: Write service tests first**

Create `internal/article/service_test.go`:

```go
package article

import (
    "context"
    "errors"
    "strings"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

type repositoryStub struct {
    createFn     func(context.Context, CreateArticleRequest) (Article, error)
    listFn       func(context.Context, int, int) ([]Article, error)
    getByIDFn    func(context.Context, int64) (Article, error)
    updateFn     func(context.Context, int64, UpdateArticleRequest) (Article, error)
    patchFn      func(context.Context, int64, PatchArticleRequest) (Article, error)
    softDeleteFn func(context.Context, int64) error
}

func (r repositoryStub) Create(ctx context.Context, req CreateArticleRequest) (Article, error) {
    return r.createFn(ctx, req)
}

func (r repositoryStub) List(ctx context.Context, limit, offset int) ([]Article, error) {
    return r.listFn(ctx, limit, offset)
}

func (r repositoryStub) GetByID(ctx context.Context, id int64) (Article, error) {
    return r.getByIDFn(ctx, id)
}

func (r repositoryStub) Update(ctx context.Context, id int64, req UpdateArticleRequest) (Article, error) {
    return r.updateFn(ctx, id, req)
}

func (r repositoryStub) Patch(ctx context.Context, id int64, req PatchArticleRequest) (Article, error) {
    return r.patchFn(ctx, id, req)
}

func (r repositoryStub) SoftDelete(ctx context.Context, id int64) error {
    return r.softDeleteFn(ctx, id)
}

func TestServiceList_RejectsLimitAbove100(t *testing.T) {
    svc := NewService(repositoryStub{})

    _, err := svc.List(context.Background(), 101, 0)

    assert.EqualError(t, err, "limit must be less than or equal to 100")
}

func TestServiceDelete_ConvertsToThrash(t *testing.T) {
    called := false
    svc := NewService(repositoryStub{
        softDeleteFn: func(ctx context.Context, id int64) error {
            called = true
            return nil
        },
    })

    err := svc.Delete(context.Background(), 9)

    assert.NoError(t, err)
    assert.True(t, called)
}

func TestServiceGetByID_ReturnsNotFound(t *testing.T) {
    svc := NewService(repositoryStub{
        getByIDFn: func(ctx context.Context, id int64) (Article, error) {
            return Article{}, ErrNotFound
        },
    })

    _, err := svc.GetByID(context.Background(), 1)

    assert.ErrorIs(t, err, ErrNotFound)
}

func TestToResponse_FormatsTimestampsInUTC(t *testing.T) {
    article := Article{
        ID:          1,
        Title:       strings.Repeat("T", 20),
        Content:     strings.Repeat("C", 200),
        Category:    "Tech",
        Status:      StatusPublish,
        CreatedDate: time.Date(2026, 7, 4, 1, 2, 3, 0, time.FixedZone("WIB", 7*3600)),
        UpdatedDate: time.Date(2026, 7, 4, 4, 5, 6, 0, time.FixedZone("WIB", 7*3600)),
    }

    resp := ToResponse(article)

    assert.Equal(t, "2026-07-03T18:02:03Z", resp.CreatedDate)
    assert.Equal(t, "2026-07-03T21:05:06Z", resp.UpdatedDate)
}

func TestServiceDelete_PropagatesRepositoryError(t *testing.T) {
    svc := NewService(repositoryStub{
        softDeleteFn: func(ctx context.Context, id int64) error {
            return errors.New("db down")
        },
    })

    err := svc.Delete(context.Background(), 9)

    assert.EqualError(t, err, "db down")
}
```

- [ ] **Step 2: Run service test to verify failure**

Run: `go test ./internal/article -run TestServiceList_RejectsLimitAbove100 -v`
Expected: FAIL because service types do not exist yet

- [ ] **Step 3: Implement service and repository contract**

Create `internal/article/service.go`:

```go
package article

import (
    "context"
    "errors"
)

var ErrNotFound = errors.New("article not found")

type Repository interface {
    Create(context.Context, CreateArticleRequest) (Article, error)
    List(context.Context, int, int) ([]Article, error)
    GetByID(context.Context, int64) (Article, error)
    Update(context.Context, int64, UpdateArticleRequest) (Article, error)
    Patch(context.Context, int64, PatchArticleRequest) (Article, error)
    SoftDelete(context.Context, int64) error
}

type Service struct {
    repo Repository
}

func NewService(repo Repository) Service {
    return Service{repo: repo}
}

func (s Service) Create(ctx context.Context, req CreateArticleRequest) (ArticleResponse, error) {
    article, err := s.repo.Create(ctx, req)
    if err != nil {
        return ArticleResponse{}, err
    }

    return ToResponse(article), nil
}

func (s Service) List(ctx context.Context, limit, offset int) ([]ArticleResponse, error) {
    if limit <= 0 {
        return nil, errors.New("limit must be greater than 0")
    }
    if limit > 100 {
        return nil, errors.New("limit must be less than or equal to 100")
    }
    if offset < 0 {
        return nil, errors.New("offset must be greater than or equal to 0")
    }

    articles, err := s.repo.List(ctx, limit, offset)
    if err != nil {
        return nil, err
    }

    responses := make([]ArticleResponse, 0, len(articles))
    for _, article := range articles {
        responses = append(responses, ToResponse(article))
    }

    return responses, nil
}

func (s Service) GetByID(ctx context.Context, id int64) (ArticleResponse, error) {
    article, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return ArticleResponse{}, err
    }

    return ToResponse(article), nil
}

func (s Service) Update(ctx context.Context, id int64, req UpdateArticleRequest) (ArticleResponse, error) {
    article, err := s.repo.Update(ctx, id, req)
    if err != nil {
        return ArticleResponse{}, err
    }

    return ToResponse(article), nil
}

func (s Service) Patch(ctx context.Context, id int64, req PatchArticleRequest) (ArticleResponse, error) {
    article, err := s.repo.Patch(ctx, id, req)
    if err != nil {
        return ArticleResponse{}, err
    }

    return ToResponse(article), nil
}

func (s Service) Delete(ctx context.Context, id int64) error {
    return s.repo.SoftDelete(ctx, id)
}

func ToResponse(article Article) ArticleResponse {
    return ArticleResponse{
        ID:          article.ID,
        Title:       article.Title,
        Content:     article.Content,
        Category:    article.Category,
        Status:      article.Status,
        CreatedDate: article.CreatedDate.UTC().Format("2006-01-02T15:04:05Z"),
        UpdatedDate: article.UpdatedDate.UTC().Format("2006-01-02T15:04:05Z"),
    }
}
```

- [ ] **Step 4: Run service tests to verify they pass**

Run: `go test ./internal/article -run TestService -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/article/service.go internal/article/service_test.go
git commit -m "feat: add article service"
```

## Task 7: Implement repository SQL access

**Files:**
- Create: `internal/article/repository.go`
- Test: `go test ./...`

- [ ] **Step 1: Write repository implementation**

Create `internal/article/repository.go`:

```go
package article

import (
    "context"
    "database/sql"
    "fmt"
    "strings"

    "github.com/jmoiron/sqlx"
)

type SQLRepository struct {
    db *sqlx.DB
}

func NewRepository(db *sqlx.DB) SQLRepository {
    return SQLRepository{db: db}
}

func (r SQLRepository) Create(ctx context.Context, req CreateArticleRequest) (Article, error) {
    result, err := r.db.ExecContext(ctx, `INSERT INTO posts (title, content, category, status) VALUES (?, ?, ?, ?)`, req.Title, req.Content, req.Category, strings.ToLower(strings.TrimSpace(req.Status)))
    if err != nil {
        return Article{}, err
    }

    id, err := result.LastInsertId()
    if err != nil {
        return Article{}, err
    }

    return r.GetByID(ctx, id)
}

func (r SQLRepository) List(ctx context.Context, limit, offset int) ([]Article, error) {
    var articles []Article
    err := r.db.SelectContext(ctx, &articles, `SELECT id, title, content, category, status, created_date, updated_date FROM posts ORDER BY id DESC LIMIT ? OFFSET ?`, limit, offset)
    return articles, err
}

func (r SQLRepository) GetByID(ctx context.Context, id int64) (Article, error) {
    var article Article
    err := r.db.GetContext(ctx, &article, `SELECT id, title, content, category, status, created_date, updated_date FROM posts WHERE id = ?`, id)
    if err != nil {
        if err == sql.ErrNoRows {
            return Article{}, ErrNotFound
        }
        return Article{}, err
    }

    return article, nil
}

func (r SQLRepository) Update(ctx context.Context, id int64, req UpdateArticleRequest) (Article, error) {
    result, err := r.db.ExecContext(ctx, `UPDATE posts SET title = ?, content = ?, category = ?, status = ? WHERE id = ?`, req.Title, req.Content, req.Category, strings.ToLower(strings.TrimSpace(req.Status)), id)
    if err != nil {
        return Article{}, err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return Article{}, err
    }
    if rowsAffected == 0 {
        return Article{}, ErrNotFound
    }

    return r.GetByID(ctx, id)
}

func (r SQLRepository) Patch(ctx context.Context, id int64, req PatchArticleRequest) (Article, error) {
    sets := make([]string, 0, 4)
    args := make([]any, 0, 5)

    if req.Title != nil {
        sets = append(sets, "title = ?")
        args = append(args, *req.Title)
    }
    if req.Content != nil {
        sets = append(sets, "content = ?")
        args = append(args, *req.Content)
    }
    if req.Category != nil {
        sets = append(sets, "category = ?")
        args = append(args, *req.Category)
    }
    if req.Status != nil {
        sets = append(sets, "status = ?")
        args = append(args, strings.ToLower(strings.TrimSpace(*req.Status)))
    }

    query := fmt.Sprintf("UPDATE posts SET %s WHERE id = ?", strings.Join(sets, ", "))
    args = append(args, id)

    result, err := r.db.ExecContext(ctx, query, args...)
    if err != nil {
        return Article{}, err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return Article{}, err
    }
    if rowsAffected == 0 {
        return Article{}, ErrNotFound
    }

    return r.GetByID(ctx, id)
}

func (r SQLRepository) SoftDelete(ctx context.Context, id int64) error {
    result, err := r.db.ExecContext(ctx, `UPDATE posts SET status = ? WHERE id = ?`, StatusThrash, id)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        return ErrNotFound
    }

    return nil
}
```

- [ ] **Step 2: Run all tests to verify compile and behavior**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/article/repository.go
git commit -m "feat: add article repository"
```

## Task 8: Write failing handler tests for HTTP behavior

**Files:**
- Create: `internal/article/handler.go`
- Create: `internal/article/routes.go`
- Create: `internal/article/handler_test.go`
- Test: `internal/article/handler_test.go`

- [ ] **Step 1: Write handler tests first**

Create `internal/article/handler_test.go`:

```go
package article

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/stretchr/testify/assert"
)

type serviceStub struct {
    createFn  func(context.Context, CreateArticleRequest) (ArticleResponse, error)
    listFn    func(context.Context, int, int) ([]ArticleResponse, error)
    getFn     func(context.Context, int64) (ArticleResponse, error)
    updateFn  func(context.Context, int64, UpdateArticleRequest) (ArticleResponse, error)
    patchFn   func(context.Context, int64, PatchArticleRequest) (ArticleResponse, error)
    deleteFn  func(context.Context, int64) error
}

func (s serviceStub) Create(ctx context.Context, req CreateArticleRequest) (ArticleResponse, error) {
    return s.createFn(ctx, req)
}
func (s serviceStub) List(ctx context.Context, limit, offset int) ([]ArticleResponse, error) {
    return s.listFn(ctx, limit, offset)
}
func (s serviceStub) GetByID(ctx context.Context, id int64) (ArticleResponse, error) {
    return s.getFn(ctx, id)
}
func (s serviceStub) Update(ctx context.Context, id int64, req UpdateArticleRequest) (ArticleResponse, error) {
    return s.updateFn(ctx, id, req)
}
func (s serviceStub) Patch(ctx context.Context, id int64, req PatchArticleRequest) (ArticleResponse, error) {
    return s.patchFn(ctx, id, req)
}
func (s serviceStub) Delete(ctx context.Context, id int64) error {
    return s.deleteFn(ctx, id)
}

func TestCreateHandler_Returns201(t *testing.T) {
    payload := `{"title":"` + strings.Repeat("T", 20) + `","content":"` + strings.Repeat("C", 200) + `","category":"Tech","status":"publish"}`
    svc := serviceStub{
        createFn: func(ctx context.Context, req CreateArticleRequest) (ArticleResponse, error) {
            return sampleResponse(), nil
        },
    }

    req := httptest.NewRequest(http.MethodPost, "/article", bytes.NewBufferString(payload))
    rec := httptest.NewRecorder()

    NewHandler(svc).Create(rec, req)

    assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestGetHandler_Returns404WhenMissing(t *testing.T) {
    svc := serviceStub{
        getFn: func(ctx context.Context, id int64) (ArticleResponse, error) {
            return ArticleResponse{}, ErrNotFound
        },
    }

    req := httptest.NewRequest(http.MethodGet, "/article/1", nil)
    req = withURLParam(req, "id", "1")
    rec := httptest.NewRecorder()

    NewHandler(svc).GetByID(rec, req)

    assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestListHandler_Returns200(t *testing.T) {
    svc := serviceStub{
        listFn: func(ctx context.Context, limit, offset int) ([]ArticleResponse, error) {
            return []ArticleResponse{sampleResponse()}, nil
        },
    }

    req := httptest.NewRequest(http.MethodGet, "/article/10/0", nil)
    req = withURLParam(req, "limit", "10")
    req = withURLParam(req, "offset", "0")
    rec := httptest.NewRecorder()

    NewHandler(svc).List(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDeleteHandler_Returns200(t *testing.T) {
    svc := serviceStub{
        deleteFn: func(ctx context.Context, id int64) error {
            return nil
        },
    }

    req := httptest.NewRequest(http.MethodDelete, "/article/2", nil)
    req = withURLParam(req, "id", "2")
    rec := httptest.NewRecorder()

    NewHandler(svc).Delete(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPatchHandler_Returns400ForInvalidPayload(t *testing.T) {
    payload := `{"title":"short"}`
    svc := serviceStub{}

    req := httptest.NewRequest(http.MethodPatch, "/article/1", bytes.NewBufferString(payload))
    req = withURLParam(req, "id", "1")
    rec := httptest.NewRecorder()

    NewHandler(svc).Patch(rec, req)

    assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPutHandler_Returns500OnServiceError(t *testing.T) {
    payload := `{"title":"` + strings.Repeat("T", 20) + `","content":"` + strings.Repeat("C", 200) + `","category":"Tech","status":"publish"}`
    svc := serviceStub{
        updateFn: func(ctx context.Context, id int64, req UpdateArticleRequest) (ArticleResponse, error) {
            return ArticleResponse{}, errors.New("db down")
        },
    }

    req := httptest.NewRequest(http.MethodPut, "/article/1", bytes.NewBufferString(payload))
    req = withURLParam(req, "id", "1")
    rec := httptest.NewRecorder()

    NewHandler(svc).Update(rec, req)

    assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func sampleResponse() ArticleResponse {
    return ArticleResponse{
        ID:          1,
        Title:       strings.Repeat("T", 20),
        Content:     strings.Repeat("C", 200),
        Category:    "Tech",
        Status:      StatusPublish,
        CreatedDate: time.Date(2026, 7, 4, 1, 2, 3, 0, time.UTC).Format(time.RFC3339),
        UpdatedDate: time.Date(2026, 7, 4, 4, 5, 6, 0, time.UTC).Format(time.RFC3339),
    }
}

func withURLParam(r *http.Request, key, value string) *http.Request {
    routeCtx := chi.NewRouteContext()
    routeCtx.URLParams.Add(key, value)
    return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
}

func TestCreateHandler_ResponseShape(t *testing.T) {
    payload := `{"title":"` + strings.Repeat("T", 20) + `","content":"` + strings.Repeat("C", 200) + `","category":"Tech","status":"publish"}`
    svc := serviceStub{
        createFn: func(ctx context.Context, req CreateArticleRequest) (ArticleResponse, error) {
            return sampleResponse(), nil
        },
    }

    req := httptest.NewRequest(http.MethodPost, "/article", bytes.NewBufferString(payload))
    rec := httptest.NewRecorder()

    NewHandler(svc).Create(rec, req)

    var got ArticleResponse
    err := json.Unmarshal(rec.Body.Bytes(), &got)

    assert.NoError(t, err)
    assert.Equal(t, int64(1), got.ID)
    assert.Equal(t, StatusPublish, got.Status)
}
```

- [ ] **Step 2: Run handler test to verify failure**

Run: `go test ./internal/article -run TestCreateHandler_Returns201 -v`
Expected: FAIL because handler types do not exist yet

- [ ] **Step 3: Implement handler and route registration**

Create `internal/article/handler.go`:

```go
package article

import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"

    "backend-article/internal/httpx"

    "github.com/go-chi/chi/v5"
)

type ServiceContract interface {
    Create(http.Context, CreateArticleRequest) (ArticleResponse, error)
    List(http.Context, int, int) ([]ArticleResponse, error)
    GetByID(http.Context, int64) (ArticleResponse, error)
    Update(http.Context, int64, UpdateArticleRequest) (ArticleResponse, error)
    Patch(http.Context, int64, PatchArticleRequest) (ArticleResponse, error)
    Delete(http.Context, int64) error
}

type Handler struct {
    service Service
}

func NewHandler(service Service) Handler {
    return Handler{service: service}
}

func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateArticleRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        httpx.WriteError(w, http.StatusBadRequest, "invalid JSON", nil)
        return
    }

    if errs := req.Validate(); len(errs) > 0 {
        httpx.WriteError(w, http.StatusBadRequest, "validation failed", errs)
        return
    }

    resp, err := h.service.Create(r.Context(), req)
    if err != nil {
        httpx.WriteError(w, http.StatusInternalServerError, "internal server error", nil)
        return
    }

    httpx.WriteJSON(w, http.StatusCreated, resp)
}

func (h Handler) List(w http.ResponseWriter, r *http.Request) {
    limit, err := strconv.Atoi(chi.URLParam(r, "limit"))
    if err != nil {
        httpx.WriteError(w, http.StatusBadRequest, "invalid limit", nil)
        return
    }

    offset, err := strconv.Atoi(chi.URLParam(r, "offset"))
    if err != nil {
        httpx.WriteError(w, http.StatusBadRequest, "invalid offset", nil)
        return
    }

    resp, err := h.service.List(r.Context(), limit, offset)
    if err != nil {
        httpx.WriteError(w, http.StatusBadRequest, err.Error(), nil)
        return
    }

    httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h Handler) GetByID(w http.ResponseWriter, r *http.Request) {
    id, ok := parseID(w, r)
    if !ok {
        return
    }

    resp, err := h.service.GetByID(r.Context(), id)
    if err != nil {
        handleServiceError(w, err)
        return
    }

    httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h Handler) Update(w http.ResponseWriter, r *http.Request) {
    id, ok := parseID(w, r)
    if !ok {
        return
    }

    var req UpdateArticleRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        httpx.WriteError(w, http.StatusBadRequest, "invalid JSON", nil)
        return
    }

    if errs := CreateArticleRequest(req).Validate(); len(errs) > 0 {
        httpx.WriteError(w, http.StatusBadRequest, "validation failed", errs)
        return
    }

    resp, err := h.service.Update(r.Context(), id, req)
    if err != nil {
        handleServiceError(w, err)
        return
    }

    httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h Handler) Patch(w http.ResponseWriter, r *http.Request) {
    id, ok := parseID(w, r)
    if !ok {
        return
    }

    var req PatchArticleRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        httpx.WriteError(w, http.StatusBadRequest, "invalid JSON", nil)
        return
    }

    if errs := req.Validate(); len(errs) > 0 {
        httpx.WriteError(w, http.StatusBadRequest, "validation failed", errs)
        return
    }

    resp, err := h.service.Patch(r.Context(), id, req)
    if err != nil {
        handleServiceError(w, err)
        return
    }

    httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h Handler) Delete(w http.ResponseWriter, r *http.Request) {
    id, ok := parseID(w, r)
    if !ok {
        return
    }

    err := h.service.Delete(r.Context(), id)
    if err != nil {
        handleServiceError(w, err)
        return
    }

    httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "article deleted"})
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
    id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
    if err != nil {
        httpx.WriteError(w, http.StatusBadRequest, "invalid id", nil)
        return 0, false
    }

    return id, true
}

func handleServiceError(w http.ResponseWriter, err error) {
    switch {
    case errors.Is(err, ErrNotFound):
        httpx.WriteError(w, http.StatusNotFound, err.Error(), nil)
    default:
        httpx.WriteError(w, http.StatusInternalServerError, "internal server error", nil)
    }
}
```

Create `internal/article/routes.go`:

```go
package article

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, handler Handler) {
    r.Route("/article", func(r chi.Router) {
        r.Post("/", handler.Create)
        r.Get("/{limit}/{offset}", handler.List)
        r.Get("/{id}", handler.GetByID)
        r.Put("/{id}", handler.Update)
        r.Patch("/{id}", handler.Patch)
        r.Delete("/{id}", handler.Delete)
    })
}
```

- [ ] **Step 4: Fix handler contract to use `context.Context` instead of invalid `http.Context`**

Update `internal/article/handler.go` service contract and constructor target:

```go
type ServiceContract interface {
    Create(context.Context, CreateArticleRequest) (ArticleResponse, error)
    List(context.Context, int, int) ([]ArticleResponse, error)
    GetByID(context.Context, int64) (ArticleResponse, error)
    Update(context.Context, int64, UpdateArticleRequest) (ArticleResponse, error)
    Patch(context.Context, int64, PatchArticleRequest) (ArticleResponse, error)
    Delete(context.Context, int64) error
}

type Handler struct {
    service ServiceContract
}

func NewHandler(service ServiceContract) Handler {
    return Handler{service: service}
}
```

Also add import:

```go
import "context"
```

- [ ] **Step 5: Run handler tests to verify they pass**

Run: `go test ./internal/article -run Handler -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/article/handler.go internal/article/routes.go internal/article/handler_test.go
git commit -m "feat: add article handlers"
```

## Task 9: Wire main application server

**Files:**
- Create: `cmd/api/main.go`
- Test: `go test ./...`

- [ ] **Step 1: Create application entrypoint**

Create `cmd/api/main.go`:

```go
package main

import (
    "log"
    "net/http"

    "backend-article/internal/article"
    "backend-article/internal/config"
    "backend-article/internal/database"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    db, err := database.NewMySQL(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    repo := article.NewRepository(db)
    service := article.NewService(repo)
    handler := article.NewHandler(service)

    router := chi.NewRouter()
    router.Use(middleware.RequestID)
    router.Use(middleware.RealIP)
    router.Use(middleware.Logger)
    router.Use(middleware.Recoverer)

    article.RegisterRoutes(router, handler)

    server := &http.Server{
        Addr:    ":" + cfg.AppPort,
        Handler: router,
    }

    log.Printf("server running on :%s", cfg.AppPort)
    log.Fatal(server.ListenAndServe())
}
```

- [ ] **Step 2: Run all tests and compile entrypoint**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 3: Smoke test startup without DB by expecting config error when env missing**

Run: `go run ./cmd/api`
Expected: exits quickly with `DB_NAME and DB_USER are required`

- [ ] **Step 4: Commit**

```bash
git add cmd/api/main.go
git commit -m "feat: wire article api server"
```

## Task 10: Add Postman collection

**Files:**
- Create: `postman/article-api.postman_collection.json`
- Test: `postman/article-api.postman_collection.json`

- [ ] **Step 1: Create collection JSON**

Create `postman/article-api.postman_collection.json`:

```json
{
  "info": {
    "name": "Article API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    { "key": "base_url", "value": "http://localhost:8080" },
    { "key": "article_id", "value": "1" },
    { "key": "limit", "value": "10" },
    { "key": "offset", "value": "0" }
  ],
  "item": [
    {
      "name": "Create Article",
      "request": {
        "method": "POST",
        "header": [{ "key": "Content-Type", "value": "application/json" }],
        "url": "{{base_url}}/article",
        "body": {
          "mode": "raw",
          "raw": "{\n  \"title\": \"This is an article title with twenty chars\",\n  \"content\": \"This article content is intentionally long so it exceeds two hundred characters. This article content is intentionally long so it exceeds two hundred characters. This article content is intentionally long so it exceeds two hundred characters.\",\n  \"category\": \"Technology\",\n  \"status\": \"publish\"\n}"
        }
      }
    },
    {
      "name": "List Articles",
      "request": {
        "method": "GET",
        "url": "{{base_url}}/article/{{limit}}/{{offset}}"
      }
    },
    {
      "name": "Get Article Detail",
      "request": {
        "method": "GET",
        "url": "{{base_url}}/article/{{article_id}}"
      }
    },
    {
      "name": "Put Article",
      "request": {
        "method": "PUT",
        "header": [{ "key": "Content-Type", "value": "application/json" }],
        "url": "{{base_url}}/article/{{article_id}}",
        "body": {
          "mode": "raw",
          "raw": "{\n  \"title\": \"This is updated article title text\",\n  \"content\": \"This updated article content is intentionally long so it exceeds two hundred characters. This updated article content is intentionally long so it exceeds two hundred characters. This updated article content is intentionally long so it exceeds two hundred characters.\",\n  \"category\": \"Business\",\n  \"status\": \"draft\"\n}"
        }
      }
    },
    {
      "name": "Patch Article",
      "request": {
        "method": "PATCH",
        "header": [{ "key": "Content-Type", "value": "application/json" }],
        "url": "{{base_url}}/article/{{article_id}}",
        "body": {
          "mode": "raw",
          "raw": "{\n  \"status\": \"publish\"\n}"
        }
      }
    },
    {
      "name": "Delete Article",
      "request": {
        "method": "DELETE",
        "url": "{{base_url}}/article/{{article_id}}"
      }
    }
  ]
}
```

- [ ] **Step 2: Validate JSON formatting**

Run: `python -m json.tool postman/article-api.postman_collection.json >/dev/null`
Expected: no output, exit code 0

- [ ] **Step 3: Commit**

```bash
git add postman/article-api.postman_collection.json
git commit -m "feat: add postman collection"
```

## Task 11: Final verification

**Files:**
- Verify: entire repository

- [ ] **Step 1: Run full test suite**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 2: Run formatter**

Run: `gofmt -w cmd internal`
Expected: files reformatted with no errors

- [ ] **Step 3: Re-run full test suite after formatting**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 4: Verify migration file with Goose**

Run: `go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations validate`
Expected: PASS

- [ ] **Step 5: Verify Postman collection JSON**

Run: `python -m json.tool postman/article-api.postman_collection.json >/dev/null`
Expected: exit code 0

- [ ] **Step 6: Build API binary**

Run: `go build ./cmd/api`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add .
git commit -m "test: verify article api build"
```

## Spec Coverage Check
- REST endpoints covered by Tasks 6, 7, 8, and 9
- MySQL + Goose migration covered by Tasks 2 and 3
- Validation rules covered by Task 4 and exercised in Task 8
- Soft delete behavior covered by Tasks 6 and 7
- Response shape with timestamps covered by Tasks 4, 6, and 8
- Postman collection covered by Task 10
- Final verification covered by Task 11

No spec gaps found.
