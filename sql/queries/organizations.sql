-- name: GetOrganizationByID :one
SELECT * FROM organizations WHERE id = sqlc.arg('id');

-- name: GetOrganizationByDomain :one
SELECT * FROM organizations WHERE domain = sqlc.arg('domain');

-- name: UpsertOrganization :one
INSERT INTO organizations 
    (name, domain)
VALUES 
    (sqlc.arg('name'), sqlc.arg('domain'))
ON CONFLICT (domain) DO UPDATE
    SET domain = organizations.domain -- no-op to return existing row
RETURNING *;

-- name: UpdateOrganization :one
UPDATE organizations
SET
    name = COALESCE(sqlc.narg('name'), name)
WHERE
    id = sqlc.arg('organization_id')
RETURNING *;

-- name: AssignUserToOrganization :exec
INSERT INTO user_organization_assignments
    (user_id, organization_id, role)
VALUES
    (sqlc.arg('user_id'), sqlc.arg('organization_id'), sqlc.arg('role'))
ON CONFLICT (user_id, organization_id) DO NOTHING;

-- name: AssignUserToOrganizationTryOwner :exec
INSERT INTO user_organization_assignments
    (user_id, organization_id, role)
VALUES
    (sqlc.arg('user_id'), sqlc.arg('organization_id'), 3)
ON CONFLICT DO NOTHING;

-- name: GetUserOrganizationAssignment :one
SELECT * FROM user_organization_assignments
WHERE user_id = sqlc.arg('user_id') AND organization_id = sqlc.arg('organization_id') LIMIT 1;

-- name: GetOrganizationByUserID :one
SELECT
    o.*
FROM
    organizations o
JOIN
    user_organization_assignments ua
ON
    ua.organization_id = o.id
WHERE
    ua.user_id = sqlc.arg('user_id')
LIMIT 1;

-- name: GetOrgUserRole :one
SELECT
    role, 
    CASE
        WHEN ua.disabled_at IS NOT NULL THEN TRUE
        ELSE FALSE
    END AS disabled
FROM
    user_organization_assignments ua
WHERE
    user_id = sqlc.arg('user_id') AND organization_id = sqlc.arg('organization_id')
LIMIT 1;