-- +goose Up
-- +goose StatementBegin
TRUNCATE repositories;
INSERT INTO repositories (uuid, namespace, name, visibility)
VALUES ('51151395-897d-4b9b-822b-a0c8796178cf',
        '4a9e18f4-64aa-4972-a586-581f31504594',
        'public-image',
        'public'),
       ('0c4e0de5-14b8-4719-9ccd-4bc73f742c65',
        '4a9e18f4-64aa-4972-a586-581f31504594',
        'private-image',
        'private'),
       ('db5bda5a-e7b9-4bde-91b9-fe3bcd687f63',
        '4970ce16-ceb2-4436-920b-e335fde10abe',
        'public-image',
        'public'),
       ('c89f0494-f9f8-419c-bc58-07949afcc7f1',
        '4970ce16-ceb2-4436-920b-e335fde10abe',
        'private-image',
        'private');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
TRUNCATE repositories;
-- +goose StatementEnd
