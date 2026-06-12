CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,
    name        TEXT        NOT NULL,
    email       TEXT        NOT NULL,
    avatar_url  TEXT,
    provider    TEXT        NOT NULL,
    provider_id TEXT        NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email      ON users (email);
CREATE        INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);

CREATE TABLE IF NOT EXISTS organizations (
    id               BIGSERIAL PRIMARY KEY,
    created_at       TIMESTAMPTZ,
    updated_at       TIMESTAMPTZ,
    deleted_at       TIMESTAMPTZ,
    name             TEXT    NOT NULL,
    description      TEXT,
    whatsapp_number  TEXT    NOT NULL,
    prefix           TEXT    NOT NULL DEFAULT 'TSK',
    owner_id         BIGINT  NOT NULL REFERENCES users (id)
);
CREATE INDEX IF NOT EXISTS idx_organizations_owner_id   ON organizations (owner_id);
CREATE INDEX IF NOT EXISTS idx_organizations_deleted_at ON organizations (deleted_at);

CREATE TABLE IF NOT EXISTS team_members (
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ,
    user_id         BIGINT NOT NULL REFERENCES users         (id),
    organization_id BIGINT NOT NULL REFERENCES organizations (id),
    role            TEXT   NOT NULL DEFAULT 'member'
);
CREATE INDEX IF NOT EXISTS idx_team_members_user_id         ON team_members (user_id);
CREATE INDEX IF NOT EXISTS idx_team_members_organization_id ON team_members (organization_id);
CREATE INDEX IF NOT EXISTS idx_team_members_deleted_at      ON team_members (deleted_at);

CREATE TABLE IF NOT EXISTS projects (
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ,
    name            TEXT   NOT NULL,
    description     TEXT,
    start_date      TIMESTAMPTZ,
    organization_id BIGINT NOT NULL REFERENCES organizations (id),
    created_by_id   BIGINT NOT NULL REFERENCES users         (id)
);
CREATE INDEX IF NOT EXISTS idx_projects_organization_id ON projects (organization_id);
CREATE INDEX IF NOT EXISTS idx_projects_deleted_at      ON projects (deleted_at);

CREATE TABLE IF NOT EXISTS states (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    name       TEXT    NOT NULL,
    color      TEXT    NOT NULL DEFAULT '#6B7280',
    position   BIGINT  NOT NULL DEFAULT 0,
    is_done    BOOLEAN NOT NULL DEFAULT FALSE,
    project_id BIGINT  NOT NULL REFERENCES projects (id)
);
CREATE INDEX IF NOT EXISTS idx_states_project_id  ON states (project_id);
CREATE INDEX IF NOT EXISTS idx_states_deleted_at  ON states (deleted_at);

CREATE TABLE IF NOT EXISTS sprints (
    id           BIGSERIAL PRIMARY KEY,
    created_at   TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ,
    name         TEXT        NOT NULL,
    goal         TEXT,
    project_id   BIGINT      NOT NULL REFERENCES projects (id),
    start_date   TIMESTAMPTZ NOT NULL,
    end_date     TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_sprints_project_id  ON sprints (project_id);
CREATE INDEX IF NOT EXISTS idx_sprints_deleted_at  ON sprints (deleted_at);

CREATE TABLE IF NOT EXISTS burndown_snapshots (
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ,
    sprint_id       BIGINT      NOT NULL REFERENCES sprints (id),
    date            TIMESTAMPTZ NOT NULL,
    remaining_tasks BIGINT      NOT NULL,
    completed_tasks BIGINT      NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_burndown_snapshots_sprint_id  ON burndown_snapshots (sprint_id);
CREATE INDEX IF NOT EXISTS idx_burndown_snapshots_deleted_at ON burndown_snapshots (deleted_at);

CREATE TABLE IF NOT EXISTS tasks (
    id           BIGSERIAL PRIMARY KEY,
    created_at   TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ,
    task_number  BIGINT  NOT NULL DEFAULT 0,
    title        TEXT    NOT NULL,
    priority     TEXT    NOT NULL DEFAULT 'medium',
    sprint_id    BIGINT           REFERENCES sprints  (id),
    description  TEXT,
    state_id     BIGINT  NOT NULL REFERENCES states   (id),
    project_id   BIGINT  NOT NULL REFERENCES projects (id),
    assigned_to  BIGINT           REFERENCES users    (id),
    created_by_id BIGINT NOT NULL REFERENCES users    (id),
    start_date   TIMESTAMPTZ,
    due_date     TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_tasks_sprint_id    ON tasks (sprint_id);
CREATE INDEX IF NOT EXISTS idx_tasks_state_id     ON tasks (state_id);
CREATE INDEX IF NOT EXISTS idx_tasks_project_id   ON tasks (project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_assigned_to  ON tasks (assigned_to);
CREATE INDEX IF NOT EXISTS idx_tasks_due_date     ON tasks (due_date);
CREATE INDEX IF NOT EXISTS idx_tasks_completed_at ON tasks (completed_at);
CREATE INDEX IF NOT EXISTS idx_tasks_deleted_at   ON tasks (deleted_at);

CREATE TABLE IF NOT EXISTS comments (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    content    TEXT   NOT NULL,
    task_id    BIGINT NOT NULL REFERENCES tasks (id),
    author_id  BIGINT NOT NULL REFERENCES users (id)
);
CREATE INDEX IF NOT EXISTS idx_comments_task_id    ON comments (task_id);
CREATE INDEX IF NOT EXISTS idx_comments_deleted_at ON comments (deleted_at);

CREATE TABLE IF NOT EXISTS invitations (
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ,
    email           TEXT        NOT NULL,
    organization_id BIGINT      NOT NULL REFERENCES organizations (id),
    token           TEXT        NOT NULL,
    role            TEXT        NOT NULL DEFAULT 'manager',
    expires_at      TIMESTAMPTZ NOT NULL,
    accepted_at     TIMESTAMPTZ,
    declined_at     TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_invitations_token           ON invitations (token);
CREATE        INDEX IF NOT EXISTS idx_invitations_organization_id ON invitations (organization_id);
CREATE        INDEX IF NOT EXISTS idx_invitations_deleted_at      ON invitations (deleted_at);

CREATE TABLE IF NOT EXISTS git_hub_integrations (
    id                  BIGSERIAL PRIMARY KEY,
    created_at          TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ,
    deleted_at          TIMESTAMPTZ,
    organization_id     BIGINT NOT NULL REFERENCES organizations (id),
    repo_owner          TEXT   NOT NULL,
    repo_name           TEXT   NOT NULL,
    project_id          BIGINT NOT NULL REFERENCES projects      (id),
    webhook_secret      TEXT   NOT NULL,
    in_review_state_id  BIGINT NOT NULL,
    done_state_id       BIGINT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_git_hub_integrations_organization_id ON git_hub_integrations (organization_id);
CREATE INDEX IF NOT EXISTS idx_git_hub_integrations_project_id      ON git_hub_integrations (project_id);
CREATE INDEX IF NOT EXISTS idx_git_hub_integrations_deleted_at      ON git_hub_integrations (deleted_at);

CREATE TABLE IF NOT EXISTS api_keys (
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,
    name        TEXT   NOT NULL,
    key_hash    TEXT   NOT NULL,
    key_prefix  TEXT   NOT NULL,
    user_id     BIGINT NOT NULL REFERENCES users         (id),
    org_id      BIGINT NOT NULL REFERENCES organizations (id),
    last_used_at TIMESTAMPTZ,
    revoked_at  TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_key_hash   ON api_keys (key_hash);
CREATE        INDEX IF NOT EXISTS idx_api_keys_user_id    ON api_keys (user_id);
CREATE        INDEX IF NOT EXISTS idx_api_keys_org_id     ON api_keys (org_id);
CREATE        INDEX IF NOT EXISTS idx_api_keys_deleted_at ON api_keys (deleted_at);

CREATE TABLE IF NOT EXISTS outbound_webhooks (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    org_id     BIGINT  NOT NULL REFERENCES organizations (id),
    url        TEXT    NOT NULL,
    secret     TEXT    NOT NULL,
    events     TEXT    NOT NULL DEFAULT 'task.moved,task.created',
    active     BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_outbound_webhooks_org_id     ON outbound_webhooks (org_id);
CREATE INDEX IF NOT EXISTS idx_outbound_webhooks_deleted_at ON outbound_webhooks (deleted_at);

CREATE TABLE IF NOT EXISTS notification_configs (
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,
    org_id      BIGINT  NOT NULL,
    channel     TEXT    NOT NULL,
    webhook_url TEXT,
    enabled     BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_notification_configs_org_id     ON notification_configs (org_id);
CREATE INDEX IF NOT EXISTS idx_notification_configs_deleted_at ON notification_configs (deleted_at);

CREATE TABLE IF NOT EXISTS notification_preferences (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    user_id    BIGINT  NOT NULL,
    org_id     BIGINT  NOT NULL,
    event      TEXT    NOT NULL,
    channel    TEXT    NOT NULL,
    enabled    BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_notification_preferences_user_id    ON notification_preferences (user_id);
CREATE INDEX IF NOT EXISTS idx_notification_preferences_org_id     ON notification_preferences (org_id);
CREATE INDEX IF NOT EXISTS idx_notification_preferences_deleted_at ON notification_preferences (deleted_at);

CREATE TABLE IF NOT EXISTS state_transitions (
    id                    BIGSERIAL PRIMARY KEY,
    created_at            TIMESTAMPTZ,
    updated_at            TIMESTAMPTZ,
    deleted_at            TIMESTAMPTZ,
    task_id               BIGINT NOT NULL REFERENCES tasks (id),
    from_state_id         BIGINT,
    to_state_id           BIGINT NOT NULL,
    changed_by_id         BIGINT NOT NULL,
    seconds_in_from_state BIGINT
);
CREATE INDEX IF NOT EXISTS idx_state_transitions_task_id    ON state_transitions (task_id);
CREATE INDEX IF NOT EXISTS idx_state_transitions_deleted_at ON state_transitions (deleted_at);

CREATE TABLE IF NOT EXISTS subtasks (
    id            BIGSERIAL PRIMARY KEY,
    created_at    TIMESTAMPTZ,
    updated_at    TIMESTAMPTZ,
    deleted_at    TIMESTAMPTZ,
    title         TEXT    NOT NULL,
    description   TEXT,
    done          BOOLEAN NOT NULL DEFAULT FALSE,
    task_id       BIGINT  NOT NULL REFERENCES tasks (id),
    created_by_id BIGINT  NOT NULL REFERENCES users (id)
);
CREATE INDEX IF NOT EXISTS idx_subtasks_task_id    ON subtasks (task_id);
CREATE INDEX IF NOT EXISTS idx_subtasks_deleted_at ON subtasks (deleted_at);

CREATE TABLE IF NOT EXISTS subtask_comments (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    content    TEXT   NOT NULL,
    subtask_id BIGINT NOT NULL REFERENCES subtasks (id),
    author_id  BIGINT NOT NULL REFERENCES users    (id)
);
CREATE INDEX IF NOT EXISTS idx_subtask_comments_subtask_id ON subtask_comments (subtask_id);
CREATE INDEX IF NOT EXISTS idx_subtask_comments_deleted_at ON subtask_comments (deleted_at);

CREATE TABLE IF NOT EXISTS time_entries (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    task_id    BIGINT      REFERENCES tasks    (id),
    subtask_id BIGINT      REFERENCES subtasks (id),
    user_id    BIGINT NOT NULL REFERENCES users (id),
    minutes    BIGINT      NOT NULL,
    note       TEXT,
    logged_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_time_entries_task_id    ON time_entries (task_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_subtask_id ON time_entries (subtask_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_user_id    ON time_entries (user_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_logged_at  ON time_entries (logged_at);
CREATE INDEX IF NOT EXISTS idx_time_entries_deleted_at ON time_entries (deleted_at);

CREATE TABLE IF NOT EXISTS active_timers (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    user_id    BIGINT      NOT NULL REFERENCES users    (id),
    task_id    BIGINT           REFERENCES tasks    (id),
    subtask_id BIGINT           REFERENCES subtasks (id),
    started_at TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_active_timers_user_id    ON active_timers (user_id);
CREATE        INDEX IF NOT EXISTS idx_active_timers_task_id    ON active_timers (task_id);
CREATE        INDEX IF NOT EXISTS idx_active_timers_subtask_id ON active_timers (subtask_id);
CREATE        INDEX IF NOT EXISTS idx_active_timers_deleted_at ON active_timers (deleted_at);

DROP TABLE IF EXISTS schema_versions;
