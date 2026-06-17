-- name: CountUsers :one
SELECT count(*)::bigint FROM users;

