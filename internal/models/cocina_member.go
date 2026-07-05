// cocina_member.go — membresía User↔Cocina con rol.
package models

import "time"

// Roles de cocina (guardados como varchar).
const (
	RolOwner  = "owner"
	RolEditor = "editor"
	RolViewer = "viewer"
)

type CocinaMember struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CocinaID  string    `gorm:"column:cocina_id;type:uuid;not null;uniqueIndex:idx_cocina_member_cocina_user"`
	UserID    string    `gorm:"column:user_id;type:uuid;not null;uniqueIndex:idx_cocina_member_cocina_user"`
	Rol       string    `gorm:"size:10;not null;default:viewer"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`

	// Asociaciones
	Cocina *Cocina `gorm:"foreignKey:CocinaID"`
	User   *User   `gorm:"foreignKey:UserID"`
}

func (CocinaMember) TableName() string { return "cocina_member" }
