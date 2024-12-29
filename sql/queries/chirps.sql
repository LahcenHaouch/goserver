-- name: CreateChirp :one
INSERT INTO chirps(id, created_at, updated_at, body, user_id) VALUES (
    gen_random_uuid (), NOW(), NOW(), $1, $2
)
returning *;

-- name: GetChirps :many
SELECT * from chirps ORDER BY created_at ASC;

-- name: GetChirpsDESC :many
SELECT * from chirps ORDER BY created_at DESC;

-- name: GetChirp :one
SELECT * from chirps WHERE id = $1;

-- name: GetChirpsByAuthorId :many
SELECT * from chirps WHERE user_id = $1 ORDER BY created_at ASC;

-- name: GetChirpsByAuthorIdDESC :many
SELECT * from chirps WHERE user_id = $1 ORDER BY created_at DESC;

-- name: DeleteChirp :exec
DELETE FROM chirps WHERE id = $1;
