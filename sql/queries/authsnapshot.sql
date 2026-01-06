-- name: GetAuthSnapshot :one
SELECT
    u.id AS user_id,
    u.email as email,
    u.name as name,
    o.id AS organization_id,
    o.name AS organization_name,
    o.domain AS organization_domain,
    ua.role AS role,
    CASE
        WHEN ua.disabled_at IS NOT NULL THEN TRUE
        ELSE FALSE
    END AS disabled
FROM users u
JOIN user_organization_assignments ua
  ON ua.user_id = u.id
JOIN organizations o
  ON o.id = ua.organization_id
WHERE u.id = sqlc.arg('user_id') AND o.id = sqlc.arg('organization_id')
LIMIT 1;
