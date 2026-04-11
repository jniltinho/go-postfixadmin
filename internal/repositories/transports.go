package repositories

import (
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"gorm.io/gorm"
)

// GetAllTransports returns all available transports
func GetAllTransports(db *gorm.DB) ([]models.TransportList, error) {
	var transports []models.TransportList
	err := db.Order("transport ASC").Find(&transports).Error
	return transports, err
}

// GetTransportByID finds a transport by its ID
func GetTransportByID(db *gorm.DB, id int) (models.TransportList, error) {
	var transport models.TransportList
	err := db.Where("id = ?", id).First(&transport).Error
	return transport, err
}

// CreateTransport creates a new transport entry
func CreateTransport(db *gorm.DB, transport models.TransportList, actorUsername, ip string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&transport).Error; err != nil {
			return err
		}
		return utils.LogAction(tx, actorUsername, ip, transport.Domain, "create_transport", transport.Transport)
	})
}

// UpdateTransport updates an existing transport entry
func UpdateTransport(db *gorm.DB, id int, updates map[string]interface{}, actorUsername, ip string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.TransportList{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		transport, _ := updates["transport"].(string)
		return utils.LogAction(tx, actorUsername, ip, transport, "edit_transport", transport)
	})
}

// DeleteTransport deletes a transport entry by ID
func DeleteTransport(db *gorm.DB, id int, actorUsername, ip string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var transport models.TransportList
		if err := tx.Where("id = ?", id).First(&transport).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).Delete(&models.TransportList{}).Error; err != nil {
			return err
		}
		return utils.LogAction(tx, actorUsername, ip, transport.Domain, "delete_transport", transport.Transport)
	})
}
