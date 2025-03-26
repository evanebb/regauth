-- +goose Up
-- +goose StatementBegin
INSERT INTO personal_access_tokens (id, hash, last_eight, description, permission, expiration_date, user_id)
VALUES ('0195cd16-2142-78e5-8425-a8db7acbc8f8',
           -- registry_pat_SVV_otfQNmSjo7viDiCrC0AKe6Qa_iFhxXJBZE1vMOByC9nbUtBPsz3r
        'f18c6dd73fb5831deae4dc5dfd8b6d347dd32b173e75e3f2b224fc8437c4422ec0881e000d282faf546ba05d6b449ba1',
        'UtBPsz3r',
        'Read-only token',
        'read_only',
        '2045-03-25 12:16:33.110405+00',
        '0195cd11-2863-71d4-a3c4-032bc264cf81'),
       ('0195cd16-2142-790c-8835-6d3430abb642',
           -- registry_pat_mKOJIyGsu6SXXcleHw1gm7q32H7zFRdpO57Nena0X2YOYNrJB04GMpjm
        'f0d8badab0e7f5c111f4a7fe8f4bf7147b4be33f37fcb3c7826705822525e3cdb5cbd4ace63395156fa3afc70385fe52',
        'B04GMpjm',
        'Read-write token',
        'read_write',
        '2045-03-25 12:16:33.110405+00',
        '0195cd11-2863-71d4-a3c4-032bc264cf81'),
       ('0195cd16-2142-7914-86f9-58a374eda416',
           -- registry_pat_tnp2rY09d3I0k9UH94f5CE4N3-LtdReVKS6ve07C7wJ3W09e26EXKtC2
        'fcb8f45b2e935f401f1b97fde4503d8377e1322d32ad77686b8b0e4b13fd330ed288975f9d007e8fc7d6f043e3302c0b',
        '26EXKtC2',
        'Read-write-delete token',
        'read_write_delete',
        '2045-03-25 12:16:33.110405+00',
        '0195cd11-2863-71d4-a3c4-032bc264cf81');

INSERT INTO personal_access_tokens_usage_log (token_id, source_ip, timestamp)
VALUES ('0195cd16-2142-78e5-8425-a8db7acbc8f8', '192.168.1.10', '2045-03-26 12:16:33.110405+00'),
       ('0195cd16-2142-78e5-8425-a8db7acbc8f8', '192.168.1.11', '2045-03-27 12:16:33.110405+00'),
       ('0195cd16-2142-78e5-8425-a8db7acbc8f8', '192.168.1.12', '2045-03-28 12:16:33.110405+00');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
