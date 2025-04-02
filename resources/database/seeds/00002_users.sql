-- +goose Up
-- +goose StatementBegin
INSERT INTO users (id, username, role, password_hash, created_at)
VALUES ('0195cd11-2863-71d4-a3c4-032bc264cf81',
        'adminuser',
        'admin',
           -- Welkom01!
        '$2y$12$sSMlPGBCt2RZnX5Od405T./kEwKZYtJoIhijrL1XXlwvr/BtPDtgS',
        '2025-01-01 00:00:00+00'),
       ('0195cd11-2863-721e-a75c-86522539d0ee',
        'normaluser',
        'user',
           -- Welkom02!
        '$2y$12$rSVIBbxHKnfQnzeaJFfoouZuBeiyjSrxPyZIz/6L2CnrwScTYWJiq',
        '2025-01-01 00:00:00+00');

INSERT INTO namespaces (name, user_id)
VALUES ('adminuser', '0195cd11-2863-71d4-a3c4-032bc264cf81'),
       ('normaluser', '0195cd11-2863-721e-a75c-86522539d0ee');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
