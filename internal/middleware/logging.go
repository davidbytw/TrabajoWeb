package middleware

import (
	"log"
	"net/http"
)

// Logging registra el metodo, la ruta y la direccion remota de cada peticion antes de pasarla al handler
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
