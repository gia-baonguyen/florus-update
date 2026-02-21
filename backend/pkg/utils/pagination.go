package utils

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

func GetPagination(c *gin.Context) Pagination {
	page, _ := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(DefaultPage)))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(DefaultLimit)))

	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	return Pagination{
		Page:  page,
		Limit: limit,
	}
}

func (p *Pagination) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

func CalculateTotalPages(total int64, limit int) int {
	return int(math.Ceil(float64(total) / float64(limit)))
}

func NewMeta(page, limit int, total int64) *Meta {
	return &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: CalculateTotalPages(total, limit),
	}
}
