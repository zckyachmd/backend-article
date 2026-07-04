package article

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

type CreateArticleRequest struct {
	Title    string `json:"title" validate:"required,min=20,max=200"`
	Content  string `json:"content" validate:"required,min=200"`
	Category string `json:"category" validate:"required,min=3,max=100"`
	Status   string `json:"status" validate:"required,oneof=publish draft thrash"`
}

type UpdateArticleRequest = CreateArticleRequest

type PatchArticleRequest struct {
	Title    *string `json:"title" validate:"omitempty,min=20,max=200"`
	Content  *string `json:"content" validate:"omitempty,min=200"`
	Category *string `json:"category" validate:"omitempty,min=3,max=100"`
	Status   *string `json:"status" validate:"omitempty,oneof=publish draft thrash"`
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

var validate = validator.New()

func (r CreateArticleRequest) Validate() map[string][]string {
	r.Title = strings.TrimSpace(r.Title)
	r.Content = strings.TrimSpace(r.Content)
	r.Category = strings.TrimSpace(r.Category)
	r.Status = strings.ToLower(strings.TrimSpace(r.Status))

	return validationErrors(validate.Struct(r))
}

func (r PatchArticleRequest) Validate() map[string][]string {
	if r.Title == nil && r.Content == nil && r.Category == nil && r.Status == nil {
		return map[string][]string{"body": {"at least one field is required"}}
	}

	if r.Title != nil {
		trimmed := strings.TrimSpace(*r.Title)
		r.Title = &trimmed
	}
	if r.Content != nil {
		trimmed := strings.TrimSpace(*r.Content)
		r.Content = &trimmed
	}
	if r.Category != nil {
		trimmed := strings.TrimSpace(*r.Category)
		r.Category = &trimmed
	}
	if r.Status != nil {
		normalized := strings.ToLower(strings.TrimSpace(*r.Status))
		r.Status = &normalized
	}

	return validationErrors(validate.Struct(r))
}

func validationErrors(err error) map[string][]string {
	if err == nil {
		return map[string][]string{}
	}

	verrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return map[string][]string{"body": {err.Error()}}
	}

	result := map[string][]string{}
	for _, verr := range verrs {
		field := strings.ToLower(verr.Field())
		switch verr.Tag() {
		case "required":
			result[field] = append(result[field], "is required")
		case "min":
			result[field] = append(result[field], "must be at least "+verr.Param()+" characters")
		case "max":
			result[field] = append(result[field], "must be at most "+verr.Param()+" characters")
		case "oneof":
			result[field] = append(result[field], "must be one of publish, draft, thrash")
		}
	}

	return result
}
