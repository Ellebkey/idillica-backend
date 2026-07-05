// cocina_member.go ≈ cocina-member.model.ts — membership User↔Cocina with role.
package models

import "time"

// Cocina roles (≈ ENUM('owner','editor','viewer'); stored as varchar here).
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

	// Associations (≈ belongsTo in the Node model)
	Cocina *Cocina `gorm:"foreignKey:CocinaID"`
	User   *User   `gorm:"foreignKey:UserID"`
}

func (CocinaMember) TableName() string { return "cocina_member" }
