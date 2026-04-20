// backend/internal/middleware/api_logging.go

package middleware

import (
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/service"
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type APILoggingMiddleware struct {
	LoggingService service.LoggingServiceInterface
	Logger         *zap.Logger
}

func NewAPILoggingMiddleware(
	loggingService service.LoggingServiceInterface,
	logger *zap.Logger,
) *APILoggingMiddleware {
	return &APILoggingMiddleware{
		LoggingService: loggingService,
		Logger:         logger,
	}
}

func (alm *APILoggingMiddleware) LogAPIRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		responseWriter := &responseCaptureWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = responseWriter

		c.Next()

		duration := time.Since(start).Nanoseconds() / 1000000

		var userID *uuid.UUID
		if userIDStr, exists := c.Get("user_id"); exists {
			if uid, ok := userIDStr.(uuid.UUID); ok {
				userID = &uid
			}
		}

		isError := c.Writer.Status() >= 400
		var errorMsg *string
		if isError && len(c.Errors) > 0 {
			errMsg := c.Errors.Last().Error()
			errorMsg = &errMsg
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				done <- alm.LoggingService.LogAPIRequest(
					c.Request.Method,
					c.Request.URL.Path,
					c.Writer.Status(),
					userID,
					httputil.GetRealIP(c),
					c.GetHeader("User-Agent"),
					int64(len(requestBody)),
					int64(responseWriter.body.Len()),
					duration,
					isError,
					errorMsg,
				)
			}()

			select {
			case err := <-done:
				if err != nil {
					alm.Logger.Error("Failed to log API request asynchronously", zap.Error(err))
				}
			case <-ctx.Done():
				alm.Logger.Warn("API request logging timed out")
			}
		}()
	}
}

type responseCaptureWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseCaptureWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}
