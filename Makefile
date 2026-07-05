# Espejo de los npm scripts del backend Node.
.PHONY: dev run build test vet lint tidy docker-up docker-down build-linux

# npm run dev  (usa air si está instalado; si no, go run sin hot reload)
dev:
	@command -v air >/dev/null 2>&1 && air || go run ./cmd/api

# node release/index.js
run:
	go run ./cmd/api

# npm run build
build:
	go build -o bin/api ./cmd/api

# npm test
test:
	go test ./...

# análisis estático de la stdlib (siempre disponible)
vet:
	go vet ./...

# npm run lint (golangci-lint si está instalado; si no, go vet)
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
