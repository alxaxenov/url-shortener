-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE users (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO users (created_at) VALUES (NOW()); -- Технический пользователь

ALTER TABLE urls ADD COLUMN user_id INTEGER;
UPDATE urls SET user_id = 1;
ALTER TABLE urls ALTER COLUMN user_id SET NOT NULL;
CREATE INDEX idx_urls_user_id ON urls(user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP INDEX IF EXISTS idx_urls_user_id;
ALTER TABLE urls DROP COLUMN IF EXISTS user_id;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
