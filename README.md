# Idílica — API de costeo de recetas

Backend en **Go** para Idílica (panadería gourmet): costeo de recetas en vivo
a partir de un catálogo de ingredientes con precios reales de compra, merma y
gastos de operación. Multi-cocina, con autenticación JWT y roles por workspace.

## Qué hace

- **Ingredientes** con N *productos de compra* (marca, presentación, precio,
  proveedor); solo el producto **activo** alimenta los costos. Cada cambio de
  precio guarda historial y fecha de frescura.
- **Merma** (desperdicio) por ingrediente: de referencia, manual o **medida con
  báscula** (mediciones guardadas). El costo unitario siempre es "por kilo, ya
  con desperdicio": `precio / cantidad / (1 − merma%)`.
- **Recetas** con líneas que apuntan a un ingrediente **o a otra receta**
  (subrecetas anidadas), porciones, rendimiento, precio de venta, alérgenos,
  pasos y fotos. El guardado valida referencias y **rechaza ciclos** en el
  grafo de subrecetas.
- **Gastos de operación** mensuales por cocina (sueldos, gas, luz, equipo)
  repartidos como % sobre las compras de ingredientes.
- Los costos **nunca se persisten**: el frontend carga el catálogo completo en
  un solo `GET` y los deriva en vivo.

## Stack

Go 1.26 · [Gin](https://gin-gonic.com) · [GORM](https://gorm.io) + PostgreSQL ·
Redis (refresh tokens) · `log/slog`.

```
cmd/api          arranque + graceful shutdown
cmd/seed         dataset de demostración
internal/
  config/        env, logger, PostgreSQL (AutoMigrate), Redis
  apperrors/     AppError + middleware de errores (JSON estándar)
  middlewares/   auth JWT, roles, rate limit por IP, headers, request log
  models/        entidades GORM
  dto/           entradas/salidas del API (validación por tags binding)
  services/      lógica de negocio + motor de costos
  controllers/   handlers delgados
  routes/        wiring explícito y registro de endpoints
```

## API

| Método | Ruta | Descripción |
| --- | --- | --- |
| POST | `/api/auth/register` | Crea usuario + su cocina (owner) |
| POST | `/api/auth/login` · `/refresh` · `/logout` | Sesión (JWT + refresh token rotado) |
| GET | `/api/cocinas` · `/api/cocinas/:id` | Cocinas del usuario |
| PUT | `/api/cocinas/:id` | Ajustes (moneda, IVA, objetivo, gastos de operación) |
| GET | `/api/cocinas/:id/catalogo` | **Catálogo completo en un solo GET** |
| POST | `/api/cocinas/:id/ingredientes` | Alta de ingrediente + primer producto |
| PUT/DELETE | `/api/ingredientes/:id` | Editar / eliminar (bloquea si está en uso) |
| POST | `/api/ingredientes/:id/productos` | Agregar producto de compra |
| PUT | `/api/ingredientes/:id/producto-activo/:productoId` | Cambiar el producto EN USO |
| PUT | `/api/ingredientes/:id/merma` | Merma manual/referencia |
| POST | `/api/ingredientes/:id/mediciones` | Medición con báscula (origen "medido") |
| PUT | `/api/productos/:id/precio` | Nuevo precio (+ historial + frescura) |
| POST | `/api/cocinas/:id/recetas` · PUT/DELETE `/api/recetas/:id` | Recetas (replace completo de líneas, detección de ciclos) |

Auth: token JWT crudo en el header `Authorization`. Errores siempre como
`{ "error": { "code", "message", "status", "details"? } }`.

## Desarrollo

Requisitos: Go ≥ 1.26, Docker (o PostgreSQL 16 + Redis 7 propios).

```bash
cp .env.example .env      # ajusta SQL_PASSWORD y JWT_SECRET
make docker-up            # postgres + redis
make dev                  # http://localhost:4051/api (hot reload con air)
go run ./cmd/seed -email <usuario>   # opcional: datos de demostración
```

`GET /api/health-check` → `OK`. El esquema se crea solo al arrancar
(AutoMigrate). Pruebas y análisis estático:

```bash
make test   # incluye las pruebas del motor de costos
make vet
```

## Producción

Compila a un binario estático — sin runtime que instalar en el servidor:

```bash
make build-linux          # bin/api-linux + bin/seed-linux (Linux x86_64)
```

El pipeline (`.github/workflows/main.yml`) construye, sube el tarball al
servidor y ejecuta el script de release; systemd supervisa el proceso. Piezas
en `deploy/`: `idilica-backend.sh` (swap de release + health check),
`idilica-api.service` (unidad systemd) y `nginx-api-idilica.conf`. Variables de
entorno de producción: ver `.env.prod.example`.

CI en cada PR: `go vet` + pruebas + build (`.github/workflows/ci.yml`).
