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
