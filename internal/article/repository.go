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
		args = append(args, strings.TrimSpace(*req.Title))
	}
	if req.Content != nil {
		sets = append(sets, "content = ?")
		args = append(args, strings.TrimSpace(*req.Content))
	}
	if req.Category != nil {
		sets = append(sets, "category = ?")
		args = append(args, strings.TrimSpace(*req.Category))
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
