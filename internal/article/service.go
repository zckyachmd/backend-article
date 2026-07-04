package article

import (
	"context"
	"errors"
	"time"
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
		CreatedDate: article.CreatedDate.UTC().Format(time.RFC3339),
		UpdatedDate: article.UpdatedDate.UTC().Format(time.RFC3339),
	}
}
