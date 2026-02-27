-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE urls ADD COLUMN active BOOLEAN NOT NULL DEFAULT TRUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE urls DROP COLUMN active;
-- +goose StatementEnd
