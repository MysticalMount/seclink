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

-- name: GetLink :one
SELECT sqlc.embed(links), sqlc.embed(posts)
FROM links
JOIN posts ON links.post_name = posts.name
WHERE links.id = ?;

-- name: GetPost :one
SELECT * FROM posts WHERE name = ?;

-- name: GetAllPosts :many
SELECT * FROM posts;

-- name: GetAllLinks :many
SELECT sqlc.embed(posts), sqlc.embed(links)
FROM links
JOIN posts ON links.post_name = posts.name;
