CREATE TYPE user_role AS ENUM ('admin', 'user');

CREATE TABLE users
(
    id            uuid PRIMARY KEY,
    username      varchar(255) UNIQUE NOT NULL,
    password_hash varchar(255),
    role          user_role           NOT NULL,
    created_at    timestamptz         NOT NULL DEFAULT now()
);

CREATE TABLE teams
(
    id         uuid PRIMARY KEY,
    name       varchar(255) UNIQUE NOT NULL,
    created_at timestamptz         NOT NULL DEFAULT now()
);

CREATE TYPE team_member_role AS ENUM ('admin', 'user');

CREATE TABLE team_members
(
    id         bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id    uuid REFERENCES users ON DELETE CASCADE NOT NULL,
    team_id    uuid REFERENCES teams ON DELETE CASCADE NOT NULL,
    role       team_member_role                        NOT NULL,
    created_at timestamptz                             NOT NULL DEFAULT now(),
    UNIQUE (user_id, team_id)
);

CREATE TABLE namespaces
(
    id         bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name       varchar(255) UNIQUE NOT NULL,
    user_id    uuid REFERENCES users ON DELETE CASCADE,
    team_id    uuid REFERENCES teams ON DELETE CASCADE,
    created_at timestamptz         NOT NULL DEFAULT now()
);

CREATE TYPE repository_visibility AS ENUM ('public', 'private');

CREATE TABLE repositories
(
    id           uuid PRIMARY KEY,
    namespace_id bigint REFERENCES namespaces ON DELETE CASCADE NOT NULL,
    name         varchar(255)                                   NOT NULL,
    visibility   repository_visibility                          NOT NULL,
    created_at   timestamptz                                    NOT NULL DEFAULT now(),
    UNIQUE (namespace_id, name)
);

CREATE TYPE token_permission AS ENUM ('read_only', 'read_write', 'read_write_delete');

CREATE TABLE personal_access_tokens
(
    id              uuid PRIMARY KEY,
    hash            varchar                                 NOT NULL,
    last_eight      varchar(8)                              NOT NULL,
    description     varchar(255)                            NOT NULL,
    permission      token_permission                        NOT NULL,
    expiration_date timestamp,
    user_id         uuid REFERENCES users ON DELETE CASCADE NOT NULL,
    created_at      timestamptz                             NOT NULL DEFAULT now()
);

CREATE TABLE personal_access_tokens_usage_log
(
    id        bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    token_id  uuid REFERENCES personal_access_tokens ON DELETE CASCADE NOT NULL,
    source_ip inet                                                     NOT NULL,
    timestamp timestamptz                                              NOT NULL
);
CREATE INDEX ON personal_access_tokens_usage_log (token_id);
