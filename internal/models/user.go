// Package models mirrors src/models (Sequelize) using GORM.
// user.go ≈ user.model.ts — same table ("user") and column names.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Roles is stored as a JSON column, like DataTypes.JSON in Sequelize.
// Implementing driver.Valuer and sql.Scanner is the Go way to map a custom
// type to a database column (no ORM magic — explicit and testable).
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

// User ≈ UserInstance. Users join one or more cocinas through CocinaMember.
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

	// Associations (≈ UserInstance.associate)
	Memberships []CocinaMember `gorm:"foreignKey:UserID"`
}

// TableName keeps the exact table name of the Node backend.
func (User) TableName() string { return "user" }
