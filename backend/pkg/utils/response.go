package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessWithMeta(c *gin.Context, statusCode int, message string, data interface{}, meta *Meta) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, message string, err string) {
	c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Error:   err,
	})
}

func BadRequest(c *gin.Context, err string) {
	ErrorResponse(c, http.StatusBadRequest, "Bad Request", err)
}

func Unauthorized(c *gin.Context, err string) {
	ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err)
}

func Forbidden(c *gin.Context, err string) {
	ErrorResponse(c, http.StatusForbidden, "Forbidden", err)
}

func NotFound(c *gin.Context, err string) {
	ErrorResponse(c, http.StatusNotFound, "Not Found", err)
}

func InternalServerError(c *gin.Context, err string) {
	ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error", err)
}

func Created(c *gin.Context, message string, data interface{}) {
	SuccessResponse(c, http.StatusCreated, message, data)
}

func OK(c *gin.Context, message string, data interface{}) {
	SuccessResponse(c, http.StatusOK, message, data)
}
