// backend/internal/pkg/httputil/request.go

package httputil

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetRealIP(c *gin.Context) string {
	if ip := c.GetHeader("CF-Connecting-IP"); ip != "" {
		return ip
	}
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	return c.ClientIP()
}

func GetUserAgent(c *gin.Context) string {
	return c.Request.UserAgent()
}

func GetUserIDFromContext(c *gin.Context) uuid.UUID {
	idVal, exists := c.Get("user_id")
	if !exists {
		idStr := c.GetString("user_id")
		if idStr == "" {
			return uuid.Nil
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			return uuid.Nil
		}
		return id
	}

	switch v := idVal.(type) {
	case uuid.UUID:
		return v
	case *uuid.UUID:
		if v == nil {
			return uuid.Nil
		}
		return *v
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return uuid.Nil
		}
		return id
	default:
		idStr := fmt.Sprintf("%v", v)
		id, err := uuid.Parse(idStr)
		if err != nil {
			return uuid.Nil
		}
		return id
	}
}

func GetPublicIDFromContext(c *gin.Context) string {
	publicID, exists := c.Get("public_id")
	if !exists {
		return ""
	}
	switch v := publicID.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func QueryUUID(c *gin.Context, key string) *uuid.UUID {
	val := c.Query(key)
	if val == "" {
		return nil
	}
	if id, err := uuid.Parse(val); err == nil {
		return &id
	}
	return nil
}

func QueryString(c *gin.Context, key string) *string {
	val := c.Query(key)
	if val == "" {
		return nil
	}
	return &val
}

func QueryBool(c *gin.Context, key string) *bool {
	val := c.Query(key)
	if val == "" {
		return nil
	}
	lower := strings.ToLower(val)
	if lower == "true" || lower == "1" || lower == "yes" {
		b := true
		return &b
	}
	if lower == "false" || lower == "0" || lower == "no" {
		b := false
		return &b
	}
	return nil
}

func QueryInt(c *gin.Context, key string) *int {
	val := c.Query(key)
	if val == "" {
		return nil
	}
	if i, err := strconv.Atoi(val); err == nil {
		return &i
	}
	return nil
}

func QueryDateRange(c *gin.Context, startKey, endKey string) (*time.Time, *time.Time) {
	var startDate, endDate *time.Time

	if startStr := c.Query(startKey); startStr != "" {
		if start, err := time.Parse("2006-01-02", startStr); err == nil {
			startDate = &start
		}
	}

	if endStr := c.Query(endKey); endStr != "" {
		if end, err := time.Parse("2006-01-02", endStr); err == nil {
			endDate = &end
		}
	}

	return startDate, endDate
}
