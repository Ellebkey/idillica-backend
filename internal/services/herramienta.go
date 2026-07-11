// herramienta.go — equipo de la cocina (alta desde el sheet de Inventario).
package services

import (
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"

	"idilica-backend-go/internal/apperrors"
	"idilica-backend-go/internal/dto"
	"idilica-backend-go/internal/models"
)

type HerramientaService struct {
	db       *gorm.DB
	catalogo *CatalogoService
	logger   *slog.Logger
}

func NewHerramientaService(db *gorm.DB, catalogoService *CatalogoService, logger *slog.Logger) *HerramientaService {
	return &HerramientaService{db: db, catalogo: catalogoService, logger: logger}
}

func (s *HerramientaService) Create(ctx context.Context, userID, cocinaID string, d *dto.HerramientaInput) (*dto.HerramientaDto, error) {
	if err := s.catalogo.RequireEditor(ctx, userID, cocinaID); err != nil {
		return nil, err
	}

	h := models.Herramienta{
		CocinaID: cocinaID,
		Nombre:   d.Nombre,
		Detalle:  d.Detalle,
		Estado:   d.Estado,
	}
	if h.Estado == "" {
		h.Estado = "Buen estado"
	}
	if err := s.db.WithContext(ctx).Create(&h).Error; err != nil {
		return nil, err
	}

	s.logger.Info("Herramienta creada", "herramientaId", h.ID, "cocinaId", cocinaID)
	out := toHerramientaDto(&h)
	return &out, nil
}

func (s *HerramientaService) Update(ctx context.Context, userID, herramientaID string, d *dto.HerramientaInput) (*dto.HerramientaDto, error) {
	h, err := s.load(ctx, userID, herramientaID)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{"nombre": d.Nombre, "detalle": d.Detalle}
	if d.Estado != "" {
		updates["estado"] = d.Estado
	}
	if err := s.db.WithContext(ctx).Model(&models.Herramienta{}).Where("id = ?", h.ID).Updates(updates).Error; err != nil {
		return nil, err
	}

	var actualizada models.Herramienta
	if err := s.db.WithContext(ctx).First(&actualizada, "id = ?", h.ID).Error; err != nil {
		return nil, err
	}
	out := toHerramientaDto(&actualizada)
	return &out, nil
}

func (s *HerramientaService) Delete(ctx context.Context, userID, herramientaID string) error {
	h, err := s.load(ctx, userID, herramientaID)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Delete(&models.Herramienta{}, "id = ?", h.ID).Error
}

func (s *HerramientaService) load(ctx context.Context, userID, herramientaID string) (*models.Herramienta, error) {
	var h models.Herramienta
	err := s.db.WithContext(ctx).First(&h, "id = ?", herramientaID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.NewNotFound("Herramienta", herramientaID)
	}
	if err != nil {
		return nil, err
	}
	if err := s.catalogo.RequireEditor(ctx, userID, h.CocinaID); err != nil {
		return nil, err
	}
	return &h, nil
}
