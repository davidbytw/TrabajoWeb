package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	// incluye las zonas horarias en el binario, por si el sistema no las tiene
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

// run devuelve el error en vez de llamar a log.Fatal, que terminaria sin ejecutar el defer db.Close()
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

	// las hojas de estilos se sirven desde la carpeta static
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	// "/" recibe todo lo demas: la pagina, los formularios y las rutas que no existen (404)
	http.Handle("/", bookHandler)

	// el middleware envuelve al mux por defecto para que se aplique a todas las rutas
	server := &http.Server{Addr: serverAddress, Handler: middleware.Logging(http.DefaultServeMux)}

	// el contexto se cancela con Ctrl+C (SIGINT) o con SIGTERM
	signalContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// el servidor atiende en otra gorutina; si no puede arrancar, el error llega por el canal
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Servidor escuchando en %s", serverAddress)
		serverErrors <- server.ListenAndServe()
	}()

	// espera lo primero que pase: que el servidor falle o que se corte con Ctrl+C
	select {
	case err := <-serverErrors:
		return err
	case <-signalContext.Done():
	}

	// deja de aceptar pedidos nuevos y espera hasta 10 segundos a que terminen los que estan en curso
	log.Println("Apagando el servidor...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}
