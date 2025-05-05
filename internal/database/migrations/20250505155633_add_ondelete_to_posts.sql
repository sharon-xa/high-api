-- +goose Up
-- Drop the existing foreign key constraint (replace with actual constraint name if different)
ALTER TABLE posts
DROP CONSTRAINT IF EXISTS fk_users_posts;

-- Add the new foreign key constraint with ON DELETE CASCADE
ALTER TABLE posts
ADD CONSTRAINT fk_users_posts
FOREIGN KEY (user_id) REFERENCES users(id)
ON DELETE CASCADE;


-- +goose Down
-- Drop the ON DELETE CASCADE constraint
ALTER TABLE posts
DROP CONSTRAINT IF EXISTS fk_users_posts;

-- Optionally, re-add the original constraint without ON DELETE CASCADE (NO ACTION is default)
ALTER TABLE posts
ADD CONSTRAINT fk_users_posts
FOREIGN KEY (user_id) REFERENCES users(id)
ON DELETE NO ACTION;
