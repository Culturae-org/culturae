package handler

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	podID string
}

func NewHealthHandler() *HealthHandler {
	podID := os.Getenv("HOSTNAME")
	if podID == "" {
		podID = os.Getenv("POD_NAME")
	}
	if podID == "" {
		podID = "unknown"
	}
	
	return &HealthHandler{
		podID: podID,
	}
}

func (h *HealthHandler) Health(c *gin.Context) {
	response := gin.H{
		"status": "healthy",
		"pod_id": h.podID,
		"version": "1.0.0",
	}

	if mode := os.Getenv("APP_MODE"); mode != "" {
		response["app_mode"] = mode
	}
	
	if env := os.Getenv("ENV"); env != "" {
		response["environment"] = env
	}

	c.JSON(http.StatusOK, response)
}