# Idílica — Backend Go (espejo de idilica-backend)

Réplica 1:1 del backend Node (`idilica-backend`) escrita en **Go + Gin + GORM**,
con la misma estructura de carpetas, el mismo contrato de API y las mismas
variables de entorno. El objetivo es doble: servir la app y ser tu campo de
práctica de Go comparando archivo por archivo contra el Node que ya dominas.

## El mapa Node → Go

| Node (Express) | Go (Gin) | Nota |
| --- | --- | --- |
| `src/index.ts` | `cmd/api/main.go` | + graceful shutdown explícito |
| `src/config/config.ts` (Joi env) | `internal/config/config.go` | mismos nombres de env vars |
| `src/config/express.ts` | `internal/routes/router.go` | middlewares + wiring |
| `src/config/sequelize.ts` | `internal/config/database.go` | `AutoMigrate` ≈ `sequelize.sync()` |
| winston | `log/slog` (stdlib) | texto en dev, JSON en prod |
| `src/errors/*` | `internal/apperrors/*` | mismo JSON `{error:{code,message,status}}` |
| `next(error)` | `c.Error(err)` + `ErrorHandler` | patrón idéntico |
| Joi schemas (`src/validations`) | tags `binding` en los DTOs | validator de Gin |
| `validateDto<LoginDto>(...)` | `utils.BindJSON[dto.LoginDto](c)` | generics de Go |
| Sequelize models | GORM models (`internal/models`) | mismas tablas/columnas |
| `withTransaction` (AsyncLocalStorage) | `db.Transaction(func(tx))` | explícito > implícito |
| express-rate-limit | `x/time/rate` por IP (a mano) | lección de concurrencia |
| singletons por `import` | inyección explícita en `router.go` | el grafo de deps se lee completo |
| `tsx watch` | `air` (`make dev`) | hot reload |
| jest | `go test ./...` | stdlib `testing` |
| eslint | `go vet` / golangci-lint | `make lint` |

## Paridad real con el backend Node

- **Mismo contrato de API**: rutas, formas JSON (camelCase), códigos y formato
  de error idénticos. La app React funciona contra cualquiera de los dos.
- **Mismo `JWT_SECRET` ⇒ tokens intercambiables**, y mismas llaves de Redis
  (`refresh_token:*`) ⇒ las sesiones sobreviven al cambiar de backend.
- **Misma forma de base de datos** (tablas `user`, `cocina`, `cocina_member`,
  UUIDs, snake_case), pero en una base PROPIA (`idilica_go`) para que ambos
  corran a la vez sin pisarse.
- Puertos: Node `4050`, **Go `4051`**.

## Puesta en marcha

```bash
cp .env.example .env      # ajusta SQL_PASSWORD y JWT_SECRET
make docker-up            # postgres + redis propios (crea la BD idilica_go sola)
make dev                  # air si está instalado; si no: go run ./cmd/api
go run ./cmd/seed -email <usuario>   # opcional: dataset demo del diseño
```

`GET http://localhost:4051/api/health-check` → `OK`.

Si ya tienes OTRO Postgres/Redis ocupando los puertos (p. ej. el stack de
maguey), puedes reutilizarlo (crea la BD con `createdb idilica_go`) o levantar
este compose en puertos alternos: `SQL_PORT=5433 REDIS_PORT=6380 make docker-up`
(y ajusta el `.env` igual). También puedes correr todo contenedorizado:
`docker compose --profile app up -d --build`.

Para probar la app React contra este backend, cambia el target del proxy en
`idilica-app/vite.config.ts` de `:4050` a `:4051`.

## Producción (droplet)

Go compila a **un binario estático**: no hay `node_modules`, ni runtime, ni
`npm install` en el servidor. El deploy es compilar, copiar un archivo y
reiniciar el servicio:

```bash
make build-linux                                  # bin/api-linux (Linux x86_64)
scp bin/api-linux ellebkey@droplet:/home/ellebkey/apps/idilica/api
scp .env.prod ellebkey@droplet:/home/ellebkey/apps/idilica/.env
sudo systemctl restart idilica-api
```

### ¿Y el pm2 de Go?

No necesitas uno: **systemd** (ya incluido en Linux) hace todo lo que pm2 hace
por Node — reiniciar si truena, arrancar en cada boot, logs centralizados. La
unidad lista está en `deploy/idilica-api.service` (instrucciones adentro).

| pm2 (Node) | systemd (Go) |
| --- | --- |
| `pm2 start` | `sudo systemctl enable --now idilica-api` |
| `pm2 status` | `systemctl status idilica-api` |
| `pm2 restart` | `sudo systemctl restart idilica-api` |
| `pm2 logs` | `journalctl -u idilica-api -f` |
| `pm2 startup` (boot) | ya incluido en `enable` |
| modo cluster (multi-core) | innecesario: las goroutines usan todos los cores en un solo proceso |

Si tu droplet ya corre pm2 para maguey y prefieres un solo panel, pm2 también
supervisa binarios: `pm2 start ./api --name idilica-api`. Funciona igual; solo
es una dependencia extra (Node) que el binario Go no necesita.

## Comandos

| npm (Node) | make (Go) |
| --- | --- |
| `npm run dev` | `make dev` |
| `npm run build` | `make build` (binario en `bin/api`) |
| `npm test` | `make test` |
| `npm run lint` | `make lint` |
| — | `make vet`, `make tidy` |

Herramientas opcionales:

```bash
go install github.com/air-verse/air@latest          # hot reload
brew install golangci-lint                           # linter completo
```

## Estructura

```
cmd/api/main.go        arranque + graceful shutdown
internal/
  config/              env, slog, GORM (Postgres), Redis
  apperrors/           AppError + middleware de errores (mismo JSON que Node)
  middlewares/         auth JWT, roles, rate limit, headers, request logger
  models/              User, Cocina, CocinaMember (GORM)
  dto/                 structs de entrada/salida con tags binding+json
  services/            jwt, email (Resend), auth, cocina
  controllers/         handlers delgados (bind → service → JSON)
  routes/              wiring explícito + registro de rutas
  utils/               BindJSON genérico, contexto de usuario
```

## Apuntes de Go para venir de Node

- **Los errores son valores**: no hay `try/catch`; cada llamada devuelve
  `(resultado, error)` y decides. `errors.Is/As` ≈ `instanceof`.
- **`ctx context.Context` viaja primero** en cada función de servicio: es la
  cancelación/deadline del request (Node lo hace implícito).
- **Punteros = opcionalidad**: `*string` en un DTO distingue "no vino" de
  "vino vacío" (lo que en Joi era `.optional()`).
- **Interfaces implícitas**: `Roles` implementa `driver.Valuer`/`sql.Scanner`
  solo por tener los métodos — así se mapea un tipo custom a una columna JSON.
- La migración a producción se hace con binario compilado (`make build`):
  no hay `node_modules` ni runtime que instalar en el servidor.
