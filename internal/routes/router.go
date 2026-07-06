// Package routes arma el engine: middlewares, wiring explícito
// (services → controllers → rutas) y registro de endpoints — el grafo de
// dependencias completo se lee en un solo lugar.
package routes

import (
	"log/slog"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"idilica-backend-go/internal/apperrors"
	"idilica-backend-go/internal/config"
	"idilica-backend-go/internal/controllers"
	"idilica-backend-go/internal/middlewares"
	"idilica-backend-go/internal/services"
)

// New builds the fully-wired Gin engine.
func New(cfg *config.Config, logger *slog.Logger, db *gorm.DB, redisClient *redis.Client) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else if cfg.IsTest() {
		gin.SetMode(gin.TestMode)
	}

	router := gin.New()
	_ = router.SetTrustedProxies(nil) // direct clients in dev; set real proxies on deploy

	// Cadena de middlewares. El ORDEN importa: el manejador de errores debe ir
	// DESPUÉS (más adentro) de gzip — si fuera antes, al abortar un request
	// (p. ej. token expirado) gzip cierra la respuesta vacía con 200 durante el
	// unwind y el 401/500 jamás se escribe.
	router.Use(middlewares.SecurityHeaders())
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	router.Use(apperrors.Recovery(cfg.Env, logger))
	if cfg.IsDevelopment() || cfg.IsTest() {
		router.Use(middlewares.RequestLogger(logger))
	}
	router.Use(apperrors.ErrorHandler(cfg.Env, logger))

	// --- Wiring ---
	jwtService := services.NewJWTService(cfg.JWTSecret, redisClient, logger)
	emailService := services.NewEmailService(cfg, logger)
	authService := services.NewAuthService(db, redisClient, cfg, jwtService, emailService, logger)
	cocinaService := services.NewCocinaService(db, logger)
	catalogoService := services.NewCatalogoService(db, cocinaService, logger)
	ingredienteService := services.NewIngredienteService(db, catalogoService, logger)
	recetaService := services.NewRecetaService(db, catalogoService, logger)

	authController := controllers.NewAuthController(authService)
	cocinaController := controllers.NewCocinaController(cocinaService)
	catalogoController := controllers.NewCatalogoController(catalogoService)
	ingredienteController := controllers.NewIngredienteController(ingredienteService)
	recetaController := controllers.NewRecetaController(recetaService)
	authMiddleware := middlewares.NewAuth(jwtService)

	// Prefijo /api + rate limit general (omitido en test)
	api := router.Group("/api")
	if !cfg.IsTest() {
		api.Use(middlewares.APIRateLimiter(cfg.Env))
	}

	api.GET("/health-check", func(c *gin.Context) {
		c.String(200, "OK")
	})

	registerAuthRoutes(api, authController, authMiddleware, cfg)
	registerCocinaRoutes(api, cocinaController, authMiddleware)
	registerDominioRoutes(api, catalogoController, ingredienteController, recetaController, authMiddleware)

	// Rutas no registradas
	router.NoRoute(apperrors.NotFoundHandler(cfg.Env, logger))

	return router
}
