package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/nikhilsahni7/SurveyX/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB    *gorm.DB
	SqlDB *sql.DB
)

func InitDB() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to the database
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get the underlying *sql.DB instance
	SqlDB, err = DB.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	// Set connection pool settings
	SqlDB.SetMaxIdleConns(10)
	SqlDB.SetMaxOpenConns(100)

	// Auto Migrate the schema
	err = DB.AutoMigrate(
		&models.User{},
		&models.Survey{},
		&models.Question{},
		&models.Option{},
		&models.Response{},
		&models.Answer{},
		&models.SurveyLink{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database connected and migrated successfully")
}

func GetDB() *gorm.DB {
	return DB
}

func GetSqlDB() *sql.DB {
	return SqlDB
}
