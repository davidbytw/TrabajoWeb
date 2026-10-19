package logic

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// BookRepository es lo que el servicio necesita de la persistencia
type BookRepository interface {
	Create(ctx context.Context, title, author string) (Book, error)
	Get(ctx context.Context, id int32) (Book, error)
	List(ctx context.Context) ([]Book, error)
	// UpdateStatus guarda solo si el libro sigue en currentStatus; si no, devuelve ErrBookNotFound
	UpdateStatus(ctx context.Context, book Book, currentStatus string) (Book, error)
	Delete(ctx context.Context, id int32) error
}

// BookService aplica las reglas del negocio sobre los libros
type BookService struct {
	books BookRepository
	now   func() time.Time // reloj reemplazable en los tests
}

// NewBookService crea el servicio con el repositorio que va a usar
func NewBookService(books BookRepository) *BookService {
	return &BookService{books: books, now: time.Now}
}

// Create valida título y autor y registra el libro en estado no_leido
func (s *BookService) Create(ctx context.Context, title, author string) (Book, error) {
	title, err := normalizeText(title, ErrEmptyTitle, ErrTitleTooLong)
	if err != nil {
		return Book{}, err
	}
	author, err = normalizeText(author, ErrEmptyAuthor, ErrAuthorTooLong)
	if err != nil {
		return Book{}, err
	}
	return s.books.Create(ctx, title, author)
}

// Get devuelve el libro con ese id, o ErrBookNotFound
func (s *BookService) Get(ctx context.Context, id int32) (Book, error) {
	return s.books.Get(ctx, id)
}

// List devuelve todos los libros ordenados por título
func (s *BookService) List(ctx context.Context) ([]Book, error) {
	return s.books.List(ctx)
}

// Update cambia el estado del libro y, si pasa a leido, carga fecha, puntaje y notas
func (s *BookService) Update(ctx context.Context, id int32, status string, rating *int32, notes *string) (Book, error) {
	if err := validateStatusFields(status, rating, notes); err != nil {
		return Book{}, err
	}

	book, err := s.books.Get(ctx, id)
	if err != nil {
		return Book{}, err
	}

	if !canTransition(book.Status, status) {
		return Book{}, fmt.Errorf("%w: de %s a %s", ErrInvalidTransition, book.Status, status)
	}

	currentStatus := book.Status
	book.Status = status
	book.ReadDate, book.Rating, book.Notes = nil, nil, nil
	if status == StatusLeido {
		readDate := calendarDay(s.now())
		book.ReadDate = &readDate
		book.Rating = rating
		book.Notes = normalizeNotes(notes)
	}

	// si otra petición cambió el estado desde la lectura, el update condicional no se aplica
	updated, err := s.books.UpdateStatus(ctx, book, currentStatus)
	if errors.Is(err, ErrBookNotFound) {
		// se relee para elegir el error: existe (409), se borró (404) o falló la base (500)
		_, getErr := s.books.Get(ctx, id)
		if getErr == nil {
			return Book{}, ErrConcurrentUpdate
		}
		return Book{}, getErr
	}
	return updated, err
}

// Delete elimina el libro, o devuelve ErrBookNotFound si no existe
func (s *BookService) Delete(ctx context.Context, id int32) error {
	return s.books.Delete(ctx, id)
}
