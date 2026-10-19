package logic

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// maxTextLength coincide con el VARCHAR(255) del esquema
const (
	maxTextLength = 255
	minRating     = 1
	maxRating     = 5
)

// reglas de los datos: cada una es un valor distinto para compararla con errors.Is
var (
	ErrEmptyTitle       = &ValidationError{"el título no puede estar vacío"}
	ErrTitleTooLong     = &ValidationError{fmt.Sprintf("el título no puede tener más de %d caracteres", maxTextLength)}
	ErrEmptyAuthor      = &ValidationError{"el autor no puede estar vacío"}
	ErrAuthorTooLong    = &ValidationError{fmt.Sprintf("el autor no puede tener más de %d caracteres", maxTextLength)}
	ErrUnknownStatus    = &ValidationError{"el estado tiene que ser no_leido, leyendo o leido"}
	ErrMissingRating    = &ValidationError{"el puntaje es obligatorio al marcar el libro como leído"}
	ErrRatingOutOfRange = &ValidationError{fmt.Sprintf("el puntaje tiene que estar entre %d y %d", minRating, maxRating)}
	ErrFieldsNotAllowed = &ValidationError{"el puntaje y las notas solo se cargan al marcar el libro como leído"}
)

// conflictos con el estado actual del libro
var (
	ErrInvalidTransition = &ConflictError{"cambio de estado no permitido"}
	ErrConcurrentUpdate  = &ConflictError{"otra petición modificó el libro al mismo tiempo; volver a intentar"}
)

// estados a los que se puede pasar desde cada estado
// se escribe solo al arrancar y después solo se lee: es seguro entre gorutinas
var allowedTransitions = map[string][]string{
	StatusNoLeido: {StatusLeyendo, StatusLeido},
	StatusLeyendo: {StatusLeido},
	StatusLeido:   {StatusLeido}, // relectura
}

func isValidStatus(status string) bool {
	_, ok := allowedTransitions[status]
	return ok
}

func canTransition(from, to string) bool {
	for _, allowed := range allowedTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// normalizeText recorta los bordes y valida que el texto no quede vacío ni supere maxTextLength
func normalizeText(value string, errEmpty, errTooLong error) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errEmpty
	}
	// se cuentan caracteres y no bytes, como VARCHAR: "á" ocupa 2 bytes
	if utf8.RuneCountInString(value) > maxTextLength {
		return "", errTooLong
	}
	return value, nil
}

// validateStatusFields aplica las reglas que dependen solo del estado pedido
func validateStatusFields(status string, rating *int32, notes *string) error {
	if !isValidStatus(status) {
		return fmt.Errorf("%w (se recibió %q)", ErrUnknownStatus, status)
	}
	if status != StatusLeido {
		if rating != nil || notes != nil {
			return ErrFieldsNotAllowed
		}
		return nil
	}
	if rating == nil {
		return ErrMissingRating
	}
	if *rating < minRating || *rating > maxRating {
		return ErrRatingOutOfRange
	}
	return nil
}

// normalizeNotes recorta las notas y devuelve nil si quedan vacías, para guardar NULL y no ""
func normalizeNotes(notes *string) *string {
	if notes == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*notes)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// calendarDay devuelve el día de t, en su zona horaria, como la medianoche UTC de ese día
func calendarDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
