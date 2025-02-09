CREATE TABLE IF NOT EXISTS "posts" (
    name TEXT PRIMARY KEY,
    path TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS "links" (
    id TEXT PRIMARY KEY,
    expires INTEGER NOT NULL,
    post_name TEXT NOT NULL,
    FOREIGN KEY (post_name) REFERENCES posts(name)
);