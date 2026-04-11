package repositories

import (
	"go-postfixadmin/internal/models"

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
func CreateTransport(db *gorm.DB, transport models.TransportList) error {
	return db.Create(&transport).Error
}

// UpdateTransport updates an existing transport entry
func UpdateTransport(db *gorm.DB, id int, updates map[string]interface{}) error {
	return db.Model(&models.TransportList{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteTransport deletes a transport entry by ID
func DeleteTransport(db *gorm.DB, id int) error {
	return db.Where("id = ?", id).Delete(&models.TransportList{}).Error
}
