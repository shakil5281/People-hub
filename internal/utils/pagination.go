package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// Pagination holds standard pagination parameters
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// DefaultPagination returns page=1, limit=20
func DefaultPagination() Pagination {
	return Pagination{Page: 1, Limit: 20}
}

// ParsePagination extracts page and limit from gin query params
func ParsePagination(c *gin.Context) Pagination {
	p := DefaultPagination()
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			p.Page = page
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			if limit > 5000 {
				limit = 5000 // cap max page size
			}
			p.Limit = limit
		}
	}
	return p
}

// Offset calculates SQL OFFSET from page and limit
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.Limit
}

// PaginatedResponse is the standard list API response wrapper
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int64       `json:"total_pages"`
}

// NewPaginatedResponse builds a paginated response from data, total count, and pagination params
func NewPaginatedResponse(data interface{}, total int64, p Pagination) PaginatedResponse {
	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + int64(p.Limit) - 1) / int64(p.Limit)
	}
	return PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       p.Page,
		Limit:      p.Limit,
		TotalPages: totalPages,
	}
}
