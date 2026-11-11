package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib" // registra el driver "pgx" para database/sql
)

// Open se conecta a la base de DATABASE_URL y comprueba que responda
func Open() (*sql.DB, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, errors.New("la variable DATABASE_URL no está definida (correr con `make run-app`)")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("abrir la base: %w", err)
	}
	// sql.Open no se conecta: el Ping es el que confirma que la base responde
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("conectar a la base (¿está levantada? `make run-app` la levanta): %w", err)
	}
	return db, nil
}
