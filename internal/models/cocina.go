// cocina.go — el workspace multi-tenant.
// impuestoDefault / foodCostObjetivo are fractions (0.16 = 16%).
package models

import "time"

type Cocina struct {
	ID               string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name             string    `gorm:"size:120;not null"`
	Moneda           string    `gorm:"size:3;not null;default:MXN"`
	ImpuestoDefault  float64   `gorm:"column:impuesto_default;type:decimal(5,4);not null;default:0.16"`
	FoodCostObjetivo float64   `gorm:"column:food_cost_objetivo;type:decimal(5,4);not null;default:0.30"`

	// Gastos de operación mensuales (modo simple del handoff):
	// tasaOperación = Σ gastos / comprasIngredientesMes;
	// costoReal(receta) = costo ingredientes × (1 + tasa).
	GastoSueldos          float64 `gorm:"column:gasto_sueldos;not null;default:0"`
	GastoGas              float64 `gorm:"column:gasto_gas;not null;default:0"`
	GastoLuz              float64 `gorm:"column:gasto_luz;not null;default:0"`
	GastoEquipo           float64 `gorm:"column:gasto_equipo;not null;default:0"`
	ComprasIngredientesMes float64 `gorm:"column:compras_ingredientes_mes;not null;default:0"`

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`

	Members []CocinaMember `gorm:"foreignKey:CocinaID"`
}

func (Cocina) TableName() string { return "cocina" }
