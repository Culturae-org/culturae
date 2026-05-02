// backend/internal/handler/admin/game_template.go

package admin

import (
	"net/http"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/pagination"
	"github.com/Culturae-org/culturae/internal/service"
	adminUsecase "github.com/Culturae-org/culturae/internal/usecase/admin"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AdminGameTemplatesHandler struct {
	usecase        *adminUsecase.AdminGameTemplatesUsecase
	LoggingService service.LoggingServiceInterface
	logger         *zap.Logger
}

func NewAdminGameTemplatesHandler(
	uc *adminUsecase.AdminGameTemplatesUsecase,
	loggingService service.LoggingServiceInterface,
	logger *zap.Logger,
) *AdminGameTemplatesHandler {
	return &AdminGameTemplatesHandler{
		usecase:        uc,
		LoggingService: loggingService,
		logger:         logger,
	}
}

// TODO : need to change functions name

// -----------------------------------------------------
// Admin Game Templates Handlers
//
// - List
// - Get
// - Create
// - Update
// - Delete
// - SeedDefaults
// -----------------------------------------------------

func (h *AdminGameTemplatesHandler) List(c *gin.Context) {
	pag := pagination.Parse(c, pagination.AdminConfig())

	params := model.GameTemplateListParams{
		IsActive: httputil.QueryBool(c, "active_only"),
		Mode:     c.Query("mode"),
		Category: c.Query("category"),
		Query:    c.Query("q"),
		Limit:    pag.Limit,
		Offset:   pag.Offset,
	}

	templates, err := h.usecase.ListGameTemplates(params)
	if err != nil {
		h.logger.Error("Failed to list game templates", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to list game templates")
		return
	}

	total, err := h.usecase.CountGameTemplates(params)
	if err != nil {
		h.logger.Error("Failed to count game templates", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to list game templates")
		return
	}

	pag.WithTotal(total)
	httputil.SuccessList(c, templates, httputil.ParamsToPagination(pag.TotalCount, pag.Limit, pag.Offset))
}

func (h *AdminGameTemplatesHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid template ID")
		return
	}

	t, err := h.usecase.GetGameTemplateByID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Game template not found")
		return
	}

	httputil.Success(c, http.StatusOK, t)
}

func (h *AdminGameTemplatesHandler) Create(c *gin.Context) {
	var req model.CreateGameTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)

	t, err := h.usecase.CreateGameTemplate(&req, adminUUID)
	if err != nil {
		h.logger.Error("Failed to create game template", zap.Error(err))
		errMsg := err.Error()
		httputil.LogAdminAction(h.LoggingService, adminUUID, c.GetString("username"), "game_template_create", "game_template", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{keyName: req.Name}, false, &errMsg)
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.LogAdminAction(h.LoggingService, adminUUID, c.GetString("username"), "game_template_create", "game_template", &t.ID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{keyName: t.Name}, true, nil)
	httputil.Success(c, http.StatusCreated, t)
}

func (h *AdminGameTemplatesHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid template ID")
		return
	}

	var req model.UpdateGameTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)

	t, err := h.usecase.UpdateGameTemplate(id, &req, adminUUID)
	if err != nil {
		h.logger.Error("Failed to update game template", zap.String("id", id.String()), zap.Error(err))
		errMsg := err.Error()
		httputil.LogAdminAction(h.LoggingService, adminUUID, c.GetString("username"), "game_template_update", "game_template", &id, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{"template_id": id}, false, &errMsg)
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.LogAdminAction(h.LoggingService, adminUUID, c.GetString("username"), "game_template_update", "game_template", &id, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{"template_id": id, keyName: t.Name}, true, nil)
	httputil.Success(c, http.StatusOK, t)
}

func (h *AdminGameTemplatesHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid template ID")
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)

	if err := h.usecase.DeleteGameTemplate(id, adminUUID); err != nil {
		h.logger.Error("Failed to delete game template", zap.String("id", id.String()), zap.Error(err))
		errMsg := err.Error()
		httputil.LogAdminAction(h.LoggingService, adminUUID, c.GetString("username"), "game_template_delete", "game_template", &id, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{"template_id": id}, false, &errMsg)
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.LogAdminAction(h.LoggingService, adminUUID, c.GetString("username"), "game_template_delete", "game_template", &id, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{"template_id": id}, true, nil)
	httputil.Success(c, http.StatusOK, gin.H{"message": "Game template deleted"})
}

func (h *AdminGameTemplatesHandler) SeedDefaults(c *gin.Context) {
	adminUUID := httputil.GetUserIDFromContext(c)
	count, err := h.usecase.SeedDefaultGameTemplates()
	if err != nil {
		h.logger.Error("Failed to seed default game templates", zap.Error(err))
		errMsg := err.Error()
		httputil.LogAdminAction(h.LoggingService, adminUUID, c.GetString("username"), "game_template_seed_defaults", "game_template", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), nil, false, &errMsg)
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to seed default templates")
		return
	}
	httputil.LogAdminAction(h.LoggingService, adminUUID, c.GetString("username"), "game_template_seed_defaults", "game_template", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{"created": count}, true, nil)
	httputil.Success(c, http.StatusOK, gin.H{"created": count})
}
