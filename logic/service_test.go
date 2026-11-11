package logic

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// fakeBookRepository guarda los libros en un mapa; no lleva candado porque los tests no corren en paralelo
type fakeBookRepository struct {
	books  map[int32]Book
	nextID int32
	// beforeUpdate, si no es nil, simula otra peticion entre la lectura y la escritura del servicio
	beforeUpdate func()
}

func newFakeBookRepository() *fakeBookRepository {
	return &fakeBookRepository{books: map[int32]Book{}, nextID: 1}
}

func (f *fakeBookRepository) Create(ctx context.Context, title, author string) (Book, error) {
	book := Book{ID: f.nextID, Title: title, Author: author, Status: StatusNoLeido}
	f.nextID++
	f.books[book.ID] = book
	return book, nil
}

func (f *fakeBookRepository) Get(ctx context.Context, id int32) (Book, error) {
	book, ok := f.books[id]
	if !ok {
		return Book{}, ErrBookNotFound
	}
	return book, nil
}

func (f *fakeBookRepository) List(ctx context.Context) ([]Book, error) {
	books := make([]Book, 0, len(f.books))
	for _, book := range f.books {
		books = append(books, book)
	}
	return books, nil
}

// UpdateStatus imita el update condicional: solo escribe si el libro existe y sigue en currentStatus
func (f *fakeBookRepository) UpdateStatus(ctx context.Context, book Book, currentStatus string) (Book, error) {
	if f.beforeUpdate != nil {
		f.beforeUpdate()
	}
	stored, ok := f.books[book.ID]
	if !ok || stored.Status != currentStatus {
		return Book{}, ErrBookNotFound
	}
	f.books[book.ID] = book
	return book, nil
}

func (f *fakeBookRepository) Delete(ctx context.Context, id int32) error {
	if _, ok := f.books[id]; !ok {
		return ErrBookNotFound
	}
	delete(f.books, id)
	return nil
}

// ptr arma en una linea los punteros de los campos opcionales (Rating, Notes)
func ptr[T any](v T) *T {
	return &v
}

// fixedNow son las 23:30 del 9/10 en Buenos Aires: en UTC ya es 10, pero la fecha de lectura tiene que ser el 9
var fixedNow = time.Date(2026, 10, 9, 23, 30, 0, 0, time.FixedZone("ART", -3*60*60))

// newTestService usa el repositorio falso y el reloj fijo
func newTestService(repo *fakeBookRepository) *BookService {
	s := NewBookService(repo)
	s.now = func() time.Time { return fixedNow }
	return s
}

// TestCreate prueba las reglas del alta: obligatorios, recortados, largo maximo y estado inicial
func TestCreate(t *testing.T) {
	cases := []struct {
		name       string
		title      string
		author     string
		wantErr    error
		wantTitle  string
		wantAuthor string
	}{
		{name: "datos válidos se guardan recortados", title: "  Rayuela  ", author: " Julio Cortázar\n",
			wantTitle: "Rayuela", wantAuthor: "Julio Cortázar"},
		{name: "título vacío", title: "", author: "Julio Cortázar", wantErr: ErrEmptyTitle},
		{name: "título de solo espacios", title: "   ", author: "Julio Cortázar", wantErr: ErrEmptyTitle},
		{name: "autor vacío", title: "Rayuela", author: "", wantErr: ErrEmptyAuthor},
		{name: "título de 256 caracteres", title: strings.Repeat("a", 256), author: "X", wantErr: ErrTitleTooLong},
		{name: "autor de 256 caracteres", title: "X", author: strings.Repeat("a", 256), wantErr: ErrAuthorTooLong},
		// 255 vocales con tilde son 510 bytes pero 255 caracteres: tiene que ser valido
		{name: "título de 255 caracteres con tilde", title: strings.Repeat("á", 255), author: "X",
			wantTitle: strings.Repeat("á", 255), wantAuthor: "X"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := newFakeBookRepository()
			s := newTestService(repo)
			book, err := s.Create(context.Background(), c.title, c.author)

			if c.wantErr != nil {
				if !errors.Is(err, c.wantErr) {
					t.Fatalf("Create(%q, %q) devolvió error %v; se esperaba %v", c.title, c.author, err, c.wantErr)
				}
				if len(repo.books) != 0 {
					t.Errorf("Create con datos inválidos guardó %d libros; se esperaba 0", len(repo.books))
				}
				return
			}
			if err != nil {
				t.Fatalf("Create(%q, %q) devolvió error: %s", c.title, c.author, err)
			}
			if book.Title != c.wantTitle || book.Author != c.wantAuthor || book.Status != StatusNoLeido {
				t.Errorf("Create(%q, %q) = (%q, %q, %s); se esperaba (%q, %q, %s)",
					c.title, c.author, book.Title, book.Author, book.Status,
					c.wantTitle, c.wantAuthor, StatusNoLeido)
			}
		})
	}
}

// TestUpdate prueba la tabla de transiciones y las reglas de puntaje, notas y fecha
func TestUpdate(t *testing.T) {
	cases := []struct {
		name    string
		from    string
		to      string
		rating  *int32
		notes   *string
		wantErr error
	}{
		// transiciones permitidas
		{name: "no_leido a leyendo", from: StatusNoLeido, to: StatusLeyendo},
		{name: "no_leido a leido", from: StatusNoLeido, to: StatusLeido, rating: ptr(int32(4))},
		{name: "leyendo a leido con notas", from: StatusLeyendo, to: StatusLeido,
			rating: ptr(int32(5)), notes: ptr("Muy bueno")},
		{name: "relectura leido a leido", from: StatusLeido, to: StatusLeido, rating: ptr(int32(3))},

		// transiciones prohibidas
		{name: "no_leido a no_leido", from: StatusNoLeido, to: StatusNoLeido, wantErr: ErrInvalidTransition},
		{name: "leyendo a leyendo", from: StatusLeyendo, to: StatusLeyendo, wantErr: ErrInvalidTransition},
		{name: "leyendo a no_leido", from: StatusLeyendo, to: StatusNoLeido, wantErr: ErrInvalidTransition},
		{name: "leido a leyendo", from: StatusLeido, to: StatusLeyendo, wantErr: ErrInvalidTransition},
		{name: "leido a no_leido", from: StatusLeido, to: StatusNoLeido, wantErr: ErrInvalidTransition},

		// reglas de los datos
		{name: "estado desconocido", from: StatusNoLeido, to: "terminado", wantErr: ErrUnknownStatus},
		{name: "estado vacío", from: StatusNoLeido, to: "", wantErr: ErrUnknownStatus},
		{name: "leido sin puntaje", from: StatusLeyendo, to: StatusLeido, wantErr: ErrMissingRating},
		{name: "puntaje 0", from: StatusLeyendo, to: StatusLeido, rating: ptr(int32(0)), wantErr: ErrRatingOutOfRange},
		{name: "puntaje 6", from: StatusLeyendo, to: StatusLeido, rating: ptr(int32(6)), wantErr: ErrRatingOutOfRange},
		{name: "puntaje fuera de leido", from: StatusNoLeido, to: StatusLeyendo, rating: ptr(int32(4)), wantErr: ErrFieldsNotAllowed},
		{name: "notas fuera de leido", from: StatusNoLeido, to: StatusLeyendo, notes: ptr("x"), wantErr: ErrFieldsNotAllowed},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := newFakeBookRepository()
			s := newTestService(repo)

			// si parte de leido, el libro tiene que cumplir el invariante: fecha y puntaje
			existing := Book{ID: 1, Title: "Rayuela", Author: "Julio Cortázar", Status: c.from}
			if c.from == StatusLeido {
				oldDate := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
				existing.ReadDate = &oldDate
				existing.Rating = ptr(int32(2))
			}
			repo.books[1] = existing

			book, err := s.Update(context.Background(), 1, c.to, c.rating, c.notes)

			if c.wantErr != nil {
				if !errors.Is(err, c.wantErr) {
					t.Fatalf("Update(%s -> %s) devolvió error %v; se esperaba %v", c.from, c.to, err, c.wantErr)
				}
				if repo.books[1].Status != c.from {
					t.Errorf("tras el rechazo el estado guardado es %s; se esperaba %s", repo.books[1].Status, c.from)
				}
				return
			}
			if err != nil {
				t.Fatalf("Update(%s -> %s) devolvió error: %s", c.from, c.to, err)
			}
			if book.Status != c.to || repo.books[1].Status != c.to {
				t.Errorf("Update(%s -> %s) dejó el estado en %s; se esperaba %s", c.from, c.to, book.Status, c.to)
			}

			// la fecha es el dia de fixedNow en Buenos Aires (9), no en UTC (10)
			if c.to == StatusLeido {
				wantDate := time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)
				if book.ReadDate == nil || !book.ReadDate.Equal(wantDate) || book.Rating == nil || *book.Rating != *c.rating {
					t.Errorf("Update(%s -> leido) = fecha %v, puntaje %v; se esperaba fecha %s y puntaje %d",
						c.from, book.ReadDate, book.Rating, wantDate.Format("2006-01-02"), *c.rating)
				}
			} else if book.ReadDate != nil || book.Rating != nil || book.Notes != nil {
				t.Errorf("Update(%s -> %s) dejó fecha, puntaje o notas cargados", c.from, c.to)
			}
		})
	}
}

// TestUpdate_ErrorKinds prueba que cada error sea del tipo que el handler traduce a 400 o 409
func TestUpdate_ErrorKinds(t *testing.T) {
	var validationError *ValidationError
	if err := validateStatusFields("terminado", nil, nil); !errors.As(err, &validationError) {
		t.Errorf("un estado desconocido devolvió %v; se esperaba un *ValidationError", err)
	}

	repo := newFakeBookRepository()
	repo.books[1] = Book{ID: 1, Title: "Rayuela", Author: "Julio Cortázar", Status: StatusLeyendo}
	_, err := newTestService(repo).Update(context.Background(), 1, StatusNoLeido, nil, nil)
	var conflictError *ConflictError
	if !errors.As(err, &conflictError) {
		t.Errorf("una transición prohibida devolvió %v; se esperaba un *ConflictError", err)
	}
}

// TestUpdate_NotFound prueba que un id inexistente devuelva ErrBookNotFound
func TestUpdate_NotFound(t *testing.T) {
	s := newTestService(newFakeBookRepository())
	_, err := s.Update(context.Background(), 99, StatusLeyendo, nil, nil)
	if !errors.Is(err, ErrBookNotFound) {
		t.Errorf("Update(99) devolvió %v; se esperaba %v", err, ErrBookNotFound)
	}
}

// TestUpdate_ConcurrentChange simula otro pedido que pasa el libro a leido entre la lectura y la escritura
func TestUpdate_ConcurrentChange(t *testing.T) {
	repo := newFakeBookRepository()
	repo.books[1] = Book{ID: 1, Title: "Rayuela", Author: "Julio Cortázar", Status: StatusNoLeido}
	finished := Book{ID: 1, Title: "Rayuela", Author: "Julio Cortázar", Status: StatusLeido,
		ReadDate: ptr(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)), Rating: ptr(int32(5))}
	repo.beforeUpdate = func() { repo.books[1] = finished }

	_, err := newTestService(repo).Update(context.Background(), 1, StatusLeyendo, nil, nil)

	if !errors.Is(err, ErrConcurrentUpdate) {
		t.Fatalf("Update devolvió %v; se esperaba %v", err, ErrConcurrentUpdate)
	}
	if repo.books[1].Status != StatusLeido {
		t.Errorf("el estado guardado quedó en %s; se esperaba %s (el de la otra petición)", repo.books[1].Status, StatusLeido)
	}
}

// TestUpdate_DeletedMeanwhile simula otra peticion que borra el libro en el medio: tiene que ser 404, no 409
func TestUpdate_DeletedMeanwhile(t *testing.T) {
	repo := newFakeBookRepository()
	repo.books[1] = Book{ID: 1, Title: "Rayuela", Author: "Julio Cortázar", Status: StatusNoLeido}
	repo.beforeUpdate = func() { delete(repo.books, 1) }

	_, err := newTestService(repo).Update(context.Background(), 1, StatusLeyendo, nil, nil)

	if !errors.Is(err, ErrBookNotFound) {
		t.Errorf("Update devolvió %v; se esperaba %v", err, ErrBookNotFound)
	}
}

// TestUpdate_NotesWithOnlySpaces prueba que las notas de solo espacios se guarden como nil y no como ""
func TestUpdate_NotesWithOnlySpaces(t *testing.T) {
	repo := newFakeBookRepository()
	repo.books[1] = Book{ID: 1, Title: "Rayuela", Author: "Julio Cortázar", Status: StatusLeyendo}

	book, err := newTestService(repo).Update(context.Background(), 1, StatusLeido, ptr(int32(5)), ptr("   "))
	if err != nil {
		t.Fatalf("Update devolvió error: %s", err)
	}
	if book.Notes != nil {
		t.Errorf("Update con notas %q guardó %q; se esperaba nil", "   ", *book.Notes)
	}
}

// TestDelete prueba borrar un libro y despues volver a borrar el mismo id, que ya no existe
func TestDelete(t *testing.T) {
	repo := newFakeBookRepository()
	repo.books[1] = Book{ID: 1, Title: "Rayuela", Author: "Julio Cortázar", Status: StatusNoLeido}
	s := newTestService(repo)

	if err := s.Delete(context.Background(), 1); err != nil {
		t.Fatalf("Delete(1) devolvió error: %s", err)
	}
	if err := s.Delete(context.Background(), 1); !errors.Is(err, ErrBookNotFound) {
		t.Errorf("Delete(1) por segunda vez devolvió %v; se esperaba %v", err, ErrBookNotFound)
	}
}
