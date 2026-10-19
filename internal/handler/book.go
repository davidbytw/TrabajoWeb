package handler

import (
	"net/http"

	"booklog/internal/dto"
	"booklog/logic"
)

// BookHandler atiende los endpoints de /books
type BookHandler struct {
	service *logic.BookService
}

// NewBookHandler crea el handler con el servicio que va a usar
func NewBookHandler(service *logic.BookService) *BookHandler {
	return &BookHandler{service: service}
}

// listBooks atiende GET /books
func (h *BookHandler) listBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.service.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toBookResponses(books))
}

// createBook atiende POST /books y responde 201 con el libro creado
func (h *BookHandler) createBook(w http.ResponseWriter, r *http.Request) {
	var request dto.CreateBookRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	book, err := h.service.Create(r.Context(), request.Title, request.Author)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toBookResponse(book))
}

// getBook atiende GET /books/{id}
func (h *BookHandler) getBook(w http.ResponseWriter, r *http.Request, id int32) {
	book, err := h.service.Get(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toBookResponse(book))
}

// updateBook atiende PUT /books/{id}
func (h *BookHandler) updateBook(w http.ResponseWriter, r *http.Request, id int32) {
	var request dto.UpdateBookRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	book, err := h.service.Update(r.Context(), id, request.Status, request.Rating, request.Notes)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toBookResponse(book))
}

// deleteBook atiende DELETE /books/{id} y responde 204 sin cuerpo
func (h *BookHandler) deleteBook(w http.ResponseWriter, r *http.Request, id int32) {
	if err := h.service.Delete(r.Context(), id); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func toBookResponse(book logic.Book) dto.BookResponse {
	response := dto.BookResponse{
		ID:     book.ID,
		Title:  book.Title,
		Author: book.Author,
		Status: book.Status,
		Rating: book.Rating,
		Notes:  book.Notes,
	}
	if book.ReadDate != nil {
		// se formatea en UTC: en hora argentina la medianoche UTC cae el día anterior
		readDate := book.ReadDate.UTC().Format("2006-01-02")
		response.ReadDate = &readDate
	}
	return response
}

// toBookResponses usa make para que, sin libros, el JSON sea [] y no null
func toBookResponses(books []logic.Book) []dto.BookResponse {
	responses := make([]dto.BookResponse, 0, len(books))
	for _, book := range books {
		responses = append(responses, toBookResponse(book))
	}
	return responses
}
