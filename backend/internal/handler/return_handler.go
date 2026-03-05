package handler

import (
	"strconv"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/middleware"
	"github.com/florus/backend/internal/service"
	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type ReturnHandler struct {
	returnService service.ReturnService
}

func NewReturnHandler(returnService service.ReturnService) *ReturnHandler {
	return &ReturnHandler{returnService: returnService}
}

func (h *ReturnHandler) CreateReturn(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.CreateReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	ret, err := h.returnService.CreateReturn(userID, req)
	if err != nil {
		if err == service.ErrInvalidReturnItems {
			utils.BadRequest(c, "Invalid return items")
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}
	utils.Created(c, "Return request created successfully", ret)
}

func (h *ReturnHandler) GetReturns(c *gin.Context) {
	userID := middleware.GetUserID(c)
	pagination := utils.GetPagination(c)
	returns, meta, err := h.returnService.GetUserReturns(userID, pagination)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}
	utils.SuccessWithMeta(c, 200, "Returns retrieved successfully", returns, meta)
}

func (h *ReturnHandler) GetReturnByID(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid return ID")
		return
	}
	ret, err := h.returnService.GetReturnByID(userID, uint(id))
	if err != nil {
		if err == service.ErrReturnNotFound {
			utils.NotFound(c, "Return not found")
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}
	utils.OK(c, "Return retrieved successfully", ret)
}

