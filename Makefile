# las credenciales están en .env, el mismo archivo que lee Docker Compose
include .env

# si falta el .env, make lo crea desde .env.example; sin prerrequisitos para no pisar uno que ya existe
.env:
	@cp .env.example .env
	@echo "-> Se creó .env a partir de .env.example"

# para la API que corre en la máquina (`make run`); la del contenedor se arma en docker-compose.yml
DATABASE_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/$(POSTGRES_DB)?sslmode=disable

.PHONY: test-api pre-test run-tests post-test test generate build up down run requests logs clean

# ciclo completo para la corrección: si fallan las pruebas, las tareas posteriores corren igual y make termina con error
test-api: pre-test
	@$(MAKE) --no-print-directory run-tests ; status=$$? ; $(MAKE) --no-print-directory post-test ; exit $$status

# 1) tareas previas
pre-test:
	@echo "=== TAREAS PREVIAS ==="
	@echo "-> 1. Generando el código de acceso a datos con sqlc..."
	@$(MAKE) --no-print-directory generate
	@echo "-> 2. Compilando el proyecto..."
	@go build ./...
	@echo "-> 3. Borrando contenedores y volúmenes previos..."
	@$(MAKE) --no-print-directory down
	@echo "-> 4. Levantando la base y la API, y esperando a que estén sanas..."
	@docker compose up -d --build --wait || ( echo "-> No llegaron a estar sanas. Limpiando..." ; docker compose logs api ; docker compose down -v ; exit 1 )

# 2) prueba de cada endpoint con Hurl
run-tests:
	@echo "=== PRUEBA DE LOS ENDPOINTS ==="
	@echo "-> 5. Corriendo requests.hurl..."
	@$(MAKE) --no-print-directory requests

# 3) tareas posteriores
post-test:
	@echo "=== TAREAS POSTERIORES ==="
	@echo "-> 6. Borrando contenedores y volúmenes..."
	@$(MAKE) --no-print-directory down

# tests de logic con un repositorio falso: no necesitan la base ni Docker
test: generate
	@go test ./...

# genera db/sqlc a partir de db/schema y db/queries
generate:
	@sqlc generate

build: generate
	@go build ./...

# levanta la base y la API en contenedores; la API queda en http://localhost:8080
up: generate
	@docker compose up -d --build --wait

# borra contenedores y volúmenes (se pierden los libros cargados)
down:
	@docker compose down -v --remove-orphans

# solo la base en contenedor y la API con Go en :8080 (no usar junto con `make up`: es el mismo puerto)
run: generate
	@docker compose up -d --wait db
	@DATABASE_URL="$(DATABASE_URL)" go run ./cmd/api

# corre requests.hurl contra la API levantada con `make up`
requests:
	@docker compose run --rm hurl

logs:
	@docker compose logs api

# borra el entorno y el código generado
clean: down
	@rm -rf db/sqlc
