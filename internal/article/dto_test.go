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
