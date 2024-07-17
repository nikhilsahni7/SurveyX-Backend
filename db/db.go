package db

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/nikhilsahni7/SurveyX/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
    DB    *gorm.DB
    SqlDB *sql.DB
)

func InitDB() {
    // Load environment variables from .env file
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using environment variables")
    }

    // Connect to the database
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        log.Fatal("DATABASE_URL environment variable is not set")
    }

    // Configure GORM logger
    gormLogger := logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags),
        logger.Config{
            SlowThreshold: time.Second, // Increase slow SQL threshold
            LogLevel:      logger.Warn, // Only log warnings and errors
            Colorful:      false,
        },
    )

    // Open database connection
    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: gormLogger,
        PrepareStmt: true, // Enables query caching
    })
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
    SqlDB.SetConnMaxLifetime(time.Hour)

    // Auto Migrate the schema
    if err := migrateSchema(); err != nil {
        log.Fatalf("Failed to migrate database: %v", err)
    }

    log.Println("Database connected and migrated successfully")
}

func migrateSchema() error {
    return DB.AutoMigrate(
        &models.User{},
        &models.Team{},
        &models.Survey{},
        &models.Question{},
        &models.Condition{},
        &models.Option{},
        &models.Response{},
        &models.Answer{},
        &models.SurveyLink{},
        &models.Webhook{},
    )
}

func GetDB() *gorm.DB {
    return DB
}

func GetSqlDB() *sql.DB {
    return SqlDB
}
