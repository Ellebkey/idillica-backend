// cocina.go — rutas de cocinas.
package routes

import (
	"github.com/gin-gonic/gin"

	"idilica-backend-go/internal/controllers"
	"idilica-backend-go/internal/middlewares"
)

func registerCocinaRoutes(api *gin.RouterGroup, ctrl *controllers.CocinaController, auth *middlewares.Auth) {
	canAccess := auth.CheckAuth()

	api.GET("/cocinas", canAccess, ctrl.List)
	api.POST("/cocinas", canAccess, ctrl.Create)
	api.GET("/cocinas/:id", canAccess, ctrl.GetByID)
	api.PUT("/cocinas/:id", canAccess, ctrl.Update)
}
