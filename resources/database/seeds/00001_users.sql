-- +goose Up
-- +goose StatementBegin
TRUNCATE users CASCADE;
INSERT INTO users (id, username, role, password_hash)
VALUES ('0195cd11-2863-71d4-a3c4-032bc264cf81',
        'adminuser',
        'admin'
            '$2y$12$sSMlPGBCt2RZnX5Od405T./kEwKZYtJoIhijrL1XXlwvr/BtPDtgS'),
       ('0195cd11-2863-721e-a75c-86522539d0ee',
        'normaluser',
        'user'
            '$2y$12$rSVIBbxHKnfQnzeaJFfoouZuBeiyjSrxPyZIz/6L2CnrwScTYWJiq');

INSERT INTO namespaces (uuid, name, user_id)
VALUES ('0195cd11-2863-7226-9664-054d3cb1c752', 'adminuser', '0195cd11-2863-71d4-a3c4-032bc264cf81'),
       ('0195cd11-2863-722a-b4d9-e7987725477b', 'normaluser', '0195cd11-2863-721e-a75c-86522539d0ee');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
TRUNCATE users CASCADE;
-- +goose StatementEnd
