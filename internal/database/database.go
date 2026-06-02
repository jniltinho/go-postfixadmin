// Package database handles connection setup and schema migration for go-postfixadmin.
package database

import (
	"fmt"
	"os"

	"go-postfixadmin/internal/models"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func viperOrEnv(key, env, fallback string) string {
	if v := viper.GetString(key); v != "" {
		return v
	}
	if v := os.Getenv(env); v != "" {
		return v
	}
	return fallback
}

// ConnectDB initializes the database connection using the configured driver
// (mysql or postgres). DSN and driver may be supplied directly or resolved from
// viper / environment variables.
func ConnectDB(dsn string, driver string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	if driver == "" {
		driver = viperOrEnv("database.driver", "DB_DRIVER", "mysql")
	}

	if dsn == "" {
		// Prefer explicit URL; fall back to individual fields.
		dsn = viper.GetString("database.url")
		if dsn == "" {
			dsn = os.Getenv("DB_URL")
		}
		if dsn == "" {
			host := viperOrEnv("database.host", "DB_HOST", "localhost")
			port := viperOrEnv("database.port", "DB_PORT", "3306")
			user := viperOrEnv("database.user", "DB_USER", "postfix")
			pass := viperOrEnv("database.pass", "DB_PASS", "")
			name := viperOrEnv("database.name", "DB_NAME", "postfix")
			if driver == "postgres" {
				dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, name)
			} else {
				dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local&multiStatements=true", user, pass, host, port, name)
			}
		}
	}

	gormConfig := &gorm.Config{}
	if viper.GetBool("database.debug") || viper.GetBool("debug") {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	if driver == "postgres" {
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	} else {
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	}

	return db, err
}

// MigrateDB runs GORM AutoMigrate for the core PostfixAdmin schema.
// Call MigrateRBAC separately (or via "migrate rbac" CLI) to add the RBAC tables.
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
		&models.Maillog{},
		&models.TransportList{},
		&models.AdminApiKey{},
	)
}

// MigrateRBAC creates the four RBAC tables (rbac_roles, rbac_permissions,
// rbac_role_permissions, rbac_admin_roles) via GORM AutoMigrate.
// This is intentionally separate from MigrateDB to allow opt-in deployment.
// After migration, callers should also invoke [rbac.Seed] to populate the
// built-in system roles and permission catalog.
func MigrateRBAC(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.RBACRole{},
		&models.RBACPermission{},
		&models.RBACAdminRole{},
	)
}
