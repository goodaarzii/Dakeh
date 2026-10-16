-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, user_agent, expires_at)
VALUES (sqlc.arg(user_id), sqlc.arg(token_hash), sqlc.arg(user_agent), sqlc.arg(expires_at))
    RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_tokens WHERE token_hash = $1;

-- Revoke ONE token, but only if it is still valid.
-- 0 rows means someone already used it (a race between two requests, or a replay): treat it as invalid.
-- name: RevokeRefreshToken :execrows
UPDATE refresh_tokens SET revoked_at = NOW()
WHERE id = $1 AND revoked_at IS NULL;

-- Logout. 0 rows is fine (already revoked or unknown): logout stays idempotent.
-- name: RevokeRefreshTokenByHash :execrows
UPDATE refresh_tokens SET revoked_at = NOW()
WHERE token_hash = $1 AND revoked_at IS NULL;

-- Log out everywhere Also the reaction to token REUSE (see the notes).
-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens SET revoked_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL;

-- For a cleanup job later: tokens that expired more than 7 days ago are useless.
-- name: DeleteOldRefreshTokens :execrows
DELETE FROM refresh_tokens WHERE expires_at < NOW() - INTERVAL '7 days';