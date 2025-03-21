-- +goose Up
-- +goose StatementBegin
TRUNCATE personal_access_tokens CASCADE;
INSERT INTO personal_access_tokens (uuid, hash, last_eight, description, permission, expiration_date, user_uuid)
VALUES ('4a73fcfe-4eb2-476d-8ce7-42989e7a71c7',
           -- registry_pat_SVV_otfQNmSjo7viDiCrC0AKe6Qa_iFhxXJBZE1vMOByC9nbUtBPsz3r
        'f18c6dd73fb5831deae4dc5dfd8b6d347dd32b173e75e3f2b224fc8437c4422ec0881e000d282faf546ba05d6b449ba1',
        'UtBPsz3r',
        'Read-only token',
        'read_only',
        '2044-12-31 12:06:30',
        '4f42bc8c-4c92-4fe7-b2d6-31a425128870'),
       ('2ecbbeb0-1d40-48c3-8647-d846d84ad93c',
           -- registry_pat_mKOJIyGsu6SXXcleHw1gm7q32H7zFRdpO57Nena0X2YOYNrJB04GMpjm
        'f0d8badab0e7f5c111f4a7fe8f4bf7147b4be33f37fcb3c7826705822525e3cdb5cbd4ace63395156fa3afc70385fe52',
        'B04GMpjm',
        'Read-write token',
        'read_write',
        '2044-12-31 12:06:30',
        '4f42bc8c-4c92-4fe7-b2d6-31a425128870'),
       ('5247eb99-4c4b-4669-b584-860327456831',
           -- registry_pat_tnp2rY09d3I0k9UH94f5CE4N3-LtdReVKS6ve07C7wJ3W09e26EXKtC2
        'fcb8f45b2e935f401f1b97fde4503d8377e1322d32ad77686b8b0e4b13fd330ed288975f9d007e8fc7d6f043e3302c0b',
        '26EXKtC2',
        'Read-write-delete token',
        'read_write_delete',
        '2044-12-31 12:06:30',
        '4f42bc8c-4c92-4fe7-b2d6-31a425128870');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
TRUNCATE personal_access_tokens CASCADE;
-- +goose StatementEnd
