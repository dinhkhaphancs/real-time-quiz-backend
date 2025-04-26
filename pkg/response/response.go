// Package response provides standardized API response handling for the application
package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response is the standard API response structure
type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Pagination represents pagination metadata for list responses
type Pagination struct {
	Total       int `json:"total"`
	PerPage     int `json:"perPage"`
	CurrentPage int `json:"currentPage"`
	LastPage    int `json:"lastPage"`
}

// PaginatedResponse extends Response with pagination data
type PaginatedResponse struct {
	Response
	Pagination Pagination `json:"pagination,omitempty"`
}

// NewResponse creates a new API response
func NewResponse(success bool, message string, data interface{}) Response {
	return Response{
		Success:   success,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(message string, err string) Response {
	return Response{
		Success:   false,
		Message:   message,
		Error:     err,
		Timestamp: time.Now(),
	}
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(message string, data interface{}, pagination Pagination) PaginatedResponse {
	return PaginatedResponse{
		Response: Response{
			Success:   true,
			Message:   message,
			Data:      data,
			Timestamp: time.Now(),
		},
		Pagination: pagination,
	}
}

// WithSuccess sends a success response with the given data
func WithSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, NewResponse(true, message, data))
}

// WithError sends an error response
func WithError(c *gin.Context, statusCode int, message string, err string) {
	c.JSON(statusCode, NewErrorResponse(message, err))
}

// WithPagination sends a paginated response
func WithPagination(c *gin.Context, message string, data interface{}, total, perPage, currentPage int) {
	lastPage := 0
	if perPage > 0 {
		lastPage = (total + perPage - 1) / perPage
	}

	pagination := Pagination{
		Total:       total,
		PerPage:     perPage,
		CurrentPage: currentPage,
		LastPage:    lastPage,
	}

	c.JSON(http.StatusOK, NewPaginatedResponse(message, data, pagination))
}

// Common response messages
const (
	MessageCreated     = "Resource created successfully"
	MessageUpdated     = "Resource updated successfully"
	MessageDeleted     = "Resource deleted successfully"
	MessageFetched     = "Resource fetched successfully"
	MessageListFetched = "Resources fetched successfully"
)
