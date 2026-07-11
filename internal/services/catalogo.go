// catalogo.go — serves the whole raw catalog of a cocina in one GET.
// The frontend's costing engine (ported from the prototype) recalculates
// everything live from this payload; the backend never persists costs.
package services

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"idilica-backend-go/internal/apperrors"
	"idilica-backend-go/internal/dto"
	"idilica-backend-go/internal/models"
	"idilica-backend-go/internal/seeddata"
)

type CatalogoService struct {
	db     *gorm.DB
	cocina *CocinaService
	logger *slog.Logger
}

func NewCatalogoService(db *gorm.DB, cocinaService *CocinaService, logger *slog.Logger) *CatalogoService {
	return &CatalogoService{db: db, cocina: cocinaService, logger: logger}
}

// RequireEditor is the write guard shared by the domain services: member of
// the cocina AND not a viewer.
func (s *CatalogoService) RequireEditor(ctx context.Context, userID, cocinaID string) error {
	member, err := s.cocina.RequireMembership(ctx, userID, cocinaID)
	if err != nil {
		return err
	}
	if member.Rol == models.RolViewer {
		return apperrors.NewForbidden("Access denied: viewers cannot edit")
	}
	return nil
}

func (s *CatalogoService) Get(ctx context.Context, userID, cocinaID string) (*dto.CatalogoDto, error) {
	member, err := s.cocina.RequireMembership(ctx, userID, cocinaID)
	if err != nil {
		return nil, err
	}

	var ingredientes []models.Ingrediente
	err = s.db.WithContext(ctx).
		Preload("Productos", func(db *gorm.DB) *gorm.DB { return db.Order("orden ASC, created_at ASC") }).
		Preload("Productos.Historial", func(db *gorm.DB) *gorm.DB { return db.Order("fecha ASC") }).
		Where("cocina_id = ?", cocinaID).
		Order("nombre ASC").
		Find(&ingredientes).Error
	if err != nil {
		return nil, err
	}

	var recetas []models.Receta
	err = s.db.WithContext(ctx).
		Preload("Lineas", func(db *gorm.DB) *gorm.DB { return db.Order("orden ASC") }).
		Where("cocina_id = ?", cocinaID).
		Order("nombre ASC").
		Find(&recetas).Error
	if err != nil {
		return nil, err
	}

	var herramientas []models.Herramienta
	err = s.db.WithContext(ctx).
		Where("cocina_id = ?", cocinaID).
		Order("nombre ASC").
		Find(&herramientas).Error
	if err != nil {
		return nil, err
	}

	catalogo := &dto.CatalogoDto{
		Cocina:       *toCocinaDto(member.Cocina, member.Rol),
		Ingredientes: make([]dto.IngredienteDto, 0, len(ingredientes)),
		Recetas:      make([]dto.RecetaDto, 0, len(recetas)),
		Herramientas: make([]dto.HerramientaDto, 0, len(herramientas)),
		Categorias:   seeddata.Categorias,
		Alergenos:    seeddata.Alergenos,
	}
	for i := range ingredientes {
		catalogo.Ingredientes = append(catalogo.Ingredientes, toIngredienteDto(&ingredientes[i]))
	}
	for i := range recetas {
		catalogo.Recetas = append(catalogo.Recetas, toRecetaDto(&recetas[i]))
	}
	for _, h := range herramientas {
		catalogo.Herramientas = append(catalogo.Herramientas, toHerramientaDto(&h))
	}
	return catalogo, nil
}

func toHerramientaDto(h *models.Herramienta) dto.HerramientaDto {
	return dto.HerramientaDto{ID: h.ID, Nombre: h.Nombre, Detalle: h.Detalle, Estado: h.Estado}
}

func toIngredienteDto(ing *models.Ingrediente) dto.IngredienteDto {
	out := dto.IngredienteDto{
		ID:         ing.ID,
		Nombre:     ing.Nombre,
		UnidadBase: ing.UnidadBase,
		Merma:      dto.MermaDto{Pct: ing.MermaPct, Origen: ing.MermaOrigen},
		Productos:  make([]dto.ProductoDto, 0, len(ing.Productos)),
		Existencia: ing.Existencia,
		Minimo:     ing.Minimo,
		Escalado:   ing.Escalado,
	}
	if ing.CaducaAt != nil {
		fecha := ing.CaducaAt.Format("2006-01-02")
		out.CaducaAt = &fecha
	}
	for i := range ing.Productos {
		out.Productos = append(out.Productos, toProductoDto(&ing.Productos[i]))
	}
	return out
}

func toProductoDto(p *models.ProductoCompra) dto.ProductoDto {
	out := dto.ProductoDto{
		ID:                  p.ID,
		Marca:               p.Marca,
		Presentacion:        p.Presentacion,
		Cantidad:            p.Cantidad,
		Precio:              p.Precio,
		Proveedor:           p.Proveedor,
		Activo:              p.Activo,
		Orden:               p.Orden,
		PrecioActualizadoAt: p.PrecioActualizadoAt.UTC().Format(time.RFC3339),
		Historial:           make([]dto.HistorialPrecioDto, 0, len(p.Historial)),
	}
	for _, h := range p.Historial {
		out.Historial = append(out.Historial, dto.HistorialPrecioDto{
			Precio: h.Precio,
			Fecha:  h.Fecha.UTC().Format(time.RFC3339),
		})
	}
	return out
}

func toRecetaDto(r *models.Receta) dto.RecetaDto {
	out := dto.RecetaDto{
		ID:               r.ID,
		Nombre:           r.Nombre,
		Categoria:        r.Categoria,
		Porciones:        r.Porciones,
		Etiqueta:         r.Etiqueta,
		EtiquetaSingular: r.EtiquetaSingular,
		RendimientoKg:    r.RendimientoKg,
		PrecioVenta:      r.PrecioVenta,
		IvaPct:           r.IvaPct,
		EsSubreceta:      r.EsSubreceta,
		Alergenos:        r.Alergenos,
		Pasos:            r.Pasos,
		Fotos:            r.Fotos,
		Lineas:           make([]dto.LineaDto, 0, len(r.Lineas)),
	}
	if out.Alergenos == nil {
		out.Alergenos = []string{}
	}
	if out.Pasos == nil {
		out.Pasos = []string{}
	}
	if out.Fotos == nil {
		out.Fotos = []string{}
	}
	for _, l := range r.Lineas {
		out.Lineas = append(out.Lineas, dto.LineaDto{
			ID:            l.ID,
			IngredienteID: l.IngredienteID,
			RecetaID:      l.SubRecetaID,
			Cantidad:      l.Cantidad,
			Orden:         l.Orden,
		})
	}
	return out
}
