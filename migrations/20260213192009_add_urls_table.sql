-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,
    short_url VARCHAR(10) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
