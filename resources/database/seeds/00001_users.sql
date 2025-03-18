-- +goose Up
-- +goose StatementBegin
TRUNCATE users CASCADE;
INSERT INTO users (uuid, username, role)
VALUES ('4f42bc8c-4c92-4fe7-b2d6-31a425128870',
        'adminuser',
        'admin'),
       ('1a74e377-33f3-416f-88da-98fc16add21e',
        'normaluser',
        'user');

INSERT INTO local_auth_users (uuid, username, password_hash)
VALUES ('4f42bc8c-4c92-4fe7-b2d6-31a425128870',
        'adminuser',
           -- Welkom01!
        '$2y$12$sSMlPGBCt2RZnX5Od405T./kEwKZYtJoIhijrL1XXlwvr/BtPDtgS'),
       ('1a74e377-33f3-416f-88da-98fc16add21e',
        'normaluser',
           -- Welkom02!
        '$2y$12$rSVIBbxHKnfQnzeaJFfoouZuBeiyjSrxPyZIz/6L2CnrwScTYWJiq');

INSERT INTO namespaces (uuid, name, user_uuid)
VALUES ('4a9e18f4-64aa-4972-a586-581f31504594', 'adminuser', '4f42bc8c-4c92-4fe7-b2d6-31a425128870'),
       ('4970ce16-ceb2-4436-920b-e335fde10abe', 'normaluser', '1a74e377-33f3-416f-88da-98fc16add21e');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
TRUNCATE users CASCADE;
-- +goose StatementEnd
