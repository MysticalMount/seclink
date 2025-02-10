-- name: GetAllLinks :many
SELECT * FROM links
ORDER BY id;

-- name: CreateLink :exec
INSERT INTO links (id, expires, post_name)
VALUES (?, ?, ?);

-- name: DeleteLink :exec
DELETE FROM links WHERE id = ?;

-- name: CreatePost :exec
INSERT INTO posts (name, path)
VALUES (?, ?);

-- name: DeletePost :exec
DELETE FROM posts WHERE name = ?;

-- -- name: ListAuthors :many
-- SELECT * FROM authors
-- ORDER BY name;

-- -- name: CreateAuthor :one
-- INSERT INTO authors (
--   name, bio
-- ) VALUES (
--   ?, ?
-- )
-- RETURNING *;

-- -- name: UpdateAuthor :exec
-- UPDATE authors
-- set name = ?,
-- bio = ?
-- WHERE id = ?;

-- -- name: DeleteAuthor :exec
-- DELETE FROM authors
-- WHERE id = ?;