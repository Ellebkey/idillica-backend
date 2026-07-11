// ingrediente.go — ingredient catalog: CRUD, purchase products, price updates
// (with history), active-product switch and merma (manual or measured).
package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"gorm.io/gorm"

	"idilica-backend-go/internal/apperrors"
	"idilica-backend-go/internal/dto"
	"idilica-backend-go/internal/models"
)

type IngredienteService struct {
	db       *gorm.DB
	catalogo *CatalogoService
	logger   *slog.Logger
}

func NewIngredienteService(db *gorm.DB, catalogoService *CatalogoService, logger *slog.Logger) *IngredienteService {
	return &IngredienteService{db: db, catalogo: catalogoService, logger: logger}
}

// load fetches the ingredient (with products) and enforces the write guard
// through its cocina.
func (s *IngredienteService) load(ctx context.Context, userID, ingredienteID string) (*models.Ingrediente, error) {
	var ing models.Ingrediente
	err := s.db.WithContext(ctx).
		Preload("Productos", func(db *gorm.DB) *gorm.DB { return db.Order("orden ASC, created_at ASC") }).
		First(&ing, "id = ?", ingredienteID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.NewNotFound("Ingrediente", ingredienteID)
	}
	if err != nil {
		return nil, err
	}
	if err := s.catalogo.RequireEditor(ctx, userID, ing.CocinaID); err != nil {
		return nil, err
	}
	return &ing, nil
}

func (s *IngredienteService) Create(ctx context.Context, userID, cocinaID string, d *dto.CreateIngredienteDto) (*dto.IngredienteDto, error) {
	if err := s.catalogo.RequireEditor(ctx, userID, cocinaID); err != nil {
		return nil, err
	}

	origen := d.Origen
	if origen == "" {
		origen = models.MermaReferencia
	}

	ing := models.Ingrediente{
		CocinaID:    cocinaID,
		Nombre:      d.Nombre,
		UnidadBase:  d.UnidadBase,
		MermaPct:    d.MermaPct,
		MermaOrigen: origen,
		Existencia:  d.Existencia,
		Minimo:      d.Minimo,
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&ing).Error; err != nil {
			return err
		}
		now := time.Now()
		for i, p := range d.Productos {
			producto := models.ProductoCompra{
				IngredienteID: ing.ID,
				Marca:         p.Marca, Presentacion: p.Presentacion,
				Cantidad: p.Cantidad, Precio: p.Precio, Proveedor: p.Proveedor,
				Activo: i == 0, Orden: i, PrecioActualizadoAt: now,
			}
			if err := tx.Create(&producto).Error; err != nil {
				return err
			}
			historial := models.HistorialPrecio{ProductoID: producto.ID, Precio: p.Precio, Fecha: now}
			if err := tx.Create(&historial).Error; err != nil {
				return err
			}
			ing.Productos = append(ing.Productos, producto)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("Ingrediente created", "ingredienteId", ing.ID, "cocinaId", cocinaID)
	out := toIngredienteDto(&ing)
	return &out, nil
}

func (s *IngredienteService) Update(ctx context.Context, userID, ingredienteID string, d *dto.UpdateIngredienteDto) (*dto.IngredienteDto, error) {
	ing, err := s.load(ctx, userID, ingredienteID)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{}
	if d.Nombre != nil {
		updates["nombre"] = *d.Nombre
	}
	if d.UnidadBase != nil {
		updates["unidad_base"] = *d.UnidadBase
	}
	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&models.Ingrediente{}).Where("id = ?", ing.ID).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return s.reload(ctx, ing.ID)
}

// Delete refuses when the ingredient is used by any recipe line of the cocina.
func (s *IngredienteService) Delete(ctx context.Context, userID, ingredienteID string) error {
	ing, err := s.load(ctx, userID, ingredienteID)
	if err != nil {
		return err
	}

	var usos int64
	err = s.db.WithContext(ctx).Model(&models.RecetaLinea{}).
		Where("ingrediente_id = ?", ing.ID).Count(&usos).Error
	if err != nil {
		return err
	}
	if usos > 0 {
		return apperrors.NewBusinessRule(
			fmt.Sprintf("No se puede eliminar: este ingrediente se usa en %d receta(s)", usos))
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("producto_id IN (SELECT id FROM producto_compra WHERE ingrediente_id = ?)", ing.ID).
			Delete(&models.HistorialPrecio{}).Error; err != nil {
			return err
		}
		if err := tx.Where("ingrediente_id = ?", ing.ID).Delete(&models.ProductoCompra{}).Error; err != nil {
			return err
		}
		if err := tx.Where("ingrediente_id = ?", ing.ID).Delete(&models.MedicionMerma{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Ingrediente{}, "id = ?", ing.ID).Error
	})
}

func (s *IngredienteService) AddProducto(ctx context.Context, userID, ingredienteID string, p *dto.ProductoInput) (*dto.IngredienteDto, error) {
	ing, err := s.load(ctx, userID, ingredienteID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	producto := models.ProductoCompra{
		IngredienteID: ing.ID,
		Marca:         p.Marca, Presentacion: p.Presentacion,
		Cantidad: p.Cantidad, Precio: p.Precio, Proveedor: p.Proveedor,
		Activo: len(ing.Productos) == 0, Orden: len(ing.Productos), PrecioActualizadoAt: now,
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&producto).Error; err != nil {
			return err
		}
		return tx.Create(&models.HistorialPrecio{ProductoID: producto.ID, Precio: p.Precio, Fecha: now}).Error
	})
	if err != nil {
		return nil, err
	}

	return s.reload(ctx, ing.ID)
}

// loadProducto fetches a product and guards through its ingredient's cocina.
func (s *IngredienteService) loadProducto(ctx context.Context, userID, productoID string) (*models.ProductoCompra, error) {
	var producto models.ProductoCompra
	err := s.db.WithContext(ctx).First(&producto, "id = ?", productoID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.NewNotFound("Producto", productoID)
	}
	if err != nil {
		return nil, err
	}
	if _, err := s.load(ctx, userID, producto.IngredienteID); err != nil {
		return nil, err
	}
	return &producto, nil
}

// UpdatePrecio — the most frequent flow of the app ("fui al súper y subió"):
// new price + history row + freshness timestamp. The frontend computes the
// impact screen locally (it still has the old catalog in memory).
func (s *IngredienteService) UpdatePrecio(ctx context.Context, userID, productoID string, d *dto.UpdatePrecioDto) (*dto.IngredienteDto, error) {
	producto, err := s.loadProducto(ctx, userID, productoID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&models.ProductoCompra{}).Where("id = ?", producto.ID).
			Updates(map[string]any{"precio": d.Precio, "precio_actualizado_at": now}).Error
		if err != nil {
			return err
		}
		return tx.Create(&models.HistorialPrecio{ProductoID: producto.ID, Precio: d.Precio, Fecha: now}).Error
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("Precio actualizado", "productoId", producto.ID, "precio", d.Precio)
	return s.reload(ctx, producto.IngredienteID)
}

func (s *IngredienteService) UpdateProducto(ctx context.Context, userID, productoID string, d *dto.UpdateProductoDto) (*dto.IngredienteDto, error) {
	producto, err := s.loadProducto(ctx, userID, productoID)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{}
	if d.Marca != nil {
		updates["marca"] = *d.Marca
	}
	if d.Presentacion != nil {
		updates["presentacion"] = *d.Presentacion
	}
	if d.Cantidad != nil {
		updates["cantidad"] = *d.Cantidad
	}
	if d.Proveedor != nil {
		updates["proveedor"] = *d.Proveedor
	}
	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&models.ProductoCompra{}).Where("id = ?", producto.ID).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return s.reload(ctx, producto.IngredienteID)
}

// ActivarProducto switches which product feeds the costs ("EN USO").
func (s *IngredienteService) ActivarProducto(ctx context.Context, userID, ingredienteID, productoID string) (*dto.IngredienteDto, error) {
	ing, err := s.load(ctx, userID, ingredienteID)
	if err != nil {
		return nil, err
	}

	found := false
	for _, p := range ing.Productos {
		if p.ID == productoID {
			found = true
			break
		}
	}
	if !found {
		return nil, apperrors.NewNotFound("Producto", productoID)
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.ProductoCompra{}).Where("ingrediente_id = ?", ing.ID).
			Update("activo", false).Error; err != nil {
			return err
		}
		return tx.Model(&models.ProductoCompra{}).Where("id = ?", productoID).
			Update("activo", true).Error
	})
	if err != nil {
		return nil, err
	}

	return s.reload(ctx, ing.ID)
}

// RegistrarCompra — sheet "Registrar compra": entran N presentaciones del
// producto ACTIVO al inventario; si el precio por unidad cambió, se actualiza
// el producto (+ historial + fecha de frescura).
func (s *IngredienteService) RegistrarCompra(ctx context.Context, userID, ingredienteID string, d *dto.RegistrarCompraDto) (*dto.IngredienteDto, error) {
	ing, err := s.load(ctx, userID, ingredienteID)
	if err != nil {
		return nil, err
	}

	activo := ProductoActivo(ing)
	if activo == nil {
		return nil, apperrors.NewBusinessRule("Este ingrediente no tiene producto de compra; agrégalo primero")
	}

	entra := float64(d.Unidades) * activo.Cantidad
	now := time.Now()

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&models.Ingrediente{}).Where("id = ?", ing.ID).
			Update("existencia", gorm.Expr("existencia + ?", entra)).Error
		if err != nil {
			return err
		}
		if d.Precio > 0 && d.Precio != activo.Precio {
			err := tx.Model(&models.ProductoCompra{}).Where("id = ?", activo.ID).
				Updates(map[string]any{"precio": d.Precio, "precio_actualizado_at": now}).Error
			if err != nil {
				return err
			}
			return tx.Create(&models.HistorialPrecio{ProductoID: activo.ID, Precio: d.Precio, Fecha: now}).Error
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("Compra registrada", "ingredienteId", ing.ID, "entra", entra)
	return s.reload(ctx, ing.ID)
}

// SetExistencia — ajuste directo de inventario (existencia, mínimo, caducidad).
func (s *IngredienteService) SetExistencia(ctx context.Context, userID, ingredienteID string, d *dto.ExistenciaDto) (*dto.IngredienteDto, error) {
	ing, err := s.load(ctx, userID, ingredienteID)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{}
	if d.Existencia != nil {
		updates["existencia"] = *d.Existencia
	}
	if d.Minimo != nil {
		updates["minimo"] = *d.Minimo
	}
	if d.CaducaAt != nil {
		if *d.CaducaAt == "" {
			updates["caduca_at"] = nil
		} else {
			fecha, err := time.Parse("2006-01-02", *d.CaducaAt)
			if err != nil {
				return nil, apperrors.NewValidation("Validation failed", []map[string]string{
					{"field": "caducaAt", "message": "usa el formato YYYY-MM-DD"},
				})
			}
			updates["caduca_at"] = fecha
		}
	}
	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&models.Ingrediente{}).Where("id = ?", ing.ID).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return s.reload(ctx, ing.ID)
}

// Conteo — "Contar mi cocina": aplica en bloque las existencias contadas.
// Devuelve los ingredientes actualizados.
func (s *IngredienteService) Conteo(ctx context.Context, userID, cocinaID string, d *dto.ConteoDto) ([]dto.IngredienteDto, error) {
	if err := s.catalogo.RequireEditor(ctx, userID, cocinaID); err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(d.Items))
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range d.Items {
			res := tx.Model(&models.Ingrediente{}).
				Where("id = ? AND cocina_id = ?", item.IngredienteID, cocinaID).
				Update("existencia", item.Cantidad)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return apperrors.NewNotFound("Ingrediente", item.IngredienteID)
			}
			ids = append(ids, item.IngredienteID)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("Conteo aplicado", "cocinaId", cocinaID, "items", len(ids))
	return s.reloadVarios(ctx, ids)
}

func (s *IngredienteService) reloadVarios(ctx context.Context, ids []string) ([]dto.IngredienteDto, error) {
	var ings []models.Ingrediente
	err := s.db.WithContext(ctx).
		Preload("Productos", func(db *gorm.DB) *gorm.DB { return db.Order("orden ASC, created_at ASC") }).
		Preload("Productos.Historial", func(db *gorm.DB) *gorm.DB { return db.Order("fecha ASC") }).
		Where("id IN ?", ids).
		Find(&ings).Error
	if err != nil {
		return nil, err
	}
	out := make([]dto.IngredienteDto, 0, len(ings))
	for i := range ings {
		out = append(out, toIngredienteDto(&ings[i]))
	}
	return out, nil
}

// SetMerma stores a manual/reference merma directly.
func (s *IngredienteService) SetMerma(ctx context.Context, userID, ingredienteID string, d *dto.SetMermaDto) (*dto.IngredienteDto, error) {
	ing, err := s.load(ctx, userID, ingredienteID)
	if err != nil {
		return nil, err
	}

	err = s.db.WithContext(ctx).Model(&models.Ingrediente{}).Where("id = ?", ing.ID).
		Updates(map[string]any{"merma_pct": d.Pct, "merma_origen": d.Origen}).Error
	if err != nil {
		return nil, err
	}

	return s.reload(ctx, ing.ID)
}

// AddMedicion — the merma wizard: weights in, percent out (origen "medido").
// merma% = round((1 − (limpio + aprovechado) / entero) × 100), clamp 0–95.
func (s *IngredienteService) AddMedicion(ctx context.Context, userID, ingredienteID string, d *dto.MedicionInput) (*dto.IngredienteDto, error) {
	ing, err := s.load(ctx, userID, ingredienteID)
	if err != nil {
		return nil, err
	}

	if d.PesoLimpio+d.Aprovechado > d.PesoEntero {
		return nil, apperrors.NewBusinessRule("El peso limpio más lo aprovechado no puede superar el peso entero")
	}

	pct := math.Round((1 - (d.PesoLimpio+d.Aprovechado)/d.PesoEntero) * 100)
	pct = math.Max(0, math.Min(95, pct))

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		medicion := models.MedicionMerma{
			IngredienteID: ing.ID,
			PesoEntero:    d.PesoEntero, PesoLimpio: d.PesoLimpio, Aprovechado: d.Aprovechado,
			PctResultante: pct,
		}
		if err := tx.Create(&medicion).Error; err != nil {
			return err
		}
		return tx.Model(&models.Ingrediente{}).Where("id = ?", ing.ID).
			Updates(map[string]any{"merma_pct": pct, "merma_origen": models.MermaMedido}).Error
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("Merma medida", "ingredienteId", ing.ID, "pct", pct)
	return s.reload(ctx, ing.ID)
}

func (s *IngredienteService) reload(ctx context.Context, ingredienteID string) (*dto.IngredienteDto, error) {
	var ing models.Ingrediente
	err := s.db.WithContext(ctx).
		Preload("Productos", func(db *gorm.DB) *gorm.DB { return db.Order("orden ASC, created_at ASC") }).
		Preload("Productos.Historial", func(db *gorm.DB) *gorm.DB { return db.Order("fecha ASC") }).
		First(&ing, "id = ?", ingredienteID).Error
	if err != nil {
		return nil, err
	}
	out := toIngredienteDto(&ing)
	return &out, nil
}
