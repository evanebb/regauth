-- +goose Up
-- +goose StatementBegin
INSERT INTO repositories (id, namespace_id, name, visibility, created_at)
SELECT '0195cd13-ba14-76fd-b43e-55f190e566bd', id, 'public-image', 'public', '2025-01-01 00:00:00+00'
FROM namespaces
WHERE name = 'adminuser';

INSERT INTO repositories (id, namespace_id, name, visibility, created_at)
SELECT '0195cd13-ba14-7728-9e48-d51b8578ea53', id, 'private-image', 'private', '2025-01-01 00:00:00+00'
FROM namespaces
WHERE name = 'adminuser';

INSERT INTO repositories (id, namespace_id, name, visibility, created_at)
SELECT '0195cd13-ba14-779d-8960-bc20595f515e', id, 'public-image', 'public', '2025-01-01 00:00:00+00'
FROM namespaces
WHERE name = 'normaluser';

INSERT INTO repositories (id, namespace_id, name, visibility, created_at)
SELECT '0195cd13-ba14-77a5-bec6-46a26a17ad2d', id, 'private-image', 'private', '2025-01-01 00:00:00+00'
FROM namespaces
WHERE name = 'normaluser';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
