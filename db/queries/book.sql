-- name: CreateBook :one
-- el estado lo pone el DEFAULT de la tabla
INSERT INTO book (title, author)
VALUES ($1, $2)
RETURNING *;

-- name: GetBook :one
SELECT * FROM book WHERE id = $1;

-- name: ListBooks :many
-- el id desempata los títulos repetidos
SELECT * FROM book ORDER BY title, id;

-- name: UpdateBookStatus :one
-- update condicional: solo se aplica si el libro sigue en el estado leído ($6)
UPDATE book
SET status = $2, read_date = $3, rating = $4, notes = $5
WHERE id = $1 AND status = $6
RETURNING *;

-- name: DeleteBook :one
-- sin fila borrada, el libro no existía (404)
DELETE FROM book WHERE id = $1
RETURNING *;
