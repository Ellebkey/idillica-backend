// ingrediente.go — an ingredient of the cocina's catalog.
//
// Costing rule (the whole point of the app):
//
//	costo unitario = precio del producto ACTIVO / cantidad presentación / (1 − merma/100)
//
// "por kilo, ya con desperdicio". Costs are NEVER persisted — always derived.
package models

import "time"

// Base units. "pieza" is for things bought/used by piece (huevo).
const (
	UnidadKg    = "kg"
	UnidadL     = "L"
	UnidadPieza = "pieza"
)

// Merma origins: reference table < manual estimate < measured with the scale.
const (
	MermaReferencia = "referencia"
	MermaManual     = "manual"
	MermaMedido     = "medido"
)

type Ingrediente struct {
	ID          string  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CocinaID    string  `gorm:"column:cocina_id;type:uuid;not null;index"`
	Nombre      string  `gorm:"size:120;not null"`
	UnidadBase  string  `gorm:"column:unidad_base;size:10;not null;default:kg"`
	MermaPct    float64 `gorm:"column:merma_pct;not null;default:0"` // 0–95, whole percent (25 = 25%)
	MermaOrigen string  `gorm:"column:merma_origen;size:12;not null;default:referencia"`

	// Inventario: existencia en unidad base, mínimo que dispara "queda poco"
	// (solo alerta cuando minimo > 0) y fecha de caducidad opcional.
	Existencia float64    `gorm:"not null;default:0"`
	Minimo     float64    `gorm:"not null;default:0"`
	CaducaAt   *time.Time `gorm:"column:caduca_at;type:date"`

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`

	Productos  []ProductoCompra `gorm:"foreignKey:IngredienteID"`
	Mediciones []MedicionMerma  `gorm:"foreignKey:IngredienteID"`
}

func (Ingrediente) TableName() string { return "ingrediente" }
