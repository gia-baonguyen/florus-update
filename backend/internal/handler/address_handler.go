package handler

import (
	"strconv"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/middleware"
	"github.com/florus/backend/internal/service"
	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type AddressHandler struct {
	addressService service.AddressService
}

func NewAddressHandler(addressService service.AddressService) *AddressHandler {
	return &AddressHandler{addressService: addressService}
}

func (h *AddressHandler) GetAddresses(c *gin.Context) {
	userID := middleware.GetUserID(c)
	addrs, err := h.addressService.GetUserAddresses(userID)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}
	utils.OK(c, "Addresses retrieved successfully", addrs)
}

func (h *AddressHandler) CreateAddress(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	addr, err := h.addressService.CreateAddress(userID, req)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}
	utils.Created(c, "Address created successfully", addr)
}

func (h *AddressHandler) UpdateAddress(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid address ID")
		return
	}
	var req dto.UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	addr, err := h.addressService.UpdateAddress(userID, uint(id), req)
	if err != nil {
		if err == service.ErrAddressNotFound {
			utils.NotFound(c, "Address not found")
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}
	utils.OK(c, "Address updated successfully", addr)
}

func (h *AddressHandler) DeleteAddress(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid address ID")
		return
	}
	if err := h.addressService.DeleteAddress(userID, uint(id)); err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}
	utils.OK(c, "Address deleted successfully", nil)
}

func (h *AddressHandler) SetDefaultAddress(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid address ID")
		return
	}
	if err := h.addressService.SetDefaultAddress(userID, uint(id)); err != nil {
		if err == service.ErrAddressNotFound {
			utils.NotFound(c, "Address not found")
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}
	utils.OK(c, "Default address updated successfully", nil)
}

