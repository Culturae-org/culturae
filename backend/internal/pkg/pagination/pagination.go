// backend/internal/pkg/pagination/pagination.go

package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Params struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Offset     int   `json:"offset"`
	TotalCount int64 `json:"total_count,omitempty"`
	TotalPages int64 `json:"total_pages,omitempty"`
}

type Config struct {
	DefaultLimit int
	MaxLimit     int
	MinLimit     int
}

func DefaultConfig() Config {
	return Config{
		DefaultLimit: 20,
		MaxLimit:     100,
		MinLimit:     1,
	}
}

func AdminConfig() Config {
	return Config{
		DefaultLimit: 50,
		MaxLimit:     1000,
		MinLimit:     1,
	}
}

func Parse(c *gin.Context, config ...Config) Params {
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := cfg.DefaultLimit
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if limit < cfg.MinLimit {
		limit = cfg.MinLimit
	}
	if limit > cfg.MaxLimit {
		limit = cfg.MaxLimit
	}

	offset := (page - 1) * limit

	return Params{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}

func (p *Params) WithTotal(totalCount int64) *Params {
	p.TotalCount = totalCount
	p.TotalPages = (totalCount + int64(p.Limit) - 1) / int64(p.Limit)
	if p.TotalPages < 1 {
		p.TotalPages = 1
	}
	return p
}

func (p *Params) HasNextPage() bool {
	return int64(p.Page) < p.TotalPages
}

func (p *Params) HasPrevPage() bool {
	return p.Page > 1
}

type Meta struct {
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	TotalCount  int64 `json:"total_count"`
	TotalPages  int64 `json:"total_pages"`
	HasNextPage bool  `json:"has_next_page"`
	HasPrevPage bool  `json:"has_prev_page"`
}
