# Tareas del proyecto.
.PHONY: dev run build test vet lint tidy docker-up docker-down build-linux

# desarrollo con hot reload (air si está instalado; si no, go run)
dev:
	@command -v air >/dev/null 2>&1 && air || go run ./cmd/api

# correr sin hot reload
run:
	go run ./cmd/api

# binario local
build:
	go build -o bin/api ./cmd/api

# pruebas
test:
	go test ./...

# análisis estático de la stdlib (siempre disponible)
vet:
	go vet ./...

# golangci-lint si está instalado; si no, go vet
lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || go vet ./...

# mantiene go.mod/go.sum en orden
tidy:
	go mod tidy

# infraestructura local (postgres + redis)
docker-up:
	docker compose up -d

docker-down:
	docker compose down

# binario de producción para el droplet (Linux x86_64, estático)
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/api-linux ./cmd/api
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/seed-linux ./cmd/seed
