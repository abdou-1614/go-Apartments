package storage

import (
	"go-appointement/model"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func connectToDb() *gorm.DB {
	err := godotenv.Load()

	if err != nil {
		panic("Error to get env")
	}

	dsn := os.Getenv("CONNECT_DB")

	db, dbErro := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if dbErro != nil {
		log.Panic("ERROR TO CONNECT TO DB")
	}

	DB = db

	return db
}

func performMigration(db *gorm.DB) {
	db.AutoMigrate(
		&model.User{},
		&model.Appartements{},
	)
}

func InitializeDb() *gorm.DB {
	db := connectToDb()
	performMigration(db)

	return db
}
