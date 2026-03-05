package service

import (
	"errors"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
)

var (
	ErrAddressNotFound = errors.New("address not found")
)

type AddressService interface {
	GetUserAddresses(userID uint) ([]dto.UserAddressResponse, error)
	CreateAddress(userID uint, req dto.CreateAddressRequest) (*dto.UserAddressResponse, error)
	UpdateAddress(userID, addressID uint, req dto.UpdateAddressRequest) (*dto.UserAddressResponse, error)
	DeleteAddress(userID, addressID uint) error
	SetDefaultAddress(userID, addressID uint) error
}

type addressService struct {
	addrRepo repository.UserAddressRepository
}

func NewAddressService(addrRepo repository.UserAddressRepository) AddressService {
	return &addressService{addrRepo: addrRepo}
}

func (s *addressService) GetUserAddresses(userID uint) ([]dto.UserAddressResponse, error) {
	addrs, err := s.addrRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	var res []dto.UserAddressResponse
	for i := range addrs {
		res = append(res, dto.ToUserAddressResponse(&addrs[i]))
	}
	return res, nil
}

func (s *addressService) CreateAddress(userID uint, req dto.CreateAddressRequest) (*dto.UserAddressResponse, error) {
	addr := &models.UserAddress{
		UserID:      userID,
		FullName:    req.FullName,
		Phone:       req.Phone,
		Street:      req.Street,
		City:        req.City,
		State:       req.State,
		PostalCode:  req.PostalCode,
		CountryCode: req.CountryCode,
		IsDefault:   req.IsDefault,
	}

	if req.IsDefault {
		// unset previous defaults
		if err := s.addrRepo.UnsetDefaultForUser(userID); err != nil {
			return nil, err
		}
	}

	if err := s.addrRepo.Create(addr); err != nil {
		return nil, err
	}

	resp := dto.ToUserAddressResponse(addr)
	return &resp, nil
}

func (s *addressService) UpdateAddress(userID, addressID uint, req dto.UpdateAddressRequest) (*dto.UserAddressResponse, error) {
	addr, err := s.addrRepo.FindByIDForUser(addressID, userID)
	if err != nil {
		return nil, ErrAddressNotFound
	}

	if req.FullName != "" {
		addr.FullName = req.FullName
	}
	if req.Phone != "" {
		addr.Phone = req.Phone
	}
	if req.Street != "" {
		addr.Street = req.Street
	}
	if req.City != "" {
		addr.City = req.City
	}
	if req.State != "" {
		addr.State = req.State
	}
	if req.PostalCode != "" {
		addr.PostalCode = req.PostalCode
	}
	if req.CountryCode != "" {
		addr.CountryCode = req.CountryCode
	}

	if req.IsDefault != nil && *req.IsDefault {
		if err := s.addrRepo.UnsetDefaultForUser(userID); err != nil {
			return nil, err
		}
		addr.IsDefault = true
	} else if req.IsDefault != nil && !*req.IsDefault {
		addr.IsDefault = false
	}

	if err := s.addrRepo.Update(addr); err != nil {
		return nil, err
	}

	resp := dto.ToUserAddressResponse(addr)
	return &resp, nil
}

func (s *addressService) DeleteAddress(userID, addressID uint) error {
	return s.addrRepo.Delete(addressID, userID)
}

func (s *addressService) SetDefaultAddress(userID, addressID uint) error {
	// ensure address belongs to user
	if _, err := s.addrRepo.FindByIDForUser(addressID, userID); err != nil {
		return ErrAddressNotFound
	}
	if err := s.addrRepo.UnsetDefaultForUser(userID); err != nil {
		return err
	}
	return s.addrRepo.SetDefault(addressID, userID)
}

