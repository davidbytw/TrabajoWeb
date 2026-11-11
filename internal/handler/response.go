package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/a-h/templ"

	"booklog/logic"
	"booklog/views"
)

// writePage responde una pagina HTML completa con ese codigo de estado
func writePage(w http.ResponseWriter, r *http.Request, status int, page templ.Component) {
	// las cabeceras y el codigo van antes del cuerpo
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := page.Render(r.Context(), w); err != nil {
		// el codigo ya se envio, asi que solo queda registrarlo
		log.Printf("error al escribir la pagina: %s", err)
	}
}

// showError vuelve a mostrar la pagina con el mensaje y la lista actualizada
func (h *BookHandler) showError(w http.ResponseWriter, r *http.Request, status int, message, title, author string) {
	books, err := h.service.List(r.Context())
	if err != nil {
		writeServerError(w, err)
		return
	}
	writePage(w, r, status, views.IndexPage(message, title, author, books))
}

// showServiceError elige el codigo HTTP segun el tipo de error que devolvio el servicio
func (h *BookHandler) showServiceError(w http.ResponseWriter, r *http.Request, err error, title, author string) {
	var validationError *logic.ValidationError
	var conflictError *logic.ConflictError
	switch {
	case errors.As(err, &validationError):
		h.showError(w, r, http.StatusBadRequest, err.Error(), title, author)
	case errors.Is(err, logic.ErrBookNotFound):
		h.showError(w, r, http.StatusNotFound, err.Error(), title, author)
	case errors.As(err, &conflictError):
		h.showError(w, r, http.StatusConflict, err.Error(), title, author)
	default:
		writeServerError(w, err)
	}
}

// writeServerError responde 500 sin mostrar el error interno, que queda en el log
func writeServerError(w http.ResponseWriter, err error) {
	log.Printf("error interno: %s", err)
	http.Error(w, "error interno del servidor", http.StatusInternalServerError)
}

// writeMethodNotAllowed responde 405 con la cabecera Allow, que HTTP exige en ese caso
func writeMethodNotAllowed(w http.ResponseWriter, allowedMethods string) {
	w.Header().Set("Allow", allowedMethods)
	http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
}
