-- name: CreateTask :one
INSERT INTO tasks (
    id,
    title,
    description,
    priority,
    due_by,
    assignee_user_id,
    assigner_user_id,
    organization_id
) VALUES (
    sqlc.arg('id'),
    sqlc.arg('title'),
    sqlc.arg('description'),
    sqlc.arg('priority')::SMALLINT,
    sqlc.narg('due_by'),
    COALESCE(sqlc.narg('assignee_user_id')::UUID, sqlc.arg('assigner_user_id')),
    sqlc.arg('assigner_user_id'),
    sqlc.arg('organization_id')
) RETURNING *;

-- name: UpdateTaskAsAssignee :one
UPDATE tasks SET
    notes = COALESCE(sqlc.narg('notes'), notes),
    progress = COALESCE(sqlc.narg('progress'), progress),
    updated_at = now()
WHERE
    id = sqlc.arg('task_id')
    AND organization_id = sqlc.arg('organization_id')
    -- AND assignee_user_id = sqlc.arg('assignee_user_id')
RETURNING *;

-- name: UpdateTaskAsManager :one
UPDATE tasks SET
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    notes = COALESCE(sqlc.narg('notes'), notes),
    due_by = CASE
        WHEN sqlc.arg('clear_due_by')::BOOLEAN THEN NULL
        ELSE COALESCE(sqlc.narg('due_by'), due_by)
    END,
    priority = COALESCE(sqlc.narg('priority'), priority),
    progress = COALESCE(sqlc.narg('progress'), progress),
    assignee_user_id = COALESCE(sqlc.narg('assignee_user_id'), assignee_user_id),
    updated_at = now()
WHERE id = sqlc.arg('id') AND organization_id = sqlc.arg('organization_id')
RETURNING *;

-- name: SetTaskStatus :one
UPDATE tasks
SET status = sqlc.arg('status'), updated_at = now()
WHERE id = sqlc.arg('id') and organization_id = sqlc.arg('organization_id')
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = sqlc.arg('id') and organization_id = sqlc.arg('organization_id');

-- name: GetTaskByID :one
SELECT * FROM tasks
WHERE id = sqlc.arg('id') AND organization_id = sqlc.arg('organization_id');

-- name: ListTasks :many
SELECT * FROM (
    SELECT *
    FROM tasks
    WHERE organization_id = sqlc.arg('organization_id')
    AND (
        (sqlc.arg('scope')::TEXT = 'assigned'
            AND assignee_user_id = sqlc.arg('user_id')
            AND (sqlc.narg('assigner_id')::UUID IS NULL OR assigner_user_id = sqlc.narg('assigner_id')::UUID)
        )
        OR
        (sqlc.arg('scope')::TEXT = 'requested'
            AND assigner_user_id = sqlc.arg('user_id')
            AND assignee_user_id != sqlc.arg('user_id')
            AND (sqlc.narg('assignee_id')::UUID IS NULL OR assignee_user_id = sqlc.narg('assignee_id')::UUID)
        )
    )
    AND (
        COALESCE(sqlc.narg('include_complete')::BOOLEAN, FALSE)
        OR progress < 100
    )
    AND (
        NOT COALESCE(sqlc.narg('past_due')::BOOLEAN, FALSE)
        OR (due_by IS NOT NULL AND due_by < CURRENT_DATE)
    )
    AND (sqlc.narg('priority_min')::SMALLINT IS NULL OR priority >= sqlc.narg('priority_min')::SMALLINT)
    AND (sqlc.narg('since')::TIMESTAMPTZ IS NULL OR created_at >= sqlc.narg('since')::TIMESTAMPTZ)
    ORDER BY created_at DESC
    LIMIT sqlc.arg('limit')::INTEGER
) ORDER BY created_at ASC;