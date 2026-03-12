package database

import (
	"4-order-api/config"
	"4-order-api/models"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.Db.Dsn

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connected successfully")
	return db, nil
}

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Session{},
		&models.Order{},
	)

	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("Migration completed successfully")
	return nil
}
