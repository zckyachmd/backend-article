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
