package article

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"backend-article/internal/httpx"

	"github.com/go-chi/chi/v5"
)

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

	if err := h.service.Delete(r.Context(), id); err != nil {
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
