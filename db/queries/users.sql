-- name: CreateUser :one
INSERT INTO users (email, phone, password_hash, full_name)
VALUES ($1, $2, $3, $4)
    RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateLastLogin :exec
UPDATE users SET last_login_at = NOW(), updated_at = NOW() WHERE id = $1;