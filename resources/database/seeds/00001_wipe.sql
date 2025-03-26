-- +goose Up
-- +goose StatementBegin
TRUNCATE users, teams, team_members, namespaces, repositories, personal_access_tokens, personal_access_tokens_usage_log RESTART IDENTITY;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
TRUNCATE users, teams, team_members, namespaces, repositories, personal_access_tokens, personal_access_tokens_usage_log RESTART IDENTITY;
-- +goose StatementEnd
