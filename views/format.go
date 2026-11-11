package views

import (
	"fmt"
	"time"

	"booklog/logic"
)

// statusLabel devuelve el estado tal como se muestra en la pagina
func statusLabel(status string) string {
	switch status {
	case logic.StatusNoLeido:
		return "No leído"
	case logic.StatusLeyendo:
		return "Leyendo"
	case logic.StatusLeido:
		return "Leído"
	default:
		return status
	}
}

// formatDate muestra la fecha como dd/mm/aaaa, en UTC como se guarda
func formatDate(date time.Time) string {
	return date.UTC().Format("02/01/2006")
}

// bookURL arma la ruta de una accion sobre un libro, por ejemplo /books/7/status
func bookURL(id int32, action string) string {
	return fmt.Sprintf("/books/%d/%s", id, action)
}
