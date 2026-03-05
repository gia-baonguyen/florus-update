package repository

import (
	"github.com/florus/backend/internal/models"
	"gorm.io/gorm"
)

type UserAddressRepository interface {
	FindByUserID(userID uint) ([]models.UserAddress, error)
	FindByIDForUser(id, userID uint) (*models.UserAddress, error)
	Create(addr *models.UserAddress) error
	Update(addr *models.UserAddress) error
	Delete(id, userID uint) error
	UnsetDefaultForUser(userID uint) error
	SetDefault(id, userID uint) error
}

type userAddressRepository struct {
	db *gorm.DB
}

func NewUserAddressRepository(db *gorm.DB) UserAddressRepository {
	return &userAddressRepository{db: db}
}

func (r *userAddressRepository) FindByUserID(userID uint) ([]models.UserAddress, error) {
	var addrs []models.UserAddress
	err := r.db.Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&addrs).Error
	return addrs, err
}

func (r *userAddressRepository) FindByIDForUser(id, userID uint) (*models.UserAddress, error) {
	var addr models.UserAddress
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&addr).Error
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

func (r *userAddressRepository) Create(addr *models.UserAddress) error {
	return r.db.Create(addr).Error
}

func (r *userAddressRepository) Update(addr *models.UserAddress) error {
	return r.db.Save(addr).Error
}

func (r *userAddressRepository) Delete(id, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.UserAddress{}).Error
}

func (r *userAddressRepository) UnsetDefaultForUser(userID uint) error {
	return r.db.Model(&models.UserAddress{}).
		Where("user_id = ?", userID).
		Update("is_default", false).Error
}

func (r *userAddressRepository) SetDefault(id, userID uint) error {
	return r.db.Model(&models.UserAddress{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_default", true).Error
}

