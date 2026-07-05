// cocina.go ≈ cocina.dto.ts. Pointers on Update = "optional field present or
// not" (the Go equivalent of a Joi .min(1) partial update).
package dto

type CreateCocinaDto struct {
	Name             string   `json:"name" binding:"required,min=1,max=120"`
	Moneda           string   `json:"moneda" binding:"omitempty,len=3"`
	ImpuestoDefault  *float64 `json:"impuestoDefault" binding:"omitempty,min=0,max=1"`
	FoodCostObjetivo *float64 `json:"foodCostObjetivo" binding:"omitempty,min=0,max=1"`
}

type UpdateCocinaDto struct {
	Name             *string  `json:"name" binding:"omitempty,min=1,max=120"`
	Moneda           *string  `json:"moneda" binding:"omitempty,len=3"`
	ImpuestoDefault  *float64 `json:"impuestoDefault" binding:"omitempty,min=0,max=1"`
	FoodCostObjetivo *float64 `json:"foodCostObjetivo" binding:"omitempty,min=0,max=1"`

	// Gastos de operación mensuales (Ajustes → "Gastos de operación · al mes")
	GastoSueldos           *float64 `json:"gastoSueldos" binding:"omitempty,min=0"`
	GastoGas               *float64 `json:"gastoGas" binding:"omitempty,min=0"`
	GastoLuz               *float64 `json:"gastoLuz" binding:"omitempty,min=0"`
	GastoEquipo            *float64 `json:"gastoEquipo" binding:"omitempty,min=0"`
	ComprasIngredientesMes *float64 `json:"comprasIngredientesMes" binding:"omitempty,min=0"`
}

// CocinaDto ≈ CocinaDto of the Node backend (same JSON shape) + opex.
type CocinaDto struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Moneda           string  `json:"moneda"`
	ImpuestoDefault  float64 `json:"impuestoDefault"`
	FoodCostObjetivo float64 `json:"foodCostObjetivo"`

	GastoSueldos           float64 `json:"gastoSueldos"`
	GastoGas               float64 `json:"gastoGas"`
	GastoLuz               float64 `json:"gastoLuz"`
	GastoEquipo            float64 `json:"gastoEquipo"`
	ComprasIngredientesMes float64 `json:"comprasIngredientesMes"`

	Rol       string `json:"rol"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
