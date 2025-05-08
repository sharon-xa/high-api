-- +goose Up
ALTER TABLE refresh_tokens
DROP constraint IF EXISTS fk_users_refresh_tokens;

ALTER TABLE refresh_tokens
ADD CONSTRAINT fk_users_refresh_tokens
FOREIGN KEY (user_id) REFERENCES users(id)
ON DELETE CASCADE;

-- +goose Down
ALTER TABLE refresh_tokens
DROP CONSTRAINT IF EXISTS fk_users_refresh_tokens;

ALTER TABLE refresh_tokens
ADD CONSTRAINT fk_users_refresh_tokens
FOREIGN KEY (user_id) REFERENCES users(id)
ON DELETE NO ACTION;
