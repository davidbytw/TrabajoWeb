package middleware

import (
	"log"
	"net/http"
)

// Logging registra el método, la ruta y la dirección remota de cada petición antes de pasarla al handler
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
