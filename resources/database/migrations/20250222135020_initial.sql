-- +goose Up
-- +goose StatementBegin
CREATE TYPE user_role AS ENUM ('admin', 'user');

CREATE TABLE users
(
    id       bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    uuid     uuid UNIQUE,
    username varchar(255) UNIQUE,
    role     user_role
);

INSERT INTO users (uuid, username, role)
VALUES ('965389fb-27ce-4f81-9e59-6ef9cb3b2472', 'admin', 'admin');

CREATE TABLE local_auth_users
(
    id            bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    uuid          uuid REFERENCES users (uuid) ON DELETE CASCADE,
    username      varchar(255) REFERENCES users (username),
    password_hash varchar(255)
);

INSERT INTO local_auth_users (uuid, username, password_hash)
VALUES ('965389fb-27ce-4f81-9e59-6ef9cb3b2472', 'admin',
        '$2y$12$9tWON20iFLgRxqfr4YkaYObrSHlSbgDLlWjXnu0tzom4EvPo35vmS');

CREATE TABLE teams
(
    id   bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    uuid uuid UNIQUE,
    name varchar(255) UNIQUE
);

CREATE TYPE team_member_role AS ENUM ('admin', 'user');

CREATE TABLE team_members
(
    id        bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_uuid uuid REFERENCES users (uuid) ON DELETE CASCADE,
    team_uuid uuid REFERENCES teams (uuid) ON DELETE CASCADE,
    role      team_member_role,
    UNIQUE (user_uuid, team_uuid)
);

CREATE TABLE namespaces
(
    id        bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    uuid      uuid UNIQUE,
    name      varchar(255) UNIQUE,
    user_uuid uuid REFERENCES users (uuid) ON DELETE CASCADE,
    team_uuid uuid REFERENCES teams (uuid) ON DELETE CASCADE
);

INSERT INTO namespaces (uuid, name, user_uuid)
VALUES ('f85d60eb-88b8-485f-ad53-5f214b1b29f8', 'admin', '965389fb-27ce-4f81-9e59-6ef9cb3b2472');

CREATE TYPE repository_visibility AS ENUM ('public', 'private');

CREATE TABLE repositories
(
    id         bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    uuid       uuid UNIQUE,
    namespace  uuid REFERENCES namespaces (uuid) ON DELETE CASCADE,
    name       varchar(255),
    visibility repository_visibility,
    UNIQUE (namespace, name)
);

CREATE TYPE token_permission AS ENUM ('read_only', 'read_write', 'read_write_delete');

CREATE TABLE personal_access_tokens
(
    id              bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    uuid            uuid UNIQUE,
    hash            varchar,
    last_eight      varchar(8),
    description     varchar(255),
    permission      token_permission,
    expiration_date timestamp,
    user_uuid       uuid REFERENCES users (uuid) ON DELETE CASCADE
);

CREATE TABLE personal_access_tokens_usage_log
(
    id         bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    token_uuid uuid REFERENCES personal_access_tokens (uuid) ON DELETE CASCADE,
    source_ip  inet,
    timestamp  timestamp
);
CREATE INDEX ON personal_access_tokens_usage_log (token_uuid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE personal_access_tokens_usage_log;
DROP TABLE personal_access_tokens;
DROP TYPE token_permission;
DROP TABLE repositories;
DROP TYPE repository_visibility;
DROP TABLE namespaces;
DROP TABLE team_members;
DROP TYPE team_member_role;
DROP TABLE teams;
DROP TABLE local_auth_users;
DROP TABLE users;
DROP TYPE user_role;
-- +goose StatementEnd
