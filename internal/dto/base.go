// Package dto define las entradas y salidas del API. Los tags `binding`
// validan al hacer bind; utils convierte los fallos en la respuesta
// VALIDATION_ERROR estándar.
package dto

// UUIDParam valida los parámetros de ruta /:id.
type UUIDParam struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// ProductoParam — rutas /ingredientes/:id/producto-activo/:productoId.
type ProductoParam struct {
	ID         string `uri:"id" binding:"required,uuid"`
	ProductoID string `uri:"productoId" binding:"required,uuid"`
}
