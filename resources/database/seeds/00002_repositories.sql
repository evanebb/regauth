-- +goose Up
-- +goose StatementBegin
TRUNCATE repositories;
INSERT INTO repositories (uuid, namespace, name, visibility, owner_uuid)
VALUES ('51151395-897d-4b9b-822b-a0c8796178cf',
        'adminuser',
        'public-image',
        'public',
        '4f42bc8c-4c92-4fe7-b2d6-31a425128870'),
       ('0c4e0de5-14b8-4719-9ccd-4bc73f742c65',
        'adminuser',
        'private-image',
        'private',
        '4f42bc8c-4c92-4fe7-b2d6-31a425128870'),
       ('db5bda5a-e7b9-4bde-91b9-fe3bcd687f63',
        'normaluser',
        'public-image',
        'public',
        '1a74e377-33f3-416f-88da-98fc16add21e'),
       ('c89f0494-f9f8-419c-bc58-07949afcc7f1',
        'normaluser',
        'private-image',
        'private',
        '1a74e377-33f3-416f-88da-98fc16add21e');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
TRUNCATE repositories;
-- +goose StatementEnd
