package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"backend/internal/models"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var SqlDB *sql.DB

func Connect() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	// Connect with database/sql for raw queries if needed
	var err error
	SqlDB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database with database/sql:", err)
	}

	// Test connection
	if err := SqlDB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Connect with GORM using separate connection (not shared)
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.Fatal("Failed to connect to database with GORM:", err)
	}

	log.Println("Database connected successfully")
}

func Close() {
	// Close GORM connection
	if DB != nil {
		if sqlDB, err := DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Error closing GORM connection: %v", err)
			}
		}
	}

	// Close raw SQL connection
	if SqlDB != nil {
		if err := SqlDB.Close(); err != nil {
			log.Printf("Error closing SQL connection: %v", err)
		}
	}
}

func Migrate() {
	log.Println("Starting database migration...")

	// Check if migration is needed to avoid redundant operations
	if !needsMigration() {
		log.Println("Database schema is up to date")
		// Still add constraints and indexes in case they're missing
		if err := addConstraintsAndIndexes(); err != nil {
			log.Printf("Warning: Failed to add some constraints/indexes: %v", err)
		}
		log.Println("Database migration completed successfully")
		return
	}

	// AutoMigrate will create tables, missing columns, missing indexes
	// but it WON'T delete unused columns or indexes
	err := DB.AutoMigrate(
		&models.User{},
		&models.Package{},
		&models.Business{},
		&models.Student{},
		&models.Teacher{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Add constraints and indexes after migration
	if err := addConstraintsAndIndexes(); err != nil {
		log.Printf("Warning: Failed to add some constraints/indexes: %v", err)
	}

	log.Println("Database migration completed successfully")
}

func needsMigration() bool {
	// Check if the users table exists with expected columns
	var count int64

	// First check if table exists
	err := DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ? AND table_schema = CURRENT_SCHEMA()", "users").Scan(&count)
	if err != nil {
		log.Printf("Error checking for existing tables: %v", err)
		return true // Run migration if we can't check
	}

	if count == 0 {
		log.Println("Users table does not exist, migration needed")
		return true
	}

	// Check if all expected columns exist
	expectedColumns := []string{"id", "name", "email", "phone", "password", "role", "status", "created_on", "updated_on"}
	var existingColumnCount int64

	err = DB.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.columns 
		WHERE table_name = ? 
		AND table_schema = CURRENT_SCHEMA() 
		AND column_name = ANY($1)
	`, "users", expectedColumns).Scan(&existingColumnCount)

	if err != nil {
		log.Printf("Error checking columns: %v", err)
		return true
	}

	if int(existingColumnCount) != len(expectedColumns) {
		log.Printf("Expected %d columns, found %d. Migration needed", len(expectedColumns), existingColumnCount)
		return true
	}

	log.Println("All expected columns exist")
	return false
}

func addConstraintsAndIndexes() error {
	log.Println("Adding constraints and indexes...")

	// Add check constraint for role validation
	err := DB.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.check_constraints 
				WHERE constraint_name = 'users_role_check' 
				AND constraint_schema = CURRENT_SCHEMA()
			) THEN
				ALTER TABLE users ADD CONSTRAINT users_role_check 
				CHECK (role IN ('admin', 'business', 'teacher', 'student'));
			END IF;
		END $$;
	`).Error

	if err != nil {
		log.Printf("Warning: Failed to add role check constraint: %v", err)
	} else {
		log.Println("Role constraint added/verified successfully")
	}

	// Add check constraint for status validation
	err = DB.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.check_constraints 
				WHERE constraint_name = 'users_status_check'
				AND constraint_schema = CURRENT_SCHEMA()
			) THEN
				ALTER TABLE users ADD CONSTRAINT users_status_check 
				CHECK (status IN (0, 1));
			END IF;
		END $$;
	`).Error

	if err != nil {
		log.Printf("Warning: Failed to add status check constraint: %v", err)
	} else {
		log.Println("Status constraint added/verified successfully")
	}

	// Create indexes for better performance
	indexes := map[string]string{
		"idx_users_email":      "CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"idx_users_role":       "CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)",
		"idx_users_status":     "CREATE INDEX IF NOT EXISTS idx_users_status ON users(status)",
		"idx_users_created_on": "CREATE INDEX IF NOT EXISTS idx_users_created_on ON users(created_on)",
	}

	for indexName, indexSQL := range indexes {
		if err := DB.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create index %s: %v", indexName, err)
		} else {
			log.Printf("Index %s added/verified successfully", indexName)
		}
	}

	return nil
}

// Helper function to get database connection info
func GetConnectionInfo() map[string]string {
	return map[string]string{
		"host":    os.Getenv("DB_HOST"),
		"user":    os.Getenv("DB_USER"),
		"dbname":  os.Getenv("DB_NAME"),
		"port":    os.Getenv("DB_PORT"),
		"sslmode": "disable",
	}
}

// Helper function to check database connection health
func HealthCheck() error {
	// Check GORM connection
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return fmt.Errorf("failed to get underlying sql.DB from GORM: %v", err)
		}
		if err := sqlDB.Ping(); err != nil {
			return fmt.Errorf("GORM connection failed: %v", err)
		}
	} else {
		return fmt.Errorf("GORM DB is nil")
	}

	// Check raw SQL connection
	if SqlDB != nil {
		if err := SqlDB.Ping(); err != nil {
			return fmt.Errorf("SQL connection failed: %v", err)
		}
	} else {
		return fmt.Errorf("SQL DB is nil")
	}

	return nil
}
