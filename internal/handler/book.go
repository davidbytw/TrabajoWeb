package handler

import (
	"errors"
	"net/http"
	"strconv"

	"booklog/logic"
	"booklog/views"
)

// BookHandler atiende la pagina y los formularios de los libros
type BookHandler struct {
	service *logic.BookService
}

// NewBookHandler crea el handler con el servicio que va a usar
func NewBookHandler(service *logic.BookService) *BookHandler {
	return &BookHandler{service: service}
}

// index atiende GET /: la pagina con el formulario de alta y la lista de libros
func (h *BookHandler) index(w http.ResponseWriter, r *http.Request) {
	books, err := h.service.List(r.Context())
	if err != nil {
		writeServerError(w, err)
		return
	}
	writePage(w, r, http.StatusOK, views.IndexPage("", "", "", books))
}

// createBook atiende POST /books: guarda el libro y redirige a la pagina (Post/Redirect/Get)
func (h *BookHandler) createBook(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.showError(w, r, http.StatusBadRequest, "el formulario no es válido", "", "")
		return
	}
	title := r.FormValue("title")
	author := r.FormValue("author")

	if _, err := h.service.Create(r.Context(), title, author); err != nil {
		// el formulario vuelve con lo que se habia escrito
		h.showServiceError(w, r, err, title, author)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// updateStatus atiende POST /books/{id}/status: cambia el estado y redirige a la pagina
func (h *BookHandler) updateStatus(w http.ResponseWriter, r *http.Request, id int32) {
	if err := r.ParseForm(); err != nil {
		h.showError(w, r, http.StatusBadRequest, "el formulario no es válido", "", "")
		return
	}
	rating, err := parseRating(r.FormValue("rating"))
	if err != nil {
		h.showError(w, r, http.StatusBadRequest, err.Error(), "", "")
		return
	}
	notes := optionalText(r.FormValue("notes"))

	if _, err := h.service.Update(r.Context(), id, r.FormValue("status"), rating, notes); err != nil {
		h.showServiceError(w, r, err, "", "")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// deleteBook atiende POST /books/{id}/delete: borra el libro y redirige a la pagina
func (h *BookHandler) deleteBook(w http.ResponseWriter, r *http.Request, id int32) {
	if err := h.service.Delete(r.Context(), id); err != nil {
		h.showServiceError(w, r, err, "", "")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// parseRating convierte el puntaje del formulario; el campo vacio es "sin puntaje" (nil)
func parseRating(value string) (*int32, error) {
	if value == "" {
		return nil, nil
	}
	rating, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return nil, errors.New("el puntaje tiene que ser un número del 1 al 5")
	}
	rating32 := int32(rating)
	return &rating32, nil
}

// optionalText devuelve nil para un campo vacio
func optionalText(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
