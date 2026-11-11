package handler

import (
	"net/http"
	"strconv"
	"strings"
)

// ServeHTTP reparte los pedidos segun la ruta y el metodo
func (h *BookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		if r.Method != http.MethodGet {
			writeMethodNotAllowed(w, "GET")
			return
		}
		h.index(w, r)
		return
	}

	// /books -> ["", "books"], /books/7/status -> ["", "books", "7", "status"]
	parts := strings.Split(r.URL.Path, "/")
	if parts[1] != "books" {
		http.NotFound(w, r)
		return
	}

	// /books: el formulario de alta
	if len(parts) == 2 {
		if r.Method != http.MethodPost {
			writeMethodNotAllowed(w, "POST")
			return
		}
		h.createBook(w, r)
		return
	}

	// /books/{id}/status y /books/{id}/delete: las acciones sobre un libro
	if len(parts) != 4 || (parts[3] != "status" && parts[3] != "delete") {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, "POST")
		return
	}

	// con 32 se rechaza lo que no entra en el int32 de la columna SERIAL
	id64, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil {
		h.showError(w, r, http.StatusBadRequest, "el id del libro no es válido", "", "")
		return
	}
	id := int32(id64)

	if parts[3] == "status" {
		h.updateStatus(w, r, id)
	} else {
		h.deleteBook(w, r, id)
	}
}
