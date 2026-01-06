CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_sessions_expiry ON sessions (expiry);

-- Organizations represent a group of users working together
CREATE TABLE organizations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       CITEXT NOT NULL,
    domain     CITEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Users represent individuals who can create and be assigned tasks.
CREATE TABLE users (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email        CITEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Association table linking users to organizations with roles and disabled status.
CREATE TABLE user_organization_assignments (
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role            SMALLINT NOT NULL DEFAULT 1, -- 1: member, 2: admin, 3: owner
    disabled_at     TIMESTAMPTZ DEFAULT NULL,
    disabled_reason TEXT DEFAULT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, organization_id)
);

-- Ensure role is within valid range
ALTER TABLE user_organization_assignments
ADD CONSTRAINT user_organization_role_check CHECK (role BETWEEN 1 AND 3);

-- Ensure only one owner per organization
CREATE UNIQUE INDEX user_org_single_owner_per_org
ON user_organization_assignments (organization_id)
WHERE role = 3;

-- Tasks represent individual work items.
CREATE TABLE tasks (
    id               CHAR(26) PRIMARY KEY, -- ULID
    title            TEXT NOT NULL, -- Brief title of the task
    description      TEXT NOT NULL DEFAULT '', -- Detailed description if any
    notes            TEXT NOT NULL DEFAULT '', -- Notes added during task progress
    due_by           TIMESTAMPTZ, -- Optional target completion date
    priority         SMALLINT NOT NULL DEFAULT 2, -- 1..4 (low, medium, high, urgent)
    progress         SMALLINT NOT NULL DEFAULT 0, -- 0..100 (percent complete)
    status           SMALLINT NOT NULL GENERATED ALWAYS AS ( -- Generated status based on progress value
                        CASE
                            WHEN progress = 100 THEN 2 -- completed
                            WHEN progress = 0 THEN 0   -- not started
                            ELSE 1                     -- in progress
                        END
                     ) STORED,
    assignee_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    assigner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    organization_id  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    -- completed_at     TIMESTAMPTZ DEFAULT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Constraints for task fields
ALTER TABLE tasks
    ADD CONSTRAINT tasks_priority_check CHECK (priority BETWEEN 1 AND 4),
    ADD CONSTRAINT tasks_progress_check CHECK (progress BETWEEN 0 AND 100);


-- Ensure assignee is part of the organization
ALTER TABLE tasks
    ADD CONSTRAINT tasks_assignee_in_org
    FOREIGN KEY (assignee_user_id, organization_id)
    REFERENCES user_organization_assignments (user_id, organization_id)
    ON DELETE RESTRICT;

-- Ensure assigner is part of the organization
ALTER TABLE tasks
    ADD CONSTRAINT tasks_assigner_in_org
    FOREIGN KEY (assigner_user_id, organization_id)
    REFERENCES user_organization_assignments (user_id, organization_id)
    ON DELETE RESTRICT;

-- Indexes
CREATE INDEX idx_user_org_assignments_org
ON user_organization_assignments (organization_id);

CREATE INDEX idx_tasks_my_tasks
ON tasks (assignee_user_id, status, priority, due_by);

CREATE INDEX idx_tasks_my_requests
ON tasks (assigner_user_id, created_at);