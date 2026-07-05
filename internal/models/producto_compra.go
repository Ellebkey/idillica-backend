// producto_compra.go — a purchase presentation of an ingredient (brand,
// package, price, supplier). Exactly ONE per ingredient is `activo`; only the
// active one feeds all recipe costs.
package models

import "time"

type ProductoCompra struct {
	ID            string  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	IngredienteID string  `gorm:"column:ingrediente_id;type:uuid;not null;index"`
	Marca         string  `gorm:"size:120;not null"`
	Presentacion  string  `gorm:"size:120;not null"` // "Bulto 44 kg"
	Cantidad      float64 `gorm:"not null"`          // contenido en unidad base (44 = 44 kg)
	Precio        float64 `gorm:"not null"`
	Proveedor     string  `gorm:"size:120"`
	Activo        bool    `gorm:"not null;default:false"`
	Orden         int     `gorm:"not null;default:0"`
	// Feeds the "precio viejo" indicator (>60 días ⇒ punto ámbar)
	PrecioActualizadoAt time.Time `gorm:"column:precio_actualizado_at"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`

	Historial []HistorialPrecio `gorm:"foreignKey:ProductoID"`
}

func (ProductoCompra) TableName() string { return "producto_compra" }

// HistorialPrecio — one row per price change, for the sparkline and trend.
type HistorialPrecio struct {
	ID         string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ProductoID string    `gorm:"column:producto_id;type:uuid;not null;index"`
	Precio     float64   `gorm:"not null"`
	Fecha      time.Time `gorm:"not null"`
}

func (HistorialPrecio) TableName() string { return "historial_precio" }
