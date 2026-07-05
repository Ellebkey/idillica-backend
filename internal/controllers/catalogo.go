// catalogo.go — GET /cocinas/:id/catalogo: the whole raw catalog in one shot.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"idilica-backend-go/internal/dto"
	"idilica-backend-go/internal/services"
	"idilica-backend-go/internal/utils"
)

type CatalogoController struct {
	catalogoService *services.CatalogoService
}

func NewCatalogoController(catalogoService *services.CatalogoService) *CatalogoController {
	return &CatalogoController{catalogoService: catalogoService}
}

func (cc *CatalogoController) Get(c *gin.Context) {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	params, err := utils.BindUri[dto.UUIDParam](c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	catalogo, err := cc.catalogoService.Get(c.Request.Context(), userID, params.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, catalogo)
}
