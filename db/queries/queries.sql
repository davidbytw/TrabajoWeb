-- name: CreateBook :one
INSERT INTO books (title, author, status, read_date, rating, notes)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetBook :one
SELECT * FROM books WHERE id = $1;

-- name: ListBooks :many
SELECT * FROM books ORDER BY title;

-- name: UpdateBook :exec
UPDATE books
SET status = $2, read_date = $3, rating = $4, notes = $5
WHERE id = $1;

-- name: DeleteBook :exec
DELETE FROM books WHERE id = $1;