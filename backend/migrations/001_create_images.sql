CREATE TABLE IF NOT EXISTS images (
	id UUID PRIMARY KEY,
	user_id TEXT NOT NULL,
	object_name TEXT NOT NULL,
	original_filename TEXT NOT NULL,
	content_type TEXT NOT NULL,
	size_bytes BIGINT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (user_id, object_name)
);

CREATE INDEX IF NOT EXISTS idx_images_user_id_created_at
	ON images (user_id, created_at DESC);
