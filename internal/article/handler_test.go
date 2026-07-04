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
	createFn func(context.Context, CreateArticleRequest) (ArticleResponse, error)
	listFn   func(context.Context, int, int) ([]ArticleResponse, error)
	getFn    func(context.Context, int64) (ArticleResponse, error)
	updateFn func(context.Context, int64, UpdateArticleRequest) (ArticleResponse, error)
	patchFn  func(context.Context, int64, PatchArticleRequest) (ArticleResponse, error)
	deleteFn func(context.Context, int64) error
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
	routeCtx, ok := r.Context().Value(chi.RouteCtxKey).(*chi.Context)
	if !ok {
		routeCtx = chi.NewRouteContext()
	}

	routeCtx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
}
