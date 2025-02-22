-- +goose Up
-- +goose StatementBegin
TRUNCATE personal_access_tokens CASCADE;
INSERT INTO personal_access_tokens (uuid, hash, last_eight, description, permission, expiration_date, user_uuid)
VALUES ('4a73fcfe-4eb2-476d-8ce7-42989e7a71c7',
           -- registry_pat_SVV_otfQNmSjo7viDiCrC0AKe6Qa_iFhxXJBZE1vMOByC9nbUtBPsz3r
        '$2a$12$V4qP3wktXIvmMlMg9rOzvek.TRqmgYNQ0W7KSIA9uFDMu68XKXQe2',
        'UtBPsz3r',
        'Read-only token',
        'read_only',
        '2044-12-31 12:06:30',
        '4f42bc8c-4c92-4fe7-b2d6-31a425128870'),
       ('2ecbbeb0-1d40-48c3-8647-d846d84ad93c',
           -- registry_pat_mKOJIyGsu6SXXcleHw1gm7q32H7zFRdpO57Nena0X2YOYNrJB04GMpjm
        '$2a$12$lQFbWKNvWXLDcsjUt3y2EegAidbflhnFxqWRlzNQykPzjd5bjOX4W',
        'B04GMpjm',
        'Read-write token',
        'read_write',
        '2044-12-31 12:06:30',
        '4f42bc8c-4c92-4fe7-b2d6-31a425128870'),
       ('5247eb99-4c4b-4669-b584-860327456831',
           -- registry_pat_tnp2rY09d3I0k9UH94f5CE4N3-LtdReVKS6ve07C7wJ3W09e26EXKtC2
        '$2a$12$H2xnJPEbnypGM.r6OlM/8eS6WAAE44taq.q6kBC4R9PjtxzKYQBba',
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
