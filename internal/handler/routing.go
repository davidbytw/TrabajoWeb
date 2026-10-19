package handler

import (
	"net/http"
	"strconv"
	"strings"
)

// ServeHTTP atiende /books y /books/{id} y despacha según el método HTTP
func (h *BookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// /books -> ["", "books"], /books/7 -> ["", "books", "7"]
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) > 3 {
		writeError(w, http.StatusNotFound, "ruta no encontrada")
		return
	}

	// sin id (/books o /books/): la colección
	if len(parts) == 2 || parts[2] == "" {
		switch r.Method {
		case http.MethodGet:
			h.listBooks(w, r)
		case http.MethodPost:
			h.createBook(w, r)
		default:
			writeMethodNotAllowed(w, "GET, POST")
		}
		return
	}

	// con id: el tamaño 32 rechaza lo que no entra en el int32 de la columna SERIAL
	id64, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "el id del libro tiene que ser un número entero")
		return
	}
	id := int32(id64)

	switch r.Method {
	case http.MethodGet:
		h.getBook(w, r, id)
	case http.MethodPut:
		h.updateBook(w, r, id)
	case http.MethodDelete:
		h.deleteBook(w, r, id)
	default:
		writeMethodNotAllowed(w, "GET, PUT, DELETE")
	}
}

// writeMethodNotAllowed responde 405 con la cabecera Allow, que HTTP exige en ese caso
func writeMethodNotAllowed(w http.ResponseWriter, allowedMethods string) {
	w.Header().Set("Allow", allowedMethods)
	writeError(w, http.StatusMethodNotAllowed, "método no permitido")
}

// NotFound responde 404 en JSON a las rutas que no son de la API
func NotFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotFound, "ruta no encontrada")
}
