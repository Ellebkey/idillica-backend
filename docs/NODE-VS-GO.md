# Node → Go: mi mapa personal

> Notas de aprendizaje. Este backend nació como espejo 1:1 de mi stack Node
> (Express 5 + Sequelize, el patrón de maguey-backend) precisamente para
> aprender Go sobre terreno conocido. El README del repo es genérico; la
> comparación vive aquí.

## El mapa archivo por archivo

| Node (Express) | Go (Gin) | Nota |
| --- | --- | --- |
| `src/index.ts` | `cmd/api/main.go` | + graceful shutdown explícito |
| `src/config/config.ts` (Joi env) | `internal/config/config.go` | validación a mano, falla rápido |
| `src/config/express.ts` | `internal/routes/router.go` | middlewares + wiring |
| `src/config/sequelize.ts` | `internal/config/database.go` | `AutoMigrate` ≈ `sequelize.sync()` |
| winston | `log/slog` (stdlib) | texto en dev, JSON en prod |
| `src/errors/*` | `internal/apperrors/*` | mismo JSON `{error:{code,message,status}}` |
| `next(error)` | `c.Error(err)` + `ErrorHandler` | patrón idéntico |
| Joi schemas (`src/validations`) | tags `binding` en los DTOs | validator de Gin |
| `validateDto<LoginDto>(...)` | `utils.BindJSON[dto.LoginDto](c)` | generics de Go |
| Sequelize models | GORM models (`internal/models`) | mismas tablas/columnas |
| `withTransaction` (AsyncLocalStorage) | `db.Transaction(func(tx))` | explícito > implícito |
| express-rate-limit | `x/time/rate` por IP (a mano) | mutex + goroutine de limpieza |
| singletons por `import` | inyección explícita en `router.go` | el grafo de deps se lee completo |
| `tsx watch` | `air` (`make dev`) | hot reload |
| jest | `go test ./...` | stdlib `testing` |
| eslint | `go vet` / golangci-lint | `make lint` |
| pm2 | systemd | reinicio, boot, logs — sin proceso extra |
| modo cluster de pm2 | innecesario | las goroutines usan todos los cores |
| `npm run deploy` + node_modules al server | `make build-linux` + scp de UN binario | sin runtime en el servidor |

## Equivalencias que me costaron / me gustaron

- **Los errores son valores**: no hay `try/catch`; cada llamada devuelve
  `(resultado, error)` y tú decides. `errors.Is/As` ≈ `instanceof`.
- **`ctx context.Context` primero** en cada función de servicio: es la
  cancelación/deadline del request que Node hace implícita.
- **Punteros = opcionalidad**: `*string` en un DTO distingue "no vino" de
  "vino vacío" (lo que en Joi era `.optional()`).
- **Interfaces implícitas**: `Roles` implementa `driver.Valuer`/`sql.Scanner`
  solo por tener los métodos — así se mapea un tipo custom a columna JSON.
  Igual: `gin.Engine` implementa `http.Handler`, por eso puede vivir dentro de
  un `http.Server` de la stdlib (y por eso el graceful shutdown es trivial).
- **La stdlib cubre lo que en Node eran paquetes**: slog (winston),
  crypto (uuid/random/sha), testing (jest), http.Server. La cultura Go es
  anti-dependencias: 40 líneas propias antes que un paquete.
- El **rate limiter a mano** (`internal/middlewares/ratelimit.go`) fue mi
  primer contacto real con concurrencia: mapa + `sync.Mutex` + goroutine con
  `time.Ticker` para limpiar IPs viejas.

## Paridad histórica con el backend Node original

El scaffold Node (`idilica-backend`, hoy congelado) y este backend compartían
contrato al 100%: mismas rutas y formas JSON, mismo esquema de llaves en Redis
(`refresh_token:<sha256>`, `refresh_tokens_user:<id>`) y mismo payload JWT —
con el mismo `JWT_SECRET`, los tokens y sesiones eran intercambiables entre
ambos. El frontend cambiaba de backend con solo mover el proxy de Vite.
Desde el 2026-07-05 el Go es el único backend del proyecto (`APP_ENV`
reemplazó a `NODE_ENV` ese mismo día).
