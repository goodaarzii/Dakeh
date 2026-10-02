-- name: CreateAddress :one
INSERT INTO addresses (
    user_id, label, recipient_name, phone, country, province, city, street, postal_code, lat, lng, is_default
) VALUES (
             sqlc.arg(user_id), sqlc.arg(label), sqlc.arg(recipient_name), sqlc.arg(phone), sqlc.arg(country),
             sqlc.arg(province), sqlc.arg(city), sqlc.arg(street), sqlc.arg(postal_code),
             sqlc.arg(lat), sqlc.arg(lng), sqlc.arg(is_default)
         )
    RETURNING *;

-- name: GetAddressByID :one
SELECT * FROM addresses
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: ListAddressesByUser :many
SELECT * FROM addresses
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY is_default DESC, created_at DESC;

-- name: UpdateAddress :one
UPDATE addresses SET
                     label          = sqlc.arg(label),
                     recipient_name = sqlc.arg(recipient_name),
                     phone          = sqlc.arg(phone),
                     country        = sqlc.arg(country),
                     province       = sqlc.arg(province),
                     city           = sqlc.arg(city),
                     street         = sqlc.arg(street),
                     postal_code    = sqlc.arg(postal_code),
                     lat            = sqlc.arg(lat),
                     lng            = sqlc.arg(lng),
                     updated_at     = NOW()
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id) AND deleted_at IS NULL
    RETURNING *;

-- name: SoftDeleteAddress :execrows
UPDATE addresses
SET deleted_at = NOW(), is_default = FALSE, updated_at = NOW()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: ClearDefaultAddress :exec
UPDATE addresses
SET is_default = FALSE, updated_at = NOW()
WHERE user_id = $1 AND is_default AND deleted_at IS NULL;

-- name: SetDefaultAddress :execrows
UPDATE addresses
SET is_default = TRUE, updated_at = NOW()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;