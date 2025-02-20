INSERT INTO
    personal_access_tokens (uuid, hash, description, permission_type, expiration_date, user_uuid)
VALUES
    ('4a73fcfe-4eb2-476d-8ce7-42989e7a71c7', '$2y$12$nAbHQ2pWUKgajr.GSoyeVuxHqgTJBHehWE4gNcntBjaUmCW4aAH76', 'Read-only token', 'read_only', '2044-12-31 12:06:30', '965389fb-27ce-4f81-9e59-6ef9cb3b2472'),
    ('2ecbbeb0-1d40-48c3-8647-d846d84ad93c', '$2y$12$31BLUeLKAPzZbGJDsKJ2EOspDJXgmvkdxIzdJByLRYCSInMLW0bLe', 'Read-write token', 'read_write', '2044-12-31 12:06:30', '965389fb-27ce-4f81-9e59-6ef9cb3b2472'),
    ('5247eb99-4c4b-4669-b584-860327456831', '$2y$12$zrJZWVuvnHMSaqMlZsBk6.JSogOBNtlekpOlMFVE5ECU34J3902JW', 'Read-write-delete token', 'read_write_delete', '2044-12-31 12:06:30', '965389fb-27ce-4f81-9e59-6ef9cb3b2472');

INSERT INTO
    repositories (uuid, namespace, name, visibility, owner_uuid)
VALUES
    ('51151395-897d-4b9b-822b-a0c8796178cf', 'admin', 'public-image', 'public', '965389fb-27ce-4f81-9e59-6ef9cb3b2472'),
    ('0c4e0de5-14b8-4719-9ccd-4bc73f742c65', 'admin', 'private-image', 'private', '965389fb-27ce-4f81-9e59-6ef9cb3b2472')
