// catalogo.go — DTOs del catálogo completo de una cocina.
//
// Arquitectura del handoff: el frontend carga TODO el catálogo crudo en un
// solo GET y su motor de costos (portado del prototipo) recalcula en vivo a
// cada tecleo. El backend persiste y valida; los costos nunca se guardan.
package dto

// ===== Salida =====

type HistorialPrecioDto struct {
	Precio float64 `json:"precio"`
	Fecha  string  `json:"fecha"`
}

type ProductoDto struct {
	ID                  string               `json:"id"`
	Marca               string               `json:"marca"`
	Presentacion        string               `json:"presentacion"`
	Cantidad            float64              `json:"cantidad"`
	Precio              float64              `json:"precio"`
	Proveedor           string               `json:"proveedor"`
	Activo              bool                 `json:"activo"`
	Orden               int                  `json:"orden"`
	PrecioActualizadoAt string               `json:"precioActualizadoAt"`
	Historial           []HistorialPrecioDto `json:"historial"`
}

type MermaDto struct {
	Pct    float64 `json:"pct"`
	Origen string  `json:"origen"`
}

type IngredienteDto struct {
	ID         string        `json:"id"`
	Nombre     string        `json:"nombre"`
	UnidadBase string        `json:"unidadBase"`
	Merma      MermaDto      `json:"merma"`
	Productos  []ProductoDto `json:"productos"`
}

type LineaDto struct {
	ID            string  `json:"id"`
	IngredienteID *string `json:"ingredienteId,omitempty"`
	RecetaID      *string `json:"recetaId,omitempty"`
	Cantidad      float64 `json:"cantidad"`
	Orden         int     `json:"orden"`
}

type RecetaDto struct {
	ID               string     `json:"id"`
	Nombre           string     `json:"nombre"`
	Categoria        string     `json:"categoria"`
	Porciones        int        `json:"porciones"`
	Etiqueta         string     `json:"etiqueta"`
	EtiquetaSingular string     `json:"etiquetaSingular"`
	RendimientoKg    float64    `json:"rendimientoKg"`
	PrecioVenta      *float64   `json:"precioVenta"`
	IvaPct           float64    `json:"ivaPct"`
	EsSubreceta      bool       `json:"esSubreceta"`
	Alergenos        []string   `json:"alergenos"`
	Pasos            []string   `json:"pasos"`
	Fotos            []string   `json:"fotos"`
	Lineas           []LineaDto `json:"lineas"`
}

type CatalogoDto struct {
	Cocina       CocinaDto        `json:"cocina"`
	Ingredientes []IngredienteDto `json:"ingredientes"`
	Recetas      []RecetaDto      `json:"recetas"`
	Categorias   []string         `json:"categorias"`
	Alergenos    []string         `json:"alergenos"`
}

// ===== Entrada =====

type ProductoInput struct {
	Marca        string  `json:"marca" binding:"required,min=1,max=120"`
	Presentacion string  `json:"presentacion" binding:"required,min=1,max=120"`
	Cantidad     float64 `json:"cantidad" binding:"required,gt=0"`
	Precio       float64 `json:"precio" binding:"required,gt=0"`
	Proveedor    string  `json:"proveedor" binding:"omitempty,max=120"`
}

type CreateIngredienteDto struct {
	Nombre     string          `json:"nombre" binding:"required,min=1,max=120"`
	UnidadBase string          `json:"unidadBase" binding:"required,oneof=kg L pieza"`
	MermaPct   float64         `json:"mermaPct" binding:"min=0,max=95"`
	Origen     string          `json:"mermaOrigen" binding:"omitempty,oneof=referencia manual medido"`
	Productos  []ProductoInput `json:"productos" binding:"required,min=1,dive"`
}

type UpdateIngredienteDto struct {
	Nombre     *string `json:"nombre" binding:"omitempty,min=1,max=120"`
	UnidadBase *string `json:"unidadBase" binding:"omitempty,oneof=kg L pieza"`
}

type SetMermaDto struct {
	Pct    float64 `json:"pct" binding:"min=0,max=95"`
	Origen string  `json:"origen" binding:"required,oneof=referencia manual medido"`
}

// MedicionInput — pesos del wizard, en kg.
type MedicionInput struct {
	PesoEntero  float64 `json:"pesoEntero" binding:"required,gt=0"`
	PesoLimpio  float64 `json:"pesoLimpio" binding:"required,gt=0"`
	Aprovechado float64 `json:"aprovechado" binding:"min=0"`
}

type UpdatePrecioDto struct {
	Precio float64 `json:"precio" binding:"required,gt=0"`
}

type UpdateProductoDto struct {
	Marca        *string  `json:"marca" binding:"omitempty,min=1,max=120"`
	Presentacion *string  `json:"presentacion" binding:"omitempty,min=1,max=120"`
	Cantidad     *float64 `json:"cantidad" binding:"omitempty,gt=0"`
	Proveedor    *string  `json:"proveedor" binding:"omitempty,max=120"`
}

type LineaInput struct {
	IngredienteID *string `json:"ingredienteId" binding:"omitempty,uuid"`
	RecetaID      *string `json:"recetaId" binding:"omitempty,uuid"`
	Cantidad      float64 `json:"cantidad" binding:"required,gt=0"`
}

// SaveRecetaDto — el editor guarda la receta completa (replace de líneas,
// pasos y alérgenos), igual que el botón "Guardar" del diseño.
type SaveRecetaDto struct {
	Nombre           string       `json:"nombre" binding:"required,min=1,max=120"`
	Categoria        string       `json:"categoria" binding:"required,max=40"`
	Porciones        int          `json:"porciones" binding:"required,min=1"`
	Etiqueta         string       `json:"etiqueta" binding:"omitempty,max=40"`
	EtiquetaSingular string       `json:"etiquetaSingular" binding:"omitempty,max=40"`
	RendimientoKg    float64      `json:"rendimientoKg" binding:"min=0"`
	PrecioVenta      *float64     `json:"precioVenta" binding:"omitempty,gt=0"`
	IvaPct           float64      `json:"ivaPct" binding:"min=0,max=100"`
	EsSubreceta      bool         `json:"esSubreceta"`
	Alergenos        []string     `json:"alergenos"`
	Pasos            []string     `json:"pasos"`
	Fotos            []string     `json:"fotos"`
	Lineas           []LineaInput `json:"lineas" binding:"dive"`
}
