// cocina.go — workspace multi-tenant y membresías.
package services

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"idilica-backend-go/internal/apperrors"
	"idilica-backend-go/internal/dto"
	"idilica-backend-go/internal/models"
)

type CocinaService struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewCocinaService(db *gorm.DB, logger *slog.Logger) *CocinaService {
	return &CocinaService{db: db, logger: logger}
}

// Create crea la cocina + la membresía owner, atómico.
func (s *CocinaService) Create(ctx context.Context, userID string, d *dto.CreateCocinaDto) (*dto.CocinaDto, error) {
	cocina := models.Cocina{
		Name:             d.Name,
		Moneda:           "MXN",
		ImpuestoDefault:  0.16,
		FoodCostObjetivo: 0.30,
	}
	if d.Moneda != "" {
		cocina.Moneda = strings.ToUpper(d.Moneda)
	}
	if d.ImpuestoDefault != nil {
		cocina.ImpuestoDefault = *d.ImpuestoDefault
	}
	if d.FoodCostObjetivo != nil {
		cocina.FoodCostObjetivo = *d.FoodCostObjetivo
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&cocina).Error; err != nil {
			return err
		}
		member := models.CocinaMember{CocinaID: cocina.ID, UserID: userID, Rol: models.RolOwner}
		return tx.Create(&member).Error
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("Cocina created", "cocinaId", cocina.ID, "userId", userID)
	return toCocinaDto(&cocina, models.RolOwner), nil
}

// FindAllForUser lista las cocinas del usuario con su rol en cada una.
func (s *CocinaService) FindAllForUser(ctx context.Context, userID string) ([]dto.CocinaDto, error) {
	var memberships []models.CocinaMember
	err := s.db.WithContext(ctx).
		Preload("Cocina").
		Where("user_id = ?", userID).
		Find(&memberships).Error
	if err != nil {
		return nil, err
	}

	cocinas := make([]dto.CocinaDto, 0, len(memberships))
	for i := range memberships {
		if memberships[i].Cocina != nil {
			cocinas = append(cocinas, *toCocinaDto(memberships[i].Cocina, memberships[i].Rol))
		}
	}

	// Ordenadas por nombre
	sort.Slice(cocinas, func(i, j int) bool { return cocinas[i].Name < cocinas[j].Name })
	return cocinas, nil
}

// FindByID — la membresía es la guarda de tenencia.
func (s *CocinaService) FindByID(ctx context.Context, userID, cocinaID string) (*dto.CocinaDto, error) {
	member, err := s.RequireMembership(ctx, userID, cocinaID)
	if err != nil {
		return nil, err
	}
	return toCocinaDto(member.Cocina, member.Rol), nil
}

// Update: los viewers no editan; update parcial con los campos no nulos.
func (s *CocinaService) Update(ctx context.Context, userID, cocinaID string, d *dto.UpdateCocinaDto) (*dto.CocinaDto, error) {
	member, err := s.RequireMembership(ctx, userID, cocinaID)
	if err != nil {
		return nil, err
	}

	if member.Rol == models.RolViewer {
		return nil, apperrors.NewForbidden("Access denied: viewers cannot edit the cocina")
	}

	updates := map[string]any{}
	if d.Name != nil {
		updates["name"] = *d.Name
	}
	if d.Moneda != nil {
		updates["moneda"] = strings.ToUpper(*d.Moneda)
	}
	if d.ImpuestoDefault != nil {
		updates["impuesto_default"] = *d.ImpuestoDefault
	}
	if d.FoodCostObjetivo != nil {
		updates["food_cost_objetivo"] = *d.FoodCostObjetivo
	}
	if d.GastoSueldos != nil {
		updates["gasto_sueldos"] = *d.GastoSueldos
	}
	if d.GastoGas != nil {
		updates["gasto_gas"] = *d.GastoGas
	}
	if d.GastoLuz != nil {
		updates["gasto_luz"] = *d.GastoLuz
	}
	if d.GastoEquipo != nil {
		updates["gasto_equipo"] = *d.GastoEquipo
	}
	if d.ComprasIngredientesMes != nil {
		updates["compras_ingredientes_mes"] = *d.ComprasIngredientesMes
	}
	if len(updates) == 0 {
		// se requiere al menos un campo
		return nil, apperrors.NewValidation("Validation failed", []map[string]string{
			{"field": "body", "message": "at least one field is required"},
		})
	}

	if err := s.db.WithContext(ctx).Model(&models.Cocina{}).Where("id = ?", cocinaID).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Re-read so the response reflects exactly what is stored
	var cocina models.Cocina
	if err := s.db.WithContext(ctx).First(&cocina, "id = ?", cocinaID).Error; err != nil {
		return nil, err
	}

	s.logger.Info("Cocina updated", "cocinaId", cocinaID, "userId", userID)
	return toCocinaDto(&cocina, member.Rol), nil
}

// RequireMembership carga la membresía del usuario (con su cocina) o regresa
// NOT_FOUND — todo acceso a datos de una cocina pasa por aquí.
func (s *CocinaService) RequireMembership(ctx context.Context, userID, cocinaID string) (*models.CocinaMember, error) {
	var member models.CocinaMember
	err := s.db.WithContext(ctx).
		Preload("Cocina").
		First(&member, "cocina_id = ? AND user_id = ?", cocinaID, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && member.Cocina == nil) {
		return nil, apperrors.NewNotFound("Cocina", cocinaID)
	}
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func toCocinaDto(cocina *models.Cocina, rol string) *dto.CocinaDto {
	return &dto.CocinaDto{
		ID:               cocina.ID,
		Name:             cocina.Name,
		Moneda:           cocina.Moneda,
		ImpuestoDefault:  cocina.ImpuestoDefault,
		FoodCostObjetivo: cocina.FoodCostObjetivo,

		GastoSueldos:           cocina.GastoSueldos,
		GastoGas:               cocina.GastoGas,
		GastoLuz:               cocina.GastoLuz,
		GastoEquipo:            cocina.GastoEquipo,
		ComprasIngredientesMes: cocina.ComprasIngredientesMes,

		Rol: rol,
		// timestamps en UTC con sufijo Z
		CreatedAt:        cocina.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:        cocina.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
