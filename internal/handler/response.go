package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"booklog/internal/dto"
	"booklog/logic"
)

// decodeJSON decodifica el cuerpo en dst y rechaza las claves que no son del contrato en vez de ignorarlas
func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("el cuerpo no es el JSON esperado: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("error al codificar la respuesta: %s", err)
	}
}

// writeError responde {"error": message} con ese código de estado
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, dto.ErrorResponse{Error: message})
}

// writeServiceError elige el código HTTP según el tipo de error que devolvió el servicio
func writeServiceError(w http.ResponseWriter, err error) {
	var validationError *logic.ValidationError
	var conflictError *logic.ConflictError
	switch {
	case errors.As(err, &validationError):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, logic.ErrBookNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.As(err, &conflictError):
		writeError(w, http.StatusConflict, err.Error())
	default:
		// el detalle queda en el log; al cliente no se le muestran errores internos
		log.Printf("error interno: %s", err)
		writeError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
