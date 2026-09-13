package main

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"booklog/db/sqlc"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// TestQueries_CRUD verifica el ciclo completo de operaciones CRUD sobre books
// usando el código generado por sqlc: crear, leer, actualizar, listar y borrar.
func TestQueries_CRUD(t *testing.T) {
	connStr := "host=localhost port=5432 user=books_user password=books_pass dbname=books_db sslmode=disable"
	conn, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("sql.Open() devolvió error: %s", err)
	}
	defer conn.Close()

	// sql.Open no conecta: solo valida los parámetros. Ping es el que confirma
	// que la base está levantada y responde.
	if err := conn.Ping(); err != nil {
		t.Fatalf("conn.Ping() devolvió error: %s", err)
	}

	queries := sqlc.New(conn)
	ctx := context.Background()

	const title = "El nombre del viento"
	const author = "Patrick Rothfuss"

	// 1. Crear. Va fuera de los subtests porque es la base de todos los demás:
	// si falla acá, seguir no tiene sentido.
	created, err := queries.CreateBook(ctx, sqlc.CreateBookParams{
		Title:    title,
		Author:   author,
		Status:   "no_leido",
		ReadDate: sql.NullTime{},   // NULL: todavía no se leyó
		Rating:   sql.NullInt32{},  // NULL
		Notes:    sql.NullString{}, // NULL
	})
	if err != nil {
		t.Fatalf("CreateBook(ctx, params) devolvió error: %s", err)
	}
	if created.ID == 0 || created.Title != title || created.Rating.Valid {
		t.Fatalf("CreateBook(ctx, params) = %+v; se esperaba un ID generado, título %q y rating NULL",
			created, title)
	}
	// Si falla algo: los CHECK del esquema tienen que rechazar estos valores.
	// Es la única verificación de que la validación vive en la base y funciona.
	t.Run("Constraints", func(t *testing.T) {
		cases := []struct {
			name   string
			params sqlc.CreateBookParams
		}{
			{
				name: "status fuera de la lista permitida",
				params: sqlc.CreateBookParams{
					Title: title, Author: author, Status: "terminado",
				},
			},
			{
				name: "rating fuera del rango 1 a 5",
				params: sqlc.CreateBookParams{
					Title: title, Author: author, Status: "leido",
					Rating: sql.NullInt32{Int32: 9, Valid: true},
				},
			},
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				if _, err := queries.CreateBook(ctx, c.params); err == nil {
					t.Errorf("CreateBook(ctx, params) con %s no devolvió error; la base debía rechazarlo", c.name)
				}
			})
		}
	})
	// 2. Leer
	t.Run("GetBook", func(t *testing.T) {
		book, err := queries.GetBook(ctx, created.ID)
		if err != nil {
			t.Fatalf("GetBook(ctx, %d) devolvió error: %s", created.ID, err)
		}
		if book.Title != title || book.Author != author || book.Status != "no_leido" {
			t.Errorf("GetBook(ctx, %d) = (%s, %s, %s); se esperaba (%s, %s, no_leido)",
				created.ID, book.Title, book.Author, book.Status, title, author)
		}
	})

	// 3. Actualizar y volver a leer para confirmar que el cambio se aplicó
	t.Run("UpdateBook", func(t *testing.T) {
		const readDate = "2026-09-10"
		date, err := time.Parse("2006-01-02", readDate)
		if err != nil {
			t.Fatalf("time.Parse(%s) devolvió error: %s", readDate, err)
		}

		err = queries.UpdateBook(ctx, sqlc.UpdateBookParams{
			ID:       created.ID,
			Status:   "leido",
			ReadDate: sql.NullTime{Time: date, Valid: true},
			Rating:   sql.NullInt32{Int32: 5, Valid: true},
			Notes:    sql.NullString{String: "Muy bueno", Valid: true},
		})
		if err != nil {
			t.Fatalf("UpdateBook(ctx, params) devolvió error: %s", err)
		}

		book, err := queries.GetBook(ctx, created.ID)
		if err != nil {
			t.Fatalf("GetBook(ctx, %d) devolvió error tras actualizar: %s", created.ID, err)
		}
		// La columna es DATE, pero Go la recibe como time.Time con hora y zona:
		// se normaliza a UTC y se compara solo la parte de fecha.
		storedDate := book.ReadDate.Time.UTC().Format("2006-01-02")
		if book.Status != "leido" || book.Rating.Int32 != 5 || storedDate != readDate {
			t.Errorf("GetBook(ctx, %d) tras actualizar = (%s, %d, %s); se esperaba (leido, 5, %s)",
				created.ID, book.Status, book.Rating.Int32, storedDate, readDate)
		}
	})

	// 4. Listar y comprobar que está el libro ya actualizado
	t.Run("ListBooks", func(t *testing.T) {
		books, err := queries.ListBooks(ctx)
		if err != nil {
			t.Fatalf("ListBooks(ctx) devolvió error: %s", err)
		}
		found := false
		for _, b := range books {
			if b.ID == created.ID && b.Status == "leido" {
				found = true
			}
		}
		if !found {
			t.Errorf("ListBooks(ctx) no contiene el libro actualizado con ID %d", created.ID)
		}
	})

	// 5. Borrar y confirmar que ya no se puede obtener
	t.Run("DeleteBook", func(t *testing.T) {
		if err := queries.DeleteBook(ctx, created.ID); err != nil {
			t.Fatalf("DeleteBook(ctx, %d) devolvió error: %s", created.ID, err)
		}
		if _, err := queries.GetBook(ctx, created.ID); !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("GetBook(ctx, %d) tras borrar devolvió %v; se esperaba %v",
				created.ID, err, sql.ErrNoRows)
		}
	})
}
