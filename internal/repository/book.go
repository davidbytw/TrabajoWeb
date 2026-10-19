package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"booklog/db/sqlc"
	"booklog/logic"
)

// PostgresBookRepository implementa logic.BookRepository con las consultas de sqlc
type PostgresBookRepository struct {
	queries *sqlc.Queries
}

// NewPostgresBookRepository crea el repositorio sobre las consultas de sqlc
func NewPostgresBookRepository(queries *sqlc.Queries) *PostgresBookRepository {
	return &PostgresBookRepository{queries: queries}
}

// Create inserta el libro y lo devuelve con el id que generó la base
func (r *PostgresBookRepository) Create(ctx context.Context, title, author string) (logic.Book, error) {
	row, err := r.queries.CreateBook(ctx, sqlc.CreateBookParams{Title: title, Author: author})
	if err != nil {
		return logic.Book{}, fmt.Errorf("crear libro: %w", err)
	}
	return toBook(row), nil
}

// Get devuelve el libro con ese id, o logic.ErrBookNotFound
func (r *PostgresBookRepository) Get(ctx context.Context, id int32) (logic.Book, error) {
	row, err := r.queries.GetBook(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return logic.Book{}, logic.ErrBookNotFound
	}
	if err != nil {
		return logic.Book{}, fmt.Errorf("obtener libro %d: %w", id, err)
	}
	return toBook(row), nil
}

// List devuelve todos los libros ordenados por título
func (r *PostgresBookRepository) List(ctx context.Context) ([]logic.Book, error) {
	rows, err := r.queries.ListBooks(ctx)
	if err != nil {
		return nil, fmt.Errorf("listar libros: %w", err)
	}
	books := make([]logic.Book, 0, len(rows))
	for _, row := range rows {
		books = append(books, toBook(row))
	}
	return books, nil
}

// UpdateStatus guarda estado, fecha, puntaje y notas solo si el libro sigue en currentStatus
func (r *PostgresBookRepository) UpdateStatus(ctx context.Context, book logic.Book, currentStatus string) (logic.Book, error) {
	row, err := r.queries.UpdateBookStatus(ctx, sqlc.UpdateBookStatusParams{
		ID:       book.ID,
		Status:   book.Status, // estado nuevo ($2)
		ReadDate: toNullTime(book.ReadDate),
		Rating:   toNullInt32(book.Rating),
		Notes:    toNullString(book.Notes),
		Status_2: currentStatus, // estado leído antes de validar ($6)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return logic.Book{}, logic.ErrBookNotFound
	}
	if err != nil {
		return logic.Book{}, fmt.Errorf("actualizar libro %d: %w", book.ID, err)
	}
	return toBook(row), nil
}

// Delete borra el libro, o devuelve logic.ErrBookNotFound si no existía
func (r *PostgresBookRepository) Delete(ctx context.Context, id int32) error {
	_, err := r.queries.DeleteBook(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return logic.ErrBookNotFound
	}
	if err != nil {
		return fmt.Errorf("borrar libro %d: %w", id, err)
	}
	return nil
}

// conversiones entre los punteros del dominio (nil) y las cajas sql.Null* de sqlc (Valid false)

func toBook(row sqlc.Book) logic.Book {
	book := logic.Book{
		ID:     row.ID,
		Title:  row.Title,
		Author: row.Author,
		Status: row.Status,
	}
	if row.ReadDate.Valid {
		readDate := row.ReadDate.Time
		book.ReadDate = &readDate
	}
	if row.Rating.Valid {
		rating := row.Rating.Int32
		book.Rating = &rating
	}
	if row.Notes.Valid {
		notes := row.Notes.String
		book.Notes = &notes
	}
	return book
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func toNullInt32(n *int32) sql.NullInt32 {
	if n == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *n, Valid: true}
}

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}
