// receta.go — a recipe (or sub-recipe) with its lines.
//
// A line points to EITHER an ingredient (cantidad in the ingredient's base
// unit) OR another recipe (cantidad in kg of that sub-recipe). Sub-recipe
// line cost = (cantidadKg / rendimientoKg of the sub) × cost of the sub —
// recursive, with cycle detection enforced in the service layer.
package models

import "time"

type Receta struct {
	ID               string      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CocinaID         string      `gorm:"column:cocina_id;type:uuid;not null;index"`
	Nombre           string      `gorm:"size:120;not null"`
	Categoria        string      `gorm:"size:40;not null;default:Pasteles"`
	Porciones        int         `gorm:"not null;default:1"`
	Etiqueta         string      `gorm:"size:40;not null;default:piezas"` // "piezas", "rebanadas", "Rinde 950 g"
	EtiquetaSingular string      `gorm:"column:etiqueta_singular;size:40;not null;default:pieza"`
	RendimientoKg    float64     `gorm:"column:rendimiento_kg;not null;default:0"`
	PrecioVenta      *float64    `gorm:"column:precio_venta"` // null = sin precio (subrecetas)
	IvaPct           float64     `gorm:"column:iva_pct;not null;default:16"`
	EsSubreceta      bool        `gorm:"column:es_subreceta;not null;default:false"`
	Alergenos        StringSlice `gorm:"type:json"`
	Pasos            StringSlice `gorm:"type:json"`
	Fotos            StringSlice `gorm:"type:json"`
	CreatedAt        time.Time   `gorm:"column:created_at"`
	UpdatedAt        time.Time   `gorm:"column:updated_at"`

	Lineas []RecetaLinea `gorm:"foreignKey:RecetaID"`
}

func (Receta) TableName() string { return "receta" }

type RecetaLinea struct {
	ID            string  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	RecetaID      string  `gorm:"column:receta_id;type:uuid;not null;index"`
	Orden         int     `gorm:"not null;default:0"`
	IngredienteID *string `gorm:"column:ingrediente_id;type:uuid"` // XOR con SubRecetaID
	SubRecetaID   *string `gorm:"column:sub_receta_id;type:uuid"`
	Cantidad      float64 `gorm:"not null"` // unidad base del ingrediente, o kg si es subreceta
}

func (RecetaLinea) TableName() string { return "receta_linea" }
