// dominio.go — rutas del dominio de costeo (catálogo, ingredientes, recetas).
package routes

import (
	"github.com/gin-gonic/gin"

	"idilica-backend-go/internal/controllers"
	"idilica-backend-go/internal/middlewares"
)

func registerDominioRoutes(
	api *gin.RouterGroup,
	catalogo *controllers.CatalogoController,
	ingredientes *controllers.IngredienteController,
	recetas *controllers.RecetaController,
	auth *middlewares.Auth,
) {
	canAccess := auth.CheckAuth()

	// Catálogo completo (una sola carga; el motor del frontend hace el resto)
	api.GET("/cocinas/:id/catalogo", canAccess, catalogo.Get)

	// Ingredientes y sus productos de compra
	api.POST("/cocinas/:id/ingredientes", canAccess, ingredientes.Create)
	api.PUT("/ingredientes/:id", canAccess, ingredientes.Update)
	api.DELETE("/ingredientes/:id", canAccess, ingredientes.Delete)
	api.POST("/ingredientes/:id/productos", canAccess, ingredientes.AddProducto)
	api.PUT("/ingredientes/:id/producto-activo/:productoId", canAccess, ingredientes.ActivarProducto)
	api.PUT("/ingredientes/:id/merma", canAccess, ingredientes.SetMerma)
	api.POST("/ingredientes/:id/mediciones", canAccess, ingredientes.AddMedicion)
	api.PUT("/productos/:id", canAccess, ingredientes.UpdateProducto)
	api.PUT("/productos/:id/precio", canAccess, ingredientes.UpdatePrecio)

	// Recetas (guardado completo desde el editor)
	api.POST("/cocinas/:id/recetas", canAccess, recetas.Create)
	api.PUT("/recetas/:id", canAccess, recetas.Update)
	api.DELETE("/recetas/:id", canAccess, recetas.Delete)
}
