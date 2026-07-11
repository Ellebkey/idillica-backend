// herramienta.go — handlers del equipo de la cocina.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"idilica-backend-go/internal/dto"
	"idilica-backend-go/internal/services"
	"idilica-backend-go/internal/utils"
)

type HerramientaController struct {
	service *services.HerramientaService
}

func NewHerramientaController(service *services.HerramientaService) *HerramientaController {
	return &HerramientaController{service: service}
}

// Create — el :id de la ruta es la cocina.
func (hc *HerramientaController) Create(c *gin.Context) {
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
	input, err := utils.BindJSON[dto.HerramientaInput](c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	herramienta, err := hc.service.Create(c.Request.Context(), userID, params.ID, input)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, herramienta)
}

func (hc *HerramientaController) Update(c *gin.Context) {
	handle(c, func(userID string, params *dto.UUIDParam, input *dto.HerramientaInput) (any, error) {
		return hc.service.Update(c.Request.Context(), userID, params.ID, input)
	})
}

func (hc *HerramientaController) Delete(c *gin.Context) {
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
	if err := hc.service.Delete(c.Request.Context(), userID, params.ID); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
