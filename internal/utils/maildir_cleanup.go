package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"go-postfixadmin/internal/models"

	"gorm.io/gorm"
)

// CleanupOrphanedMaildir verifica se uma mailbox existe no banco de dados.
// Se não existir, mas o diretório físico existir no servidor (/var/vmail/domain/user),
// o diretório será apagado.
// baseDir is typically "/var/vmail".
func CleanupOrphanedMaildir(db *gorm.DB, baseDir, domain, localPart string) error {
	username := fmt.Sprintf("%s@%s", localPart, domain)

	// Verifica se a mailbox existe
	var mailbox models.Mailbox
	err := db.Where("username = ?", username).First(&mailbox).Error
	if err == nil {
		// A mailbox existe, não devemos apagar a pasta
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		// Ocorreu um erro no banco de dados
		return fmt.Errorf("erro ao verificar mailbox no banco de dados: %w", err)
	}

	// A mailbox não existe, vamos verificar se o diretório físico existe
	maildirPath := filepath.Join(baseDir, domain, localPart)

	if _, err := os.Stat(maildirPath); os.IsNotExist(err) {
		// O diretório não existe, nada a fazer
		return nil
	}

	// O diretório existe, vamos apagá-lo
	if err := os.RemoveAll(maildirPath); err != nil {
		return fmt.Errorf("erro ao apagar diretório %s: %w", maildirPath, err)
	}

	return nil
}
