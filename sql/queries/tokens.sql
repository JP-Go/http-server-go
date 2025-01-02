-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at)
VALUES ($1, now(), now(), $2, $3)
RETURNING token;

-- name: GetUserFromRefreshToken :one
SELECT users.*, 
    tokens.token, 
    tokens.expires_at, 
    tokens.revoked_at
FROM refresh_tokens tokens
INNER JOIN users ON tokens.user_id = users.id
WHERE tokens.token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens 
SET 
    revoked_at = now(), 
    updated_at = now()
WHERE token = $1;
