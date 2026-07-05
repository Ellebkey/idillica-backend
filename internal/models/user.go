// Package models define las entidades GORM del dominio.
// user.go — cuenta de usuario (tabla "user").
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Roles se guarda como columna JSON. Implementar driver.Valuer y sql.Scanner
// es la forma estándar de mapear un tipo propio a una columna.
type Roles []string

// Value serializes the slice to JSON when writing to the DB.
func (r Roles) Value() (driver.Value, error) {
	if r == nil {
		return "[]", nil
	}
	b, err := json.Marshal(r)
	return string(b), err
}

// Scan deserializes the JSON column when reading from the DB.
func (r *Roles) Scan(value any) error {
	if value == nil {
		*r = Roles{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("roles: unsupported column type %T", value)
	}
	if len(bytes) == 0 {
		*r = Roles{}
		return nil
	}
	return json.Unmarshal(bytes, r)
}

// User — los usuarios pertenecen a una o más cocinas vía CocinaMember.
type User struct {
	ID             string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Username       string    `gorm:"uniqueIndex;not null"`
	Fullname       *string   `gorm:"size:100"`
	HashedPassword string    `gorm:"column:hashed_password;not null"`
	Email          string    `gorm:"not null"`
	EmailVerified  bool      `gorm:"not null;default:false"`
	Roles          Roles     `gorm:"type:json"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`

	// Asociaciones
	Memberships []CocinaMember `gorm:"foreignKey:UserID"`
}

// TableName fija el nombre de tabla en singular.
func (User) TableName() string { return "user" }
