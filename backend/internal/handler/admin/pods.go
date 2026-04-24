// backend/internal/handler/admin/pods.go

package admin

import (
	"context"
	"net/http"
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/service"

	"github.com/gin-gonic/gin"
)

type PodsHandler struct {
	podDiscovery service.PodDiscoveryServiceInterface
	wsService   service.WebSocketServiceInterface
}

func NewPodsHandler(
	podDiscovery service.PodDiscoveryServiceInterface,
	wsService service.WebSocketServiceInterface,
) *PodsHandler {
	return &PodsHandler{
		podDiscovery: podDiscovery,
		wsService:   wsService,
	}
}

func (h *PodsHandler) GetAllPods(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var pods []model.PodInfo

	if h.podDiscovery == nil {
		httputil.Success(c, http.StatusOK, model.PodsDiscovery{
			Pods: pods,
			Meta: model.PodsMeta{
				TotalPods:     0,
				MainPods:     0,
				HeadlessPods: 0,
				TotalClients: 0,
				TotalUsers:  0,
			},
		})
		return
	}

	var err error
	pods, err = h.podDiscovery.GetAllPods(ctx)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch pods")
		return
	}

	currentPodID := ""
	if h.wsService != nil {
		if podID, err := h.wsService.PodID(); err == nil {
			currentPodID = podID
		}
	}

	var totalClients, totalUsers int64
	for i := range pods {
		pods[i].IsCurrent = pods[i].PodID == currentPodID
		totalClients += pods[i].ConnectedClients
		totalUsers += pods[i].OnlineUsers
	}

	discovery := model.PodsDiscovery{
		Pods: pods,
		Meta: model.PodsMeta{
			TotalPods:    int64(len(pods)),
			MainPods:     0,
			HeadlessPods: 0,
			TotalClients: totalClients,
			TotalUsers:   totalUsers,
		},
	}

	for _, pod := range pods {
		if pod.PodType == model.PodTypeMain {
			discovery.Meta.MainPods++
		} else {
			discovery.Meta.HeadlessPods++
		}
	}

	httputil.Success(c, http.StatusOK, discovery)
}
