// backend/internal/pkg/httputil/pagination.go

package httputil

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Total   int  `json:"total"`
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	HasMore bool `json:"has_more"`
}

func ParsePagination(c *gin.Context, defaultLimit, maxLimit int) (limit, offset int) {
	limit = defaultLimit
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	offset = 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}

func ParamsToPagination(totalCount int64, limit, offset int) Pagination {
	hasMore := totalCount > int64(offset+limit)
	return Pagination{
		Total:   int(totalCount),
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}
}
