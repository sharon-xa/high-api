-- +goose Up
ALTER TABLE users
ADD COLUMN IF NOT EXISTS birthdate DATE;

-- +goose Down
ALTER TABLE users
DROP COLUMN IF EXISTS birthdate;
