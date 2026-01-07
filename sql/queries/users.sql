-- name: UpsertUser :one
INSERT INTO users
    (email, name)
VALUES 
    (sqlc.arg('email'), sqlc.arg('name'))
ON CONFLICT (email) DO UPDATE
    SET email = users.email -- no-op to return existing row
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = sqlc.arg('email');

-- name: GetUserByID :one
SELECT * FROM users WHERE id = sqlc.arg('id');

-- name: ListAllUsersWithAssignmentsByOrganizationID :many
SELECT
    u.id AS user_id,
    ua.organization_id,
    u.email,
    u.name,
    ua.role,
    ua.disabled_at,
    ua.disabled_reason
FROM users u
JOIN user_organization_assignments ua
  ON ua.user_id = u.id
WHERE ua.organization_id = sqlc.arg('organization_id')::UUID
ORDER BY u.email;

-- name: ListEnabledUsersByOrganizationID :many
SELECT u.*
FROM users u
JOIN user_organization_assignments ua
  ON ua.user_id = u.id
WHERE ua.organization_id = sqlc.arg('organization_id')::UUID
  AND ua.disabled_at IS NULL
AND (
    CASE WHEN sqlc.narg('search')::TEXT IS NULL THEN TRUE
    ELSE
        name ILIKE '%' || sqlc.narg('search')::TEXT || '%'
        OR
        email ILIKE '%' || sqlc.narg('search')::TEXT || '%'
    END
)
ORDER BY u.name;


-- name: EnableUser :one
UPDATE user_organization_assignments
SET
    disabled_at = NULL,
    disabled_reason = NULL
WHERE
  user_id = sqlc.arg('user_id')::UUID
  AND
  organization_id = sqlc.arg('organization_id')::UUID
RETURNING *;

-- name: DisableUser :one
UPDATE user_organization_assignments
SET
    disabled_at = now(),
    disabled_reason = sqlc.arg('disabled_reason')::TEXT
WHERE
  user_id = sqlc.arg('user_id')::UUID
  AND
  organization_id = sqlc.arg('organization_id')::UUID
RETURNING *;

-- name: UpdateUserName :one
UPDATE users
SET
    name = sqlc.arg('name')::TEXT
WHERE
  id = sqlc.arg('user_id')::UUID
RETURNING *;

-- name: UpdateUserRole :one
UPDATE user_organization_assignments
SET
    role = sqlc.arg('role')::SMALLINT
WHERE
  user_id = sqlc.arg('user_id')::UUID
  AND
  organization_id = sqlc.arg('organization_id')::UUID
RETURNING *;