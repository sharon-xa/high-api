-- +goose Up
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS deleted_at;

-- +goose Down
ALTER TABLE refresh_tokens ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;
