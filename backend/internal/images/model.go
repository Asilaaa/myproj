package images

import "time"

type Image struct {
	ID               string    `db:"id" json:"id"`
	UserID           string    `db:"user_id" json:"user_id"`
	ObjectName       string    `db:"object_name" json:"object_name"`
	OriginalFilename string    `db:"original_filename" json:"original_filename"`
	ContentType      string    `db:"content_type" json:"content_type"`
	SizeBytes        int64     `db:"size_bytes" json:"size_bytes"`
	Description      string    `db:"description" json:"description"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}
