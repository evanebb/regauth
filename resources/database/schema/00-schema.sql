CREATE TYPE user_role AS ENUM ('admin', 'user');

CREATE TABLE users
(
    id            uuid PRIMARY KEY    NOT NULL,
    username      varchar(255) UNIQUE NOT NULL,
    password_hash varchar(255),
    role          user_role           NOT NULL
);

CREATE TABLE teams
(
    id   uuid PRIMARY KEY    NOT NULL,
    name varchar(255) UNIQUE NOT NULL
);

CREATE TYPE team_member_role AS ENUM ('admin', 'user');

CREATE TABLE team_members
(
    id      bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id uuid REFERENCES users ON DELETE CASCADE NOT NULL,
    team_id uuid REFERENCES teams ON DELETE CASCADE NOT NULL,
    role    team_member_role                        NOT NULL,
    UNIQUE (user_id, team_id)
);

CREATE TABLE namespaces
(
    id      uuid PRIMARY KEY    NOT NULL,
    name    varchar(255) UNIQUE NOT NULL,
    user_id uuid REFERENCES users ON DELETE CASCADE,
    team_id uuid REFERENCES teams ON DELETE CASCADE
);

CREATE TYPE repository_visibility AS ENUM ('public', 'private');

CREATE TABLE repositories
(
    id           uuid PRIMARY KEY                             NOT NULL,
    namespace_id uuid REFERENCES namespaces ON DELETE CASCADE NOT NULL,
    name         varchar(255)                                 NOT NULL,
    visibility   repository_visibility                        NOT NULL,
    UNIQUE (namespace_id, name)
);

CREATE TYPE token_permission AS ENUM ('read_only', 'read_write', 'read_write_delete');

CREATE TABLE personal_access_tokens
(
    id              uuid PRIMARY KEY                        NOT NULL,
    hash            varchar                                 NOT NULL,
    last_eight      varchar(8)                              NOT NULL,
    description     varchar(255)                            NOT NULL,
    permission      token_permission                        NOT NULL,
    expiration_date timestamp,
    user_id         uuid REFERENCES users ON DELETE CASCADE NOT NULL
);

CREATE TABLE personal_access_tokens_usage_log
(
    id        bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    token_id  uuid REFERENCES personal_access_tokens ON DELETE CASCADE NOT NULL,
    source_ip inet                                                     NOT NULL,
    timestamp timestamp                                                NOT NULL
);
CREATE INDEX ON personal_access_tokens_usage_log (token_id);
