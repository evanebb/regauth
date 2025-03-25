-- +goose Up
-- +goose StatementBegin
TRUNCATE repositories;
INSERT INTO repositories (id, namespace_id, name, visibility)
VALUES ('0195cd13-ba14-76fd-b43e-55f190e566bd',
        '0195cd11-2863-7226-9664-054d3cb1c752',
        'public-image',
        'public'),
       ('0195cd13-ba14-7728-9e48-d51b8578ea53',
        '0195cd11-2863-7226-9664-054d3cb1c752',
        'private-image',
        'private'),
       ('0195cd13-ba14-779d-8960-bc20595f515e',
        '0195cd11-2863-722a-b4d9-e7987725477b',
        'public-image',
        'public'),
       ('0195cd13-ba14-77a5-bec6-46a26a17ad2d',
        '0195cd11-2863-722a-b4d9-e7987725477b',
        'private-image',
        'private');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
TRUNCATE repositories;
-- +goose StatementEnd
