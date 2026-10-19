package logic

import (
	"errors"
	"time"
)

// estados posibles de un libro, los mismos que acepta el CHECK book_status_valid
const (
	StatusNoLeido = "no_leido"
	StatusLeyendo = "leyendo"
	StatusLeido   = "leido"
)

// Book es un libro del registro: los campos opcionales son punteros y nil equivale a NULL
type Book struct {
	ID       int32
	Title    string
	Author   string
	Status   string
	ReadDate *time.Time // el día se guarda como la medianoche UTC de esa fecha
	Rating   *int32
	Notes    *string
}

// ErrBookNotFound indica que no existe un libro con ese id
var ErrBookNotFound = errors.New("libro no encontrado")

// ValidationError es un error de una regla de los datos (el handler responde 400)
type ValidationError struct {
	message string
}

// Error devuelve el mensaje de la regla
func (e *ValidationError) Error() string {
	return e.message
}

// ConflictError es un cambio que choca con el estado actual del libro (el handler responde 409)
type ConflictError struct {
	message string
}

// Error devuelve el mensaje del conflicto
func (e *ConflictError) Error() string {
	return e.message
}
