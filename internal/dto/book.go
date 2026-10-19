package dto

// CreateBookRequest es el cuerpo de POST /books
type CreateBookRequest struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

// UpdateBookRequest es el cuerpo de PUT /books/{id}; puntaje y notas son nil si no vienen
type UpdateBookRequest struct {
	Status string  `json:"status"`
	Rating *int32  `json:"rating"`
	Notes  *string `json:"notes"`
}

// BookResponse es un libro tal como lo devuelve la API
type BookResponse struct {
	ID       int32   `json:"id"`
	Title    string  `json:"title"`
	Author   string  `json:"author"`
	Status   string  `json:"status"`
	ReadDate *string `json:"read_date"` // "AAAA-MM-DD"
	Rating   *int32  `json:"rating"`
	Notes    *string `json:"notes"`
}

// ErrorResponse es el cuerpo de las respuestas de error
type ErrorResponse struct {
	Error string `json:"error"`
}
