package main

import (
	"log"
	"net/http"
	// la imagen alpine no trae las zonas horarias: sin esto la variable TZ no tendría efecto
	_ "time/tzdata"

	"booklog/db/sqlc"
	"booklog/internal/database"
	"booklog/internal/handler"
	"booklog/internal/middleware"
	"booklog/internal/repository"
	"booklog/logic"
)

const serverAddress = ":8080"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run devuelve el error en vez de llamar a log.Fatal, que terminaría sin ejecutar el defer db.Close()
func run() error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	queries := sqlc.New(db)

	bookRepository := repository.NewPostgresBookRepository(queries)
	bookService := logic.NewBookService(bookRepository)
	bookHandler := handler.NewBookHandler(bookService)

	// "/books" es la colección y "/books/" todo lo que sigue a la barra; el resto responde 404 en JSON
	http.Handle("/books", bookHandler)
	http.Handle("/books/", bookHandler)
	http.HandleFunc("/", handler.NotFound)

	log.Printf("Servidor escuchando en %s", serverAddress)
	// el middleware envuelve al mux por defecto para que se aplique a todas las rutas
	return http.ListenAndServe(serverAddress, middleware.Logging(http.DefaultServeMux))
}
