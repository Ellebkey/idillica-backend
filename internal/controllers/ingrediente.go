// ingrediente.go — handlers del catálogo de ingredientes.
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"idilica-backend-go/internal/dto"
	"idilica-backend-go/internal/services"
	"idilica-backend-go/internal/utils"
)

type IngredienteController struct {
	service *services.IngredienteService
}

func NewIngredienteController(service *services.IngredienteService) *IngredienteController {
	return &IngredienteController{service: service}
}

// handle deduplicates the bind→service→respond dance shared by all handlers.
func handle[I any](c *gin.Context, call func(userID string, params *dto.UUIDParam, input *I) (any, error)) {
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
	input, err := utils.BindJSON[I](c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	result, err := call(userID, params, input)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ic *IngredienteController) Create(c *gin.Context) {
	handle(c, func(userID string, params *dto.UUIDParam, input *dto.CreateIngredienteDto) (any, error) {
		return ic.service.Create(c.Request.Context(), userID, params.ID, input)
	})
}

func (ic *IngredienteController) Update(c *gin.Context) {
	handle(c, func(userID string, params *dto.UUIDParam, input *dto.UpdateIngredienteDto) (any, error) {
		return ic.service.Update(c.Request.Context(), userID, params.ID, input)
	})
}

func (ic *IngredienteController) Delete(c *gin.Context) {
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
	if err := ic.service.Delete(c.Request.Context(), userID, params.ID); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (ic *IngredienteController) AddProducto(c *gin.Context) {
	handle(c, func(userID string, params *dto.UUIDParam, input *dto.ProductoInput) (any, error) {
		return ic.service.AddProducto(c.Request.Context(), userID, params.ID, input)
	})
}

func (ic *IngredienteController) UpdatePrecio(c *gin.Context) {
	handle(c, func(userID string, params *dto.UUIDParam, input *dto.UpdatePrecioDto) (any, error) {
		return ic.service.UpdatePrecio(c.Request.Context(), userID, params.ID, input)
	})
}

func (ic *IngredienteController) UpdateProducto(c *gin.Context) {
	handle(c, func(userID string, params *dto.UUIDParam, input *dto.UpdateProductoDto) (any, error) {
		return ic.service.UpdateProducto(c.Request.Context(), userID, params.ID, input)
	})
}

func (ic *IngredienteController) ActivarProducto(c *gin.Context) {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	params, err := utils.BindUri[dto.ProductoParam](c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	result, err := ic.service.ActivarProducto(c.Request.Context(), userID, params.ID, params.ProductoID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ic *IngredienteController) RegistrarCompra(c *gin.Context) {
	handle(c, func(userID string, params *dto.UUIDParam, input *dto.RegistrarCompraDto) (any, error) {
		return ic.service.RegistrarCompra(c.Request.Context(), userID, params.ID, input)
	})
}

func (ic *IngredienteController) SetExistencia(c *gin.Context) {
	handle(c, func(userID string, params *dto.UUIDParam, input *dto.ExistenciaDto) (any, error) {
		return ic.service.SetExistencia(c.Request.Context(), userID, params.ID, input)
	})
}

// Conteo — el :id de la ruta es la COCINA (aplica existencias en bloque).
func (ic *IngredienteController) Conteo(c *gin.Context) {
	handle(c, func(userID string, params *dto.UUIDParam, input *dto.ConteoDto) (any, error) {
		return ic.service.Conteo(c.Request.Context(), userID, params.ID, input)
	})
}

func (ic *IngredienteController) SetMerma(c *gin.Context) {
	handle(c, func(userID string, params *dto.UUIDParam, input *dto.SetMermaDto) (any, error) {
		return ic.service.SetMerma(c.Request.Context(), userID, params.ID, input)
	})
}

func (ic *IngredienteController) AddMedicion(c *gin.Context) {
	handle(c, func(userID string, params *dto.UUIDParam, input *dto.MedicionInput) (any, error) {
		return ic.service.AddMedicion(c.Request.Context(), userID, params.ID, input)
	})
}
