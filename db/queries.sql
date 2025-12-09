-- name: CreateUser :one
INSERT INTO users (username, password_hash, created_at)
VALUES (?, ?, datetime('now'))
RETURNING id, username, password_hash, created_at;

-- name: GetUserByUsername :one
SELECT id, username, password_hash, created_at
FROM users
WHERE username = ?
LIMIT 1;

-- name: GetUserByID :one
SELECT id, username, password_hash, created_at
FROM users
WHERE id = ?
LIMIT 1;

-- name: CreateSession :one
INSERT INTO sessions (user_id, session_token, created_at, expires_at)
VALUES (?, ?, datetime('now'), ?)
RETURNING id, user_id, session_token, created_at, expires_at;

-- name: GetSessionByToken :one
SELECT id, user_id, session_token, created_at, expires_at
FROM sessions
WHERE session_token = ?
LIMIT 1;

-- name: DeleteSessionByToken :exec
DELETE FROM sessions WHERE session_token = ?;

-- name: CountUsersByIP :one
SELECT COUNT(*) FROM users
WHERE ip = ?;

-- name: CreateUserWithIP :one
INSERT INTO users (username, password_hash, ip, created_at)
VALUES (?, ?, ?, datetime('now'))
RETURNING id, username, password_hash, created_at;

-- name: CreateUserWithRole :one
INSERT INTO users (username, password_hash, ip, role, created_at)
VALUES (?, ?, ?, ?, datetime('now'))
RETURNING id, username, password_hash, created_at, role;

-- name: DeleteUserByUsername :exec
DELETE FROM users WHERE username = ?;

-- name: UpdateUserRole :one
UPDATE users
SET role = ?
WHERE username = ?
RETURNING id, username, password_hash, created_at, role;

-- name: ListUsers :many
SELECT id, username, password_hash, created_at, role
FROM users
ORDER BY created_at DESC;

-- name: GetUserByUsernameWithRole :one
SELECT id, username, password_hash, created_at, role
FROM users
WHERE username = ?
LIMIT 1;

