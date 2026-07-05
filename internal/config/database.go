// database.go ≈ sequelize.ts: GORM plays the role of Sequelize.
// AutoMigrate ≈ sequelize.sync(); the NamingStrategy mirrors the Node models
// (singular snake_case table names: user, cocina, cocina_member).
package config

import (
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"idilica-backend-go/internal/models"
)

// NewDatabase opens the PostgreSQL connection, ensures the uuid extension,
// runs AutoMigrate and configures the pool (MAX_POOL/MIN_POOL of the Node app).
func NewDatabase(cfg *Config, logger *slog.Logger) (*gorm.DB, error) {
	logger.Info("Initializing PostgreSQL Database")

	sslMode := "disable"
	if cfg.IsProduction() {
		sslMode = "require"
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		cfg.SQLHost, cfg.SQLUser, cfg.SQLPassword, cfg.SQLDB, cfg.SQLPort, sslMode,
	)

	// Query logging: verbose in development (mirror of `logging: logger.verbose`),
	// quiet elsewhere.
	gormLogLevel := gormlogger.Silent
	if cfg.IsDevelopment() {
		gormLogLevel = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormLogLevel),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // table "user", not "users" — same as the Node models
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to connect to the database: %w", err)
	}

	// uuid_generate_v4() is the PK default on every model (docker/init.sql also
	// creates it; this is a safety net for non-docker databases).
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		logger.Warn("could not ensure uuid-ossp extension", "error", err)
	}

	// ≈ sequelize.sync()
	err = db.AutoMigrate(
		&models.User{}, &models.Cocina{}, &models.CocinaMember{},
		&models.Ingrediente{}, &models.ProductoCompra{}, &models.HistorialPrecio{},
		&models.MedicionMerma{}, &models.Receta{}, &models.RecetaLinea{},
	)
	if err != nil {
		return nil, fmt.Errorf("database synchronization failed: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxPool)
	sqlDB.SetMaxIdleConns(cfg.MinPool)
	sqlDB.SetConnMaxIdleTime(10 * time.Second)

	logger.Info("Connection has been established successfully.")
	logger.Info("PostgreSQL Database synchronized")
	return db, nil
}
