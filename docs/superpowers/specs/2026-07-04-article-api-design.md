# Article API Design

## Goal
Build a simple Golang article API microservice with MySQL, Goose migrations, REST endpoints, request validation, soft delete via status, and Postman collection output. Avoid over-engineering.

## Scope
This service manages articles in a single `posts` table and exposes CRUD-style REST endpoints for create, list, detail, update, and soft delete.

## Decisions
- Language: Go
- HTTP stack: `net/http` with `github.com/go-chi/chi/v5`
- Database: MySQL
- Migration tool: `github.com/pressly/goose/v3`
- DB access: `database/sql` with `github.com/jmoiron/sqlx`
- MySQL driver: `github.com/go-sql-driver/mysql`
- Validation: `github.com/go-playground/validator/v10`
- Env loading: `github.com/joho/godotenv`
- Test helpers: `github.com/stretchr/testify`
- Delete behavior: soft delete by setting `status = "thrash"`
- Timestamp behavior: `created_date` and `updated_date` set automatically by database

## Architecture
Keep structure simple and common for Go APIs:

- `cmd/api/main.go`
  - load config
  - open database
  - register router
  - start HTTP server
- `internal/config`
  - parse environment configuration
- `internal/database`
  - build MySQL connection
- `internal/article`
  - `model.go` for persistent model
  - `dto.go` for request and response payloads
  - `repository.go` for SQL queries
  - `service.go` for light business rules
  - `handler.go` for HTTP handlers
- `internal/httpx`
  - JSON response and error helpers
- `migrations/`
  - Goose SQL migration files
- `postman/`
  - Postman collection JSON

This keeps layers shallow: handler → service → repository.

## Database Design
Table name: `posts`

Columns:
- `id` INT AUTO_INCREMENT PRIMARY KEY
- `title` VARCHAR(200) NOT NULL
- `content` TEXT NOT NULL
- `category` VARCHAR(100) NOT NULL
- `status` VARCHAR(100) NOT NULL
- `created_date` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
- `updated_date` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

Recommended indexes:
- primary key on `id`
- index on `status`
- index on `category`

Status allowed values at application level:
- `publish`
- `draft`
- `thrash`

Soft delete means `DELETE /article/{id}` updates status to `thrash` instead of removing row.

## API Design
Base resource path: `/article`

### 1. Create article
- Method: `POST`
- Path: `/article`
- Request body:
```json
{
  "title": "...",
  "content": "...",
  "category": "...",
  "status": "publish"
}
```
- Success response: `201 Created`
- Response body returns created resource including `id`, `created_date`, `updated_date`

### 2. List articles
- Method: `GET`
- Path: `/article/{limit}/{offset}`
- Success response: `200 OK`
- Response body: array of article objects
- Sorting recommendation: newest first by `id DESC`

### 3. Get article detail
- Method: `GET`
- Path: `/article/{id}`
- Success response: `200 OK`
- Response body: single article object

### 4. Replace article
- Method: `PUT`
- Path: `/article/{id}`
- Full payload required
- Success response: `200 OK`

### 5. Partially update article
- Method: `PATCH`
- Path: `/article/{id}`
- Partial payload allowed
- If a field is present, it must pass validation rules
- Success response: `200 OK`

### 6. Soft delete article
- Method: `DELETE`
- Path: `/article/{id}`
- Behavior: update `status` to `thrash`
- Success response: `200 OK`

## Validation Rules
### Create and PUT
- `title`: required, minimum 20 characters, maximum 200 characters
- `content`: required, minimum 200 characters
- `category`: required, minimum 3 characters, maximum 100 characters
- `status`: required, must be one of `publish`, `draft`, `thrash`

### PATCH
- all fields optional
- if `title` provided: minimum 20, maximum 200
- if `content` provided: minimum 200
- if `category` provided: minimum 3, maximum 100
- if `status` provided: one of `publish`, `draft`, `thrash`
- reject empty PATCH body

## Response Shape
Standard article response:
```json
{
  "id": 1,
  "title": "...",
  "content": "...",
  "category": "...",
  "status": "publish",
  "created_date": "2026-07-04T08:00:00Z",
  "updated_date": "2026-07-04T08:00:00Z"
}
```

Error response shape:
```json
{
  "message": "validation failed",
  "errors": {
    "title": ["must be at least 20 characters"]
  }
}
```

## Error Handling
- `400 Bad Request`
  - invalid JSON
  - validation failure
  - invalid path parameter
  - empty PATCH body
- `404 Not Found`
  - article ID not found
- `500 Internal Server Error`
  - database or unexpected server failure

## Query Behavior
- List endpoint reads `limit` and `offset` from path parameters
- Both must be integers
- `limit` should be constrained in handler or service to prevent abuse
- Recommended default ceiling: if supplied `limit` > 100, reject request with `400`

## Business Rules
- `DELETE /article/{id}` never removes row physically
- fetching detail still returns row even if status is `thrash`, unless later requirements say otherwise
- list endpoint returns all rows including `thrash`, matching current spec literally; filtering can be added later if needed

## Configuration
Environment variables:
- `APP_PORT`
- `APP_ENV`
- `DB_HOST`
- `DB_PORT`
- `DB_NAME`
- `DB_USER`
- `DB_PASSWORD`
- `DB_PARAMS` optional additional DSN params

Optional `.env` file for local development via `godotenv`.

## Migration Strategy
Use Goose SQL migrations in `migrations/`.

Initial migration responsibilities:
- create `posts` table
- define timestamp defaults
- create indexes for `status` and `category`

Recommended commands:
- `goose create create_posts_table sql`
- `goose up`
- `goose status`

## Postman Collection
Collection should include:
- create article
- list articles
- get article detail
- put article
- patch article
- delete article

Collection should use variables for:
- `base_url`
- `article_id`
- `limit`
- `offset`

## Testing Strategy
Keep tests practical and small.

Priorities:
1. validation tests for create and patch payloads
2. service tests for soft delete behavior and not found handling
3. handler tests for HTTP status and JSON response behavior

No need for heavy integration test setup in first pass.

## Non-Goals
- authentication and authorization
- search
- filtering
- sorting customization
- pagination metadata wrapper
- Swagger/OpenAPI generation
- Docker/Kubernetes setup
- message broker or async workflow

## Implementation Notes
- prefer explicit SQL in repository layer
- keep interfaces only where they help testing or decoupling; avoid interface-per-struct habit
- keep helper packages small; do not create generic utility sprawl
- use JSON field names exactly as API contract requires
- use UTC-compatible timestamp handling in responses
