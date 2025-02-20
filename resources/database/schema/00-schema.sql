CREATE TYPE user_role AS ENUM ('admin', 'user');

DROP TABLE IF EXISTS users;
CREATE TABLE users
(
    id        serial PRIMARY KEY,
    uuid      uuid UNIQUE,
    username  varchar(255) UNIQUE,
    firstname varchar(255),
    lastname  varchar(255),
    role      user_role
);

INSERT INTO users (uuid, username, firstname, lastname, role)
VALUES ('965389fb-27ce-4f81-9e59-6ef9cb3b2472', 'admin', 'Initial', 'Admin', 'admin');

DROP TABLE IF EXISTS local_auth_users;
CREATE TABLE local_auth_users
(
    id            serial PRIMARY KEY,
    uuid          uuid REFERENCES users (uuid) ON DELETE CASCADE,
    username      varchar(255) REFERENCES users (username),
    password_hash varchar(255)
);
INSERT INTO local_auth_users (uuid, username, password_hash)
VALUES ('965389fb-27ce-4f81-9e59-6ef9cb3b2472', 'admin',
        '$2y$12$9tWON20iFLgRxqfr4YkaYObrSHlSbgDLlWjXnu0tzom4EvPo35vmS');

CREATE TYPE repository_visibility AS ENUM ('public', 'private');

DROP TABLE IF EXISTS repositories;
CREATE TABLE repositories
(
    id         serial PRIMARY KEY,
    uuid       uuid UNIQUE,
    namespace  varchar(255),
    name       varchar(255),
    visibility repository_visibility,
    owner_uuid uuid REFERENCES users (uuid) ON DELETE CASCADE,
    UNIQUE (namespace, name)
);

CREATE TYPE token_permission_type AS ENUM ('read_only', 'read_write', 'read_write_delete');

DROP TABLE IF EXISTS personal_access_tokens;
CREATE TABLE personal_access_tokens
(
    id              serial PRIMARY KEY,
    uuid            uuid UNIQUE,
    hash            varchar,
    description     varchar(255),
    permission_type token_permission_type,
    expiration_date timestamp,
    user_uuid       uuid REFERENCES users (uuid) ON DELETE CASCADE
);

DROP TABLE IF EXISTS personal_access_tokens_usage_log;
CREATE TABLE personal_access_tokens_usage_log
(
    id         serial PRIMARY KEY,
    token_uuid uuid REFERENCES personal_access_tokens (uuid) ON DELETE CASCADE,
    source_ip  inet,
    timestamp  timestamp
);
CREATE INDEX ON personal_access_tokens_usage_log (token_uuid);

DROP TABLE IF EXISTS sessions;
CREATE TABLE sessions
(
    id         serial PRIMARY KEY,
    session_id varchar(255) UNIQUE,
    data       bytea
);
