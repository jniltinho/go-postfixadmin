package utils

import (
	"log"
	"os"

	"gorm.io/gorm"
)

// ImportSQL reads a .sql file and executes it against the MySQL database.
//
// Note: For this to work with multiple SQL statements in a single file,
// ensure your MySQL DSN includes "multiStatements=true".
// Example: "user:pass@tcp(localhost:3306)/dbname?multiStatements=true"
func ImportSQL(db *gorm.DB, filePath string) error {
	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading SQL file %s: %v", filePath, err)
		return err
	}

	// Convert content to string
	sqlCommands := string(content)

	// Execute the SQL commands
	if err := db.Exec(sqlCommands).Error; err != nil {
		log.Printf("Error executing SQL command from file %s: %v", filePath, err)
		return err
	}

	log.Printf("Successfully executed SQL file: %s", filePath)
	return nil
}
