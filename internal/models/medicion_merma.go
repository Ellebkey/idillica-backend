// medicion_merma.go — a scale measurement from the merma wizard.
// merma% = round((1 − (limpio + aprovechado) / entero) × 100), clamped 0–95.
package models

import "time"

type MedicionMerma struct {
	ID            string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	IngredienteID string    `gorm:"column:ingrediente_id;type:uuid;not null;index"`
	PesoEntero    float64   `gorm:"column:peso_entero;not null"`  // kg
	PesoLimpio    float64   `gorm:"column:peso_limpio;not null"`  // kg
	Aprovechado   float64   `gorm:"not null;default:0"`           // kg reutilizados del desperdicio
	PctResultante float64   `gorm:"column:pct_resultante;not null"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (MedicionMerma) TableName() string { return "medicion_merma" }
