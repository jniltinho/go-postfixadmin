package utils

import (
	"os"

	"go-postfixadmin/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB initializes the database connection
func ConnectDB(dsn string, driver string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}

	if dsn == "" {
		dsn = "user:password@tcp(localhost:3306)/postfixadmin?charset=utf8mb4&parseTime=True&loc=Local"
	}

	if driver == "" {
		driver = os.Getenv("DB_DRIVER")
	}

	if driver == "postgres" {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	} else {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	}

	return db, err
}

// MigrateDB migrates the database schema using GORM AutoMigrate
func MigrateDB(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Admin{},
		&models.Alias{},
		&models.AliasDomain{},
		&models.Config{},
		&models.DKIM{},
		&models.DKIMSigning{},
		&models.Domain{},
		&models.DomainAdmin{},
		&models.Fetchmail{},
		&models.Log{},
		&models.Mailbox{},
		&models.MailboxAppPassword{},
		&models.Quota{},
		&models.Quota2{},
		&models.TOTPExceptionAddress{},
		&models.Vacation{},
		&models.VacationNotification{},
	)
}
