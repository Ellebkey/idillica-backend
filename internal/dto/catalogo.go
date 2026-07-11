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

	// Inventario
	Existencia float64 `json:"existencia"`
	Minimo     float64 `json:"minimo"`
	CaducaAt   *string `json:"caducaAt"` // YYYY-MM-DD o null

	// Escalado de recetas (normal | leudante | sazon)
	Escalado string `json:"escalado"`
}

type HerramientaDto struct {
	ID      string `json:"id"`
	Nombre  string `json:"nombre"`
	Detalle string `json:"detalle"`
	Estado  string `json:"estado"`
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
	Herramientas []HerramientaDto `json:"herramientas"`
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
	Existencia float64         `json:"existencia" binding:"min=0"`
	Minimo     float64         `json:"minimo" binding:"min=0"`
	Escalado   string          `json:"escalado" binding:"omitempty,oneof=normal leudante sazon"`
	Productos  []ProductoInput `json:"productos" binding:"required,min=1,dive"`
}

type UpdateIngredienteDto struct {
	Nombre     *string `json:"nombre" binding:"omitempty,min=1,max=120"`
	UnidadBase *string `json:"unidadBase" binding:"omitempty,oneof=kg L pieza"`
	Escalado   *string `json:"escalado" binding:"omitempty,oneof=normal leudante sazon"`
}

// ProducirDto — multiplicador del lote ("produje ×3"); omitido = 1.
type ProducirDto struct {
	Factor float64 `json:"factor" binding:"omitempty,gt=0"`
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

// RegistrarCompraDto — sheet "Registrar compra": N presentaciones del producto
// activo entran al inventario; si el precio cambió, se actualiza (+historial).
type RegistrarCompraDto struct {
	Unidades int     `json:"unidades" binding:"required,min=1"`
	Precio   float64 `json:"precio" binding:"required,gt=0"`
}

// ExistenciaDto — ajuste directo de inventario (conteo individual / edición).
// CaducaAt: "YYYY-MM-DD" para fijar, "" para limpiar, ausente para no tocar.
type ExistenciaDto struct {
	Existencia *float64 `json:"existencia" binding:"omitempty,min=0"`
	Minimo     *float64 `json:"minimo" binding:"omitempty,min=0"`
	CaducaAt   *string  `json:"caducaAt"`
}

// ConteoDto — "Contar mi cocina": aplica las existencias contadas en bloque.
type ConteoDto struct {
	Items []ConteoItem `json:"items" binding:"required,min=1,dive"`
}

type ConteoItem struct {
	IngredienteID string  `json:"ingredienteId" binding:"required,uuid"`
	Cantidad      float64 `json:"cantidad" binding:"min=0"`
}

type HerramientaInput struct {
	Nombre  string `json:"nombre" binding:"required,min=1,max=120"`
	Detalle string `json:"detalle" binding:"omitempty,max=160"`
	Estado  string `json:"estado" binding:"omitempty,max=60"`
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
