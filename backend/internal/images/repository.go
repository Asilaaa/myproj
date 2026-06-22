package images

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"myproj/internal/database"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, image *Image) error {
	if image == nil {
		return errors.New("image is nil")
	}

	if image.ID == "" {
		image.ID = uuid.NewString()
	}

	query := `
		INSERT INTO images (
			id,
			user_id,
			object_name,
			original_filename,
			content_type,
			size_bytes,
			description
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at
	`

	return r.db.Pool.QueryRow(
		ctx,
		query,
		image.ID,
		image.UserID,
		image.ObjectName,
		image.OriginalFilename,
		image.ContentType,
		image.SizeBytes,
		image.Description,
	).Scan(&image.CreatedAt)
}

func (r *Repository) GetByIDForUser(ctx context.Context, id string, userID string) (*Image, error) {
	query := `
		SELECT id, user_id, object_name, original_filename, content_type, size_bytes, description, created_at
		FROM images
		WHERE id = $1 AND user_id = $2
	`

	var image Image
	err := r.db.Pool.QueryRow(ctx, query, id, userID).Scan(
		&image.ID,
		&image.UserID,
		&image.ObjectName,
		&image.OriginalFilename,
		&image.ContentType,
		&image.SizeBytes,
		&image.Description,
		&image.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &image, nil
}

func (r *Repository) GetByObjectNameForUser(ctx context.Context, userID string, objectName string) (*Image, error) {
	query := `
		SELECT id, user_id, object_name, original_filename, content_type, size_bytes, description, created_at
		FROM images
		WHERE user_id = $1 AND object_name = $2
	`

	var image Image
	err := r.db.Pool.QueryRow(ctx, query, userID, objectName).Scan(
		&image.ID,
		&image.UserID,
		&image.ObjectName,
		&image.OriginalFilename,
		&image.ContentType,
		&image.SizeBytes,
		&image.Description,
		&image.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &image, nil
}

func (r *Repository) ListByUserID(ctx context.Context, userID string) ([]Image, error) {
	query := `
		SELECT id, user_id, object_name, original_filename, content_type, size_bytes, description, created_at
		FROM images
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	images := make([]Image, 0)
	for rows.Next() {
		var image Image
		if err := rows.Scan(
			&image.ID,
			&image.UserID,
			&image.ObjectName,
			&image.OriginalFilename,
			&image.ContentType,
			&image.SizeBytes,
			&image.Description,
			&image.CreatedAt,
		); err != nil {
			return nil, err
		}

		images = append(images, image)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return images, nil
}

func (r *Repository) UpdateDescription(ctx context.Context, id string, userID string, description string) error {
	commandTag, err := r.db.Pool.Exec(ctx, `UPDATE images SET description = $1 WHERE id = $2 AND user_id = $3`, description, id, userID)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *Repository) DeleteByIDForUser(ctx context.Context, id string, userID string) error {
	commandTag, err := r.db.Pool.Exec(ctx, `DELETE FROM images WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
