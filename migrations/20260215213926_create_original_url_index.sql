-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE UNIQUE INDEX original_url_unique_idx ON urls (original_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP INDEX original_url_unique_idx;
-- +goose StatementEnd
