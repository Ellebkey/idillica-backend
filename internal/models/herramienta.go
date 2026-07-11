// herramienta.go — equipo de la cocina (batidora, moldes, charolas…).
// Las herramientas NO se costean por receta: su desgaste vive en los gastos
// de operación mensuales de la cocina.
package models

import "time"

type Herramienta struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CocinaID  string    `gorm:"column:cocina_id;type:uuid;not null;index"`
	Nombre    string    `gorm:"size:120;not null"`
	Detalle   string    `gorm:"size:160"`
	Estado    string    `gorm:"size:60;not null;default:Buen estado"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Herramienta) TableName() string { return "herramienta" }
