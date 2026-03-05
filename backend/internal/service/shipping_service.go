package service

import (
	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/repository"
)

type ShippingService interface {
	GetShippingMethods() ([]dto.ShippingMethodResponse, error)
}

type shippingService struct {
	shippingRepo repository.ShippingRepository
}

func NewShippingService(shippingRepo repository.ShippingRepository) ShippingService {
	return &shippingService{shippingRepo: shippingRepo}
}

// For phase 1 we just return all active methods with static fees/eta from zone_methods or defaults.
func (s *shippingService) GetShippingMethods() ([]dto.ShippingMethodResponse, error) {
	methods, err := s.shippingRepo.GetMethods()
	if err != nil {
		return nil, err
	}

	var res []dto.ShippingMethodResponse
	for _, m := range methods {
		// For now we don't calculate fee per zone; that can be extended later.
		res = append(res, dto.ShippingMethodResponse{
			Code:        m.Code,
			Name:        m.Name,
			Description: m.Description,
		})
	}
	return res, nil
}

