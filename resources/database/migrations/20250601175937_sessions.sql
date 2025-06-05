-- +goose Up
-- +goose StatementBegin
CREATE TABLE sessions
(
    id         serial PRIMARY KEY,
    session_id varchar(255) UNIQUE,
    data       bytea
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE sessions;
-- +goose StatementEnd
