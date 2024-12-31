-- name: CreateUser :one
INSERT INTO users (
    email,
    password_hash,
    first_name,
    last_name
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET 
    first_name = COALESCE($2, first_name),
    last_name = COALESCE($3, last_name),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
