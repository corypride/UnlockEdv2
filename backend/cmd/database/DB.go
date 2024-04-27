package database

import (
	"backend/cmd/models"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	Conn *gorm.DB
}

func InitDB(isTesting bool) *DB {
	var db *gorm.DB
	var err error

	if isTesting {
		db, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect to SQLite database:", err)
		}
		log.Println("Connected to the SQLite database in memory")
	} else {
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
		db, err = gorm.Open(postgres.New(postgres.Config{
			DSN: dsn,
		}), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL database: %v", err)
		}
		log.Println("Connected to the PostgreSQL database")
	}
	database := &DB{Conn: db}
	if isTesting {
		database.MigrateFresh(isTesting)
		database.SeedTestData()
	}
	return database
}

/**
* Register Migrations here
**/
func (db *DB) Migrate(isTesting bool) {
	if !isTesting {
		log.Println("Creating or replacing PostgreSQL function...")
		db.Conn.Exec(`
    CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
    BEGIN
        NEW.updated_at = now();
        RETURN NEW;
    END;
    $$ LANGUAGE plpgsql;
    `)
	}
	log.Println("Migrating User table...")
	if err := db.Conn.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}
	log.Println("Migrating ProviderPlatform table...")
	if err := db.Conn.AutoMigrate(&models.ProviderPlatform{}); err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}
	if err := db.Conn.AutoMigrate(&models.UserActivity{}); err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}
	if err := db.Conn.AutoMigrate(&models.ProviderUserMapping{}); err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}
	tables := []string{"users", "provider_platforms", "user_activities", "provider_user_mappings"}
	if !isTesting {
		log.Println("Applying update triggers...")
		ApplyUpdateTriggers(db.Conn, tables)
	}
	log.Println("Database successfully migrated.")
}

func (db *DB) MigrateFresh(isTesting bool) {
	if err := db.Conn.Migrator().DropTable(&models.User{}); err != nil {
		log.Fatalf("failed to drop tables: %v", err)
	}
	if err := db.Conn.Migrator().DropTable(&models.ProviderPlatform{}); err != nil {
		log.Fatalf("failed to drop tables: %v", err)
	}
	if err := db.Conn.Migrator().DropTable(&models.UserActivity{}); err != nil {
		log.Fatalf("failed to drop tables: %v", err)
	}
	if err := db.Conn.Migrator().DropTable(&models.ProviderUserMapping{}); err != nil {
		log.Fatalf("failed to drop tables: %v", err)
	}
	db.Migrate(isTesting)
	user := models.User{
		Username:      "SuperAdmin",
		NameFirst:     "Super",
		NameLast:      "Admin",
		Email:         "admin@unlocked.v2",
		PasswordReset: true,
		Role:          "admin",
		Password:      "ChangeMe!",
	}
	log.Printf("Creating user: %v", user)
	err := user.HashPassword()
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	log.Printf("Hashed password %s", user.Password)
	if err := db.Conn.Create(&user).Error; err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	log.Println("Database successfully migrated from fresh state.")
}

func ApplyUpdateTriggers(db *gorm.DB, tables []string) {
	db.Exec(`
        CREATE OR REPLACE FUNCTION universal_update_timestamp()
        RETURNS TRIGGER AS $$
        BEGIN
            NEW.updated_at = now();
            RETURN NEW;
        END;
        $$ LANGUAGE plpgsql;
    `)

	for _, table := range tables {
		db.Exec(fmt.Sprintf(`
            DO $$
            BEGIN
                IF NOT EXISTS (
                    SELECT 1 FROM pg_trigger WHERE tgname = 'update_%[1]s_modtime'
                ) THEN
                    CREATE TRIGGER update_%[1]s_modtime
                    BEFORE UPDATE ON %[1]s
                    FOR EACH ROW
                    EXECUTE PROCEDURE universal_update_timestamp();
                END IF;
            END
            $$;
        `, table))
	}
}

func (db *DB) SeedTestData() {
	users, err := os.ReadFile("test_data/users.json")
	if err != nil {
		log.Fatalf("Failed to read test data: %v", err)
	}
	var user []models.User
	if err := json.Unmarshal(users, &user); err != nil {
		log.Fatalf("Failed to unmarshal test data: %v", err)
	}
	for _, u := range user {
		if err := db.Conn.Create(&u).Error; err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
	}
	platforms, err := os.ReadFile("test_data/provider_platforms.json")
	if err != nil {
		log.Fatalf("Failed to read test data: %v", err)
	}
	var platform []models.ProviderPlatform
	if err := json.Unmarshal(platforms, &platform); err != nil {
		log.Fatalf("Failed to unmarshal test data: %v", err)
	}
	for _, p := range platform {
		if err := db.Conn.Create(&p).Error; err != nil {
			log.Fatalf("Failed to create platform: %v", err)
		}
	}
}
