# las credenciales estan en .env, el mismo archivo que lee Docker Compose
include .env

# si falta el .env, make lo crea desde .env.example; sin prerrequisitos para no pisar uno que ya existe
.env:
	@cp .env.example .env
	@echo "-> Se creó .env a partir de .env.example"

# la aplicacion corre en la maquina y llega a la base del contenedor por el puerto publicado
DATABASE_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/$(POSTGRES_DB)?sslmode=disable

# la fecha de lectura es el dia de hoy en Argentina y no en UTC
APP_TZ := America/Argentina/Buenos_Aires

BINARY := bin/booklog

# version del generador de templ: se lee de go.mod para que sea la misma que la de la biblioteca
TEMPL_VERSION = $(shell go list -m -f '{{.Version}}' github.com/a-h/templ)

.PHONY: run-app test-app pre-tasks post-tasks test generate build down clean

# tareas previas, la aplicacion y, al cortarla con Ctrl+C, las tareas posteriores
# el trap de INT hace falta para que corra el de EXIT al cortar con Ctrl+C
run-app: pre-tasks
	@echo "=== APLICACION ==="
	@echo "-> 5. Aplicación en http://localhost:8080 (Ctrl+C para terminar)"
	@trap 'trap "" INT; $(MAKE) --no-print-directory post-tasks' EXIT; trap 'exit 0' INT TERM HUP; \
		DATABASE_URL="$(DATABASE_URL)" TZ=$(APP_TZ) ./$(BINARY)

# ciclo de prueba: tareas previas, la aplicacion en segundo plano, Hurl y las tareas posteriores
test-app: pre-tasks
	@echo "=== PRUEBA DE LAS RUTAS ==="
	@DATABASE_URL="$(DATABASE_URL)" TZ=$(APP_TZ) ./$(BINARY) & app=$$!; \
		trap 'trap "" INT; kill $$app 2>/dev/null; wait $$app 2>/dev/null; $(MAKE) --no-print-directory post-tasks' EXIT; \
		trap 'exit 130' INT TERM HUP; \
		echo "-> 5. Corriendo requests.hurl contra la aplicación..."; \
		docker compose run --rm -T hurl

# 1) tareas previas
pre-tasks:
	@echo "=== TAREAS PREVIAS ==="
	@echo "-> 1. Generando código (sqlc y templ)..."
	@$(MAKE) --no-print-directory generate
	@echo "-> 2. Compilando la aplicación..."
	@go build -o $(BINARY) ./cmd/server
	@echo "-> 3. Borrando contenedores y volúmenes previos..."
	@$(MAKE) --no-print-directory down
	@echo "-> 4. Levantando la base y esperando a que esté lista..."
	@docker compose up -d --wait db || ( echo "-> La base no llegó a estar lista. Limpiando..." ; docker compose logs db ; docker compose down -v ; exit 1 )

# 3) tareas posteriores
# en otra sesion, para que otro Ctrl+C no corte la limpieza
post-tasks:
	@echo "=== TAREAS POSTERIORES ==="
	@echo "-> 6. Borrando contenedores y volúmenes..."
	@setsid -w $(MAKE) --no-print-directory down

# tests de logic con un repositorio falso: no necesitan la base ni Docker
test: generate
	@go test ./...

# genera db/sqlc desde db/schema y db/queries, y los _templ.go desde views/*.templ
generate:
	@sqlc generate
	@go run github.com/a-h/templ/cmd/templ@$(TEMPL_VERSION) generate

build: generate
	@go build -o $(BINARY) ./cmd/server

# borra contenedores y volumenes (se pierden los libros cargados)
down:
	@docker compose down -v --remove-orphans

# borra el entorno, el codigo generado y el binario
clean: down
	@rm -rf db/sqlc bin views/*_templ.go
