// receta.go — recipes: full-save from the editor (replace lines/steps/
// allergens) with referential and CYCLE validation over the sub-recipe graph.
package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"idilica-backend-go/internal/apperrors"
	"idilica-backend-go/internal/dto"
	"idilica-backend-go/internal/models"
)

type RecetaService struct {
	db       *gorm.DB
	catalogo *CatalogoService
	logger   *slog.Logger
}

func NewRecetaService(db *gorm.DB, catalogoService *CatalogoService, logger *slog.Logger) *RecetaService {
	return &RecetaService{db: db, catalogo: catalogoService, logger: logger}
}

func (s *RecetaService) Create(ctx context.Context, userID, cocinaID string, d *dto.SaveRecetaDto) (*dto.RecetaDto, error) {
	if err := s.catalogo.RequireEditor(ctx, userID, cocinaID); err != nil {
		return nil, err
	}
	return s.save(ctx, cocinaID, "", d)
}

func (s *RecetaService) Update(ctx context.Context, userID, recetaID string, d *dto.SaveRecetaDto) (*dto.RecetaDto, error) {
	receta, err := s.loadReceta(ctx, recetaID)
	if err != nil {
		return nil, err
	}
	if err := s.catalogo.RequireEditor(ctx, userID, receta.CocinaID); err != nil {
		return nil, err
	}
	return s.save(ctx, receta.CocinaID, receta.ID, d)
}

// Delete refuses when the recipe is used as a sub-recipe elsewhere.
func (s *RecetaService) Delete(ctx context.Context, userID, recetaID string) error {
	receta, err := s.loadReceta(ctx, recetaID)
	if err != nil {
		return err
	}
	if err := s.catalogo.RequireEditor(ctx, userID, receta.CocinaID); err != nil {
		return err
	}

	var usos int64
	if err := s.db.WithContext(ctx).Model(&models.RecetaLinea{}).
		Where("sub_receta_id = ?", receta.ID).Count(&usos).Error; err != nil {
		return err
	}
	if usos > 0 {
		return apperrors.NewBusinessRule(
			fmt.Sprintf("No se puede eliminar: esta receta se usa como subreceta en %d receta(s)", usos))
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("receta_id = ?", receta.ID).Delete(&models.RecetaLinea{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Receta{}, "id = ?", receta.ID).Error
	})
}

// Producir — "Produje esta receta": calcula las necesidades RECURSIVAS
// (incluye subrecetas) y las descuenta del inventario (clamp a 0).
// Devuelve los ingredientes afectados ya actualizados.
func (s *RecetaService) Producir(ctx context.Context, userID, recetaID string) ([]dto.IngredienteDto, error) {
	receta, err := s.loadReceta(ctx, recetaID)
	if err != nil {
		return nil, err
	}
	if err := s.catalogo.RequireEditor(ctx, userID, receta.CocinaID); err != nil {
		return nil, err
	}

	grafo, err := s.buildGrafo(ctx, receta.CocinaID)
	if err != nil {
		return nil, err
	}
	needs := grafo.GatherNeeds(recetaID)
	if len(needs) == 0 {
		return nil, apperrors.NewBusinessRule("Esta receta no tiene ingredientes que descontar")
	}

	ids := make([]string, 0, len(needs))
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for ingredienteID, cantidad := range needs {
			err := tx.Model(&models.Ingrediente{}).
				Where("id = ? AND cocina_id = ?", ingredienteID, receta.CocinaID).
				Update("existencia", gorm.Expr("GREATEST(existencia - ?, 0)", cantidad)).Error
			if err != nil {
				return err
			}
			ids = append(ids, ingredienteID)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("Producción descontada", "recetaId", recetaID, "ingredientes", len(ids))

	var ings []models.Ingrediente
	err = s.db.WithContext(ctx).
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

// save validates lines (XOR, same-cocina references, no cycles) and writes
// the recipe with a full line replace — the editor's "Guardar".
func (s *RecetaService) save(ctx context.Context, cocinaID, recetaID string, d *dto.SaveRecetaDto) (*dto.RecetaDto, error) {
	if err := s.validateLineas(ctx, cocinaID, recetaID, d.Lineas); err != nil {
		return nil, err
	}

	receta := models.Receta{
		CocinaID: cocinaID, Nombre: d.Nombre, Categoria: d.Categoria,
		Porciones: d.Porciones, Etiqueta: defaultStr(d.Etiqueta, "piezas"),
		EtiquetaSingular: defaultStr(d.EtiquetaSingular, "pieza"),
		RendimientoKg:    d.RendimientoKg, PrecioVenta: d.PrecioVenta,
		IvaPct: d.IvaPct, EsSubreceta: d.EsSubreceta,
		Alergenos: emptyIfNil(d.Alergenos), Pasos: emptyIfNil(d.Pasos), Fotos: emptyIfNil(d.Fotos),
	}
	if receta.IvaPct == 0 {
		receta.IvaPct = 16
	}
	// Las subrecetas no tienen precio de venta (se venden dentro de otras)
	if receta.EsSubreceta {
		receta.PrecioVenta = nil
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if recetaID == "" {
			if err := tx.Create(&receta).Error; err != nil {
				return err
			}
		} else {
			receta.ID = recetaID
			err := tx.Model(&models.Receta{}).Where("id = ?", recetaID).
				Select("nombre", "categoria", "porciones", "etiqueta", "etiqueta_singular",
					"rendimiento_kg", "precio_venta", "iva_pct", "es_subreceta",
					"alergenos", "pasos", "fotos").
				Updates(&receta).Error
			if err != nil {
				return err
			}
			if err := tx.Where("receta_id = ?", recetaID).Delete(&models.RecetaLinea{}).Error; err != nil {
				return err
			}
		}

		for i, l := range d.Lineas {
			linea := models.RecetaLinea{
				RecetaID: receta.ID, Orden: i, Cantidad: l.Cantidad,
				IngredienteID: l.IngredienteID, SubRecetaID: l.RecetaID,
			}
			if err := tx.Create(&linea).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("Receta guardada", "recetaId", receta.ID, "cocinaId", cocinaID)

	var saved models.Receta
	err = s.db.WithContext(ctx).
		Preload("Lineas", func(db *gorm.DB) *gorm.DB { return db.Order("orden ASC") }).
		First(&saved, "id = ?", receta.ID).Error
	if err != nil {
		return nil, err
	}
	out := toRecetaDto(&saved)
	return &out, nil
}

// validateLineas: each line points to exactly ONE of ingrediente/receta, the
// reference exists in the same cocina, and sub-recipe lines create no cycle.
func (s *RecetaService) validateLineas(ctx context.Context, cocinaID, recetaID string, lineas []dto.LineaInput) error {
	subIDs := []string{}
	for i, l := range lineas {
		hasIng := l.IngredienteID != nil
		hasRec := l.RecetaID != nil
		if hasIng == hasRec { // both or neither
			return apperrors.NewValidation("Validation failed", []map[string]string{{
				"field":   fmt.Sprintf("lineas[%d]", i),
				"message": "cada línea debe apuntar a un ingrediente O a una receta",
			}})
		}

		if hasIng {
			var count int64
			if err := s.db.WithContext(ctx).Model(&models.Ingrediente{}).
				Where("id = ? AND cocina_id = ?", *l.IngredienteID, cocinaID).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				return apperrors.NewNotFound("Ingrediente", *l.IngredienteID)
			}
		} else {
			if recetaID != "" && *l.RecetaID == recetaID {
				return apperrors.NewBusinessRule("Una receta no puede contenerse a sí misma")
			}
			var count int64
			if err := s.db.WithContext(ctx).Model(&models.Receta{}).
				Where("id = ? AND cocina_id = ?", *l.RecetaID, cocinaID).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				return apperrors.NewNotFound("Receta", *l.RecetaID)
			}
			subIDs = append(subIDs, *l.RecetaID)
		}
	}

	// Cycle check over the whole sub-recipe graph of the cocina
	if recetaID != "" && len(subIDs) > 0 {
		grafo, err := s.buildGrafo(ctx, cocinaID)
		if err != nil {
			return err
		}
		for _, subID := range subIDs {
			if grafo.UsaReceta(subID, recetaID) {
				sub := grafo.Recetas[subID]
				nombre := subID
				if sub != nil {
					nombre = sub.Nombre
				}
				return apperrors.NewBusinessRule(
					fmt.Sprintf("Agregar %q crearía un ciclo: esa receta ya contiene a esta", nombre))
			}
		}
	}
	return nil
}

func (s *RecetaService) buildGrafo(ctx context.Context, cocinaID string) (*Catalogo, error) {
	var recetas []models.Receta
	err := s.db.WithContext(ctx).Preload("Lineas").Where("cocina_id = ?", cocinaID).Find(&recetas).Error
	if err != nil {
		return nil, err
	}
	grafo := &Catalogo{Recetas: map[string]*models.Receta{}, Ingredientes: map[string]*models.Ingrediente{}}
	for i := range recetas {
		grafo.Recetas[recetas[i].ID] = &recetas[i]
	}
	return grafo, nil
}

func (s *RecetaService) loadReceta(ctx context.Context, recetaID string) (*models.Receta, error) {
	var receta models.Receta
	err := s.db.WithContext(ctx).First(&receta, "id = ?", recetaID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.NewNotFound("Receta", recetaID)
	}
	if err != nil {
		return nil, err
	}
	return &receta, nil
}

func defaultStr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func emptyIfNil(v []string) models.StringSlice {
	if v == nil {
		return models.StringSlice{}
	}
	return models.StringSlice(v)
}
