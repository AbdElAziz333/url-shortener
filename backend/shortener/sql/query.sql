-- name: CreateLink :one
INSERT INTO link (user_id, code, original_url, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListLinksByUser :many
SELECT * FROM link
WHERE user_id = $1 AND is_active
ORDER BY created_at DESC, id
LIMIT $2 OFFSET $3;

-- name: UpdateLinkExpiry :one
UPDATE link SET expires_at = $3
WHERE code = $1 AND user_id = $2 AND is_active
RETURNING *;

-- name: DeactivateLink :execrows
UPDATE link SET is_active = false
WHERE code = $1 AND user_id = $2 AND is_active;

-- name: GetActiveLinkByCode :one
SELECT original_url FROM link
WHERE code = $1 AND is_active AND (expires_at IS NULL OR expires_at > now());

-- name: GetLinkByCode :one
SELECT * FROM link
WHERE code = $1 AND user_id = $2 AND is_active;