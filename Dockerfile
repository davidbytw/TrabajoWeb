# etapa 1: compilación (db/sqlc no está en el repo: lo genera make antes del build)
FROM golang:1.26.5-alpine AS builder
WORKDIR /app

# las dependencias van primero para que Docker reutilice esta capa mientras go.mod y go.sum no cambien
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# binario estático, sin librerías del sistema, para que corra en la imagen mínima
RUN CGO_ENABLED=0 go build -o /app/api ./cmd/api

# etapa 2: imagen final, solo con el binario
FROM alpine:3.22
# usuario sin privilegios para no correr el servidor como root
RUN adduser -D -u 1000 app
USER app
COPY --from=builder /app/api /usr/local/bin/api
EXPOSE 8080
CMD ["api"]
