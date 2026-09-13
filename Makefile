.PHONY: test pre-test run-test post-test up down generate build clean

test: pre-test run-test

# 1) Tareas previas, en el orden que lista la consigna.
pre-test:
	@echo "=== TAREAS PREVIAS ==="
	@echo "-> 1. Generando el código de acceso a datos con sqlc..."
	@$(MAKE) --no-print-directory generate
	@echo "-> 2. Compilando el proyecto..."
	@go build ./...
	@echo "-> 3. Borrando contenedores y volúmenes previos..."
	@$(MAKE) --no-print-directory down
	@echo "-> 4. Levantando la base y esperando a que esté sana..."
	@$(MAKE) --no-print-directory up

# 2) Ejecucion de tests. 
run-test:
	@echo "=== EJECUCIÓN DE TESTS ==="
		@go test -v -count=1 ./... ; status=$$? ; $(MAKE) --no-print-directory post-test ; exit $$status

# 3) Tareas posteriores
post-test:
	@echo "=== TAREAS POSTERIORES ==="
	@echo "-> Borrando contenedores y volúmenes..."
	@$(MAKE) --no-print-directory down

up:
	@docker compose up -d --wait || ( echo "-> La base no llegó a estar sana. Limpiando..." ; docker compose down -v ; exit 1 )

down:
	@docker compose down -v

generate:
	@sqlc generate

build: generate
	@go build ./...

clean: down
	@rm -rf db/sqlc