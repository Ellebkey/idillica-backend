// receta.go — handlers de recetas (guardado completo desde el editor).
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"idilica-backend-go/internal/dto"
	"idilica-backend-go/internal/services"
	"idilica-backend-go/internal/utils"
)

type RecetaController struct {
	service *services.RecetaService
}

func NewRecetaController(service *services.RecetaService) *RecetaController {
	return &RecetaController{service: service}
}

func (rc *RecetaController) Create(c *gin.Context) {
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
	input, err := utils.BindJSON[dto.SaveRecetaDto](c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	receta, err := rc.service.Create(c.Request.Context(), userID, params.ID, input)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, receta)
}

func (rc *RecetaController) Update(c *gin.Context) {
	handle(c, func(userID string, params *dto.UUIDParam, input *dto.SaveRecetaDto) (any, error) {
		return rc.service.Update(c.Request.Context(), userID, params.ID, input)
	})
}

// Producir — "Produje esta receta": descuenta el inventario recursivamente.
// El body {factor} es opcional (default 1; ×3 = receta escalada a triple).
func (rc *RecetaController) Producir(c *gin.Context) {
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
	input := &dto.ProducirDto{}
	if c.Request.ContentLength > 0 {
		if input, err = utils.BindJSON[dto.ProducirDto](c); err != nil {
			_ = c.Error(err)
			return
		}
	}
	afectados, err := rc.service.Producir(c.Request.Context(), userID, params.ID, input.Factor)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, afectados)
}

func (rc *RecetaController) Delete(c *gin.Context) {
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
	if err := rc.service.Delete(c.Request.Context(), userID, params.ID); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
