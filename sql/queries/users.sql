-- name: CreateUser :one
INSERT INTO users(id, created_at, updated_at, email, hashed_password)
VALUES(gen_random_uuid(), NOW(), NOW(), @email::TEXT, @hashed_password::TEXT)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByRefreshToken :one
SELECT users.* 
FROM users INNER JOIN refresh_tokens 
ON users.id = refresh_tokens.user_id
WHERE refresh_tokens.token = $1;

-- name: UpdateUser :one
UPDATE users
SET email = $2, hashed_password = $3
WHERE id = $1
RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM users;