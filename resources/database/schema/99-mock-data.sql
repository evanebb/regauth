INSERT INTO personal_access_tokens (uuid, hash, last_eight, description, permission, expiration_date, user_uuid)
VALUES
    -- registry_pat_SVV_otfQNmSjo7viDiCrC0AKe6Qa_iFhxXJBZE1vMOByC9nbUtBPsz3r
    ('4a73fcfe-4eb2-476d-8ce7-42989e7a71c7', '$2a$12$V4qP3wktXIvmMlMg9rOzvek.TRqmgYNQ0W7KSIA9uFDMu68XKXQe2', 'UtBPsz3r',
     'Read-only token', 'read_only', '2044-12-31 12:06:30', '965389fb-27ce-4f81-9e59-6ef9cb3b2472'),
    -- registry_pat_mKOJIyGsu6SXXcleHw1gm7q32H7zFRdpO57Nena0X2YOYNrJB04GMpjm
    ('2ecbbeb0-1d40-48c3-8647-d846d84ad93c', '$2a$12$lQFbWKNvWXLDcsjUt3y2EegAidbflhnFxqWRlzNQykPzjd5bjOX4W', 'B04GMpjm',
     'Read-write token', 'read_write', '2044-12-31 12:06:30', '965389fb-27ce-4f81-9e59-6ef9cb3b2472'),
    -- registry_pat_tnp2rY09d3I0k9UH94f5CE4N3-LtdReVKS6ve07C7wJ3W09e26EXKtC2
    ('5247eb99-4c4b-4669-b584-860327456831', '$2a$12$H2xnJPEbnypGM.r6OlM/8eS6WAAE44taq.q6kBC4R9PjtxzKYQBba', '26EXKtC2',
     'Read-write-delete token', 'read_write_delete', '2044-12-31 12:06:30', '965389fb-27ce-4f81-9e59-6ef9cb3b2472');

INSERT INTO repositories (uuid, namespace, name, visibility, owner_uuid)
VALUES ('51151395-897d-4b9b-822b-a0c8796178cf', 'admin', 'public-image', 'public',
        '965389fb-27ce-4f81-9e59-6ef9cb3b2472'),
       ('0c4e0de5-14b8-4719-9ccd-4bc73f742c65', 'admin', 'private-image', 'private',
        '965389fb-27ce-4f81-9e59-6ef9cb3b2472')
