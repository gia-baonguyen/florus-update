package repository

import (
	"github.com/florus/backend/internal/models"
	"gorm.io/gorm"
)

type ShippingRepository interface {
	GetZones() ([]models.ShippingZone, error)
	GetMethods() ([]models.ShippingMethod, error)
	GetZoneMethodsByZoneID(zoneID uint) ([]models.ShippingZoneMethod, error)
}

type shippingRepository struct {
	db *gorm.DB
}

func NewShippingRepository(db *gorm.DB) ShippingRepository {
	return &shippingRepository{db: db}
}

func (r *shippingRepository) GetZones() ([]models.ShippingZone, error) {
	var zones []models.ShippingZone
	err := r.db.Order("id ASC").Find(&zones).Error
	return zones, err
}

func (r *shippingRepository) GetMethods() ([]models.ShippingMethod, error) {
	var methods []models.ShippingMethod
	err := r.db.Where("is_active = 1").
		Order("id ASC").
		Find(&methods).Error
	return methods, err
}

func (r *shippingRepository) GetZoneMethodsByZoneID(zoneID uint) ([]models.ShippingZoneMethod, error) {
	var records []models.ShippingZoneMethod
	err := r.db.Where("zone_id = ?", zoneID).Find(&records).Error
	return records, err
}

