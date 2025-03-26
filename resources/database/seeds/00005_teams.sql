-- +goose Up
-- +goose StatementBegin
INSERT INTO teams (id, name)
VALUES ('0195d46e-cfbf-7324-b9aa-4c9c78d3b722', 'team-1'),
       ('0195d46f-fde4-7b27-b542-e41ed0917ace', 'team-2');

INSERT INTO namespaces (name, team_id)
VALUES ('team-1', '0195d46e-cfbf-7324-b9aa-4c9c78d3b722'),
       ('team-2', '0195d46f-fde4-7b27-b542-e41ed0917ace');

INSERT INTO team_members (user_id, team_id, role)
VALUES ('0195cd11-2863-71d4-a3c4-032bc264cf81', '0195d46e-cfbf-7324-b9aa-4c9c78d3b722', 'admin'),
       ('0195cd11-2863-721e-a75c-86522539d0ee', '0195d46e-cfbf-7324-b9aa-4c9c78d3b722', 'user'),
       ('0195cd11-2863-71d4-a3c4-032bc264cf81', '0195d46f-fde4-7b27-b542-e41ed0917ace', 'admin');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
