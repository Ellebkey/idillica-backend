// cocina.go — handlers de cocinas.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"idilica-backend-go/internal/dto"
	"idilica-backend-go/internal/services"
	"idilica-backend-go/internal/utils"
)

type CocinaController struct {
	cocinaService *services.CocinaService
}

func NewCocinaController(cocinaService *services.CocinaService) *CocinaController {
	return &CocinaController{cocinaService: cocinaService}
}

func (cc *CocinaController) Create(c *gin.Context) {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	d, err := utils.BindJSON[dto.CreateCocinaDto](c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	cocina, err := cc.cocinaService.Create(c.Request.Context(), userID, d)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, cocina)
}

func (cc *CocinaController) List(c *gin.Context) {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	cocinas, err := cc.cocinaService.FindAllForUser(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, cocinas)
}

func (cc *CocinaController) GetByID(c *gin.Context) {
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
	cocina, err := cc.cocinaService.FindByID(c.Request.Context(), userID, params.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, cocina)
}

func (cc *CocinaController) Update(c *gin.Context) {
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
	d, err := utils.BindJSON[dto.UpdateCocinaDto](c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	cocina, err := cc.cocinaService.Update(c.Request.Context(), userID, params.ID, d)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, cocina)
}
