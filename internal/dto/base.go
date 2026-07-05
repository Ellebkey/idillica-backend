// Package dto mirrors src/interfaces (the *.dto.ts files). The `binding` tags
// replace the Joi schemas of src/validations: Gin validates on bind and the
// utils package converts failures into the same VALIDATION_ERROR response.
package dto

// UUIDParam ≈ entityUuid schema (shared.validation.ts) for /:id route params.
type UUIDParam struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// ProductoParam — rutas /ingredientes/:id/producto-activo/:productoId.
type ProductoParam struct {
	ID         string `uri:"id" binding:"required,uuid"`
	ProductoID string `uri:"productoId" binding:"required,uuid"`
}
