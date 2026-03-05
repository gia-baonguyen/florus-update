package service

import (
	"errors"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
	"github.com/florus/backend/pkg/utils"
	"gorm.io/gorm"
)

var (
	ErrReturnNotFound       = errors.New("return not found")
	ErrInvalidReturnItems   = errors.New("invalid return items")
)

type ReturnService interface {
	CreateReturn(userID uint, req dto.CreateReturnRequest) (*dto.ReturnResponse, error)
	GetUserReturns(userID uint, pagination utils.Pagination) ([]dto.ReturnResponse, *utils.Meta, error)
	GetReturnByID(userID uint, id uint) (*dto.ReturnResponse, error)
	UpdateReturnStatus(id uint, status models.ReturnStatus) error
}

type returnService struct {
	returnRepo repository.ReturnRepository
	orderRepo  repository.OrderRepository
}

func NewReturnService(returnRepo repository.ReturnRepository, orderRepo repository.OrderRepository) ReturnService {
	return &returnService{
		returnRepo: returnRepo,
		orderRepo:  orderRepo,
	}
}

func (s *returnService) CreateReturn(userID uint, req dto.CreateReturnRequest) (*dto.ReturnResponse, error) {
	// load order to validate ownership and items
	order, err := s.orderRepo.FindByID(req.OrderID)
	if err != nil || order.UserID != userID {
		return nil, ErrInvalidReturnItems
	}

	// map order_item_id -> max quantity
	itemQty := make(map[uint]int)
	for _, it := range order.OrderItems {
		itemQty[it.ID] = it.Quantity
	}

	var items []models.ReturnItem
	for _, rItem := range req.Items {
		maxQty, ok := itemQty[rItem.OrderItemID]
		if !ok || rItem.Quantity <= 0 || rItem.Quantity > maxQty {
			return nil, ErrInvalidReturnItems
		}
		// Simple refund amount: unitPrice * qty
		var orderItem models.OrderItem
		for _, oi := range order.OrderItems {
			if oi.ID == rItem.OrderItemID {
				orderItem = oi
				break
			}
		}
		refundAmount := orderItem.UnitPrice * float64(rItem.Quantity)
		items = append(items, models.ReturnItem{
			OrderItemID:  rItem.OrderItemID,
			Quantity:     rItem.Quantity,
			RefundAmount: refundAmount,
		})
	}

	if len(items) == 0 {
		return nil, ErrInvalidReturnItems
	}

	ret := &models.Return{
		OrderID: req.OrderID,
		UserID:  userID,
		Status:  models.ReturnStatusRequested,
		Reason:  req.Reason,
		Note:    req.Note,
	}

	if err := s.returnRepo.Create(nil, ret, items); err != nil {
		return nil, err
	}

	loaded, err := s.returnRepo.FindByIDForUser(ret.ID, userID)
	if err != nil {
		return nil, err
	}
	resp := dto.ToReturnResponse(loaded)
	return &resp, nil
}

func (s *returnService) GetUserReturns(userID uint, pagination utils.Pagination) ([]dto.ReturnResponse, *utils.Meta, error) {
	returns, total, err := s.returnRepo.FindByUserID(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	var res []dto.ReturnResponse
	for i := range returns {
		res = append(res, dto.ToReturnResponse(&returns[i]))
	}
	meta := utils.NewMeta(pagination.Page, pagination.Limit, total)
	return res, meta, nil
}

func (s *returnService) GetReturnByID(userID uint, id uint) (*dto.ReturnResponse, error) {
	ret, err := s.returnRepo.FindByIDForUser(id, userID)
	if err != nil {
		return nil, ErrReturnNotFound
	}
	resp := dto.ToReturnResponse(ret)
	return &resp, nil
}

func (s *returnService) UpdateReturnStatus(id uint, status models.ReturnStatus) error {
	db := s.returnRepo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		return s.returnRepo.UpdateStatus(tx, id, status)
	})
}

