// backend/internal/handler/admin/questions.go

package admin

import (
	"net/http"
	"strings"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/pagination"
	"github.com/Culturae-org/culturae/internal/service"
	adminUsecase "github.com/Culturae-org/culturae/internal/usecase/admin"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AdminQuestionHandler struct {
	Usecase              *adminUsecase.AdminQuestionsUsecase
	AdminDatasetsUsecase *adminUsecase.AdminDatasetsUsecase
	LoggingService       service.LoggingServiceInterface
	wsService            service.WebSocketServiceInterface
	logger               *zap.Logger
}

func NewAdminQuestionHandler(
	usecase *adminUsecase.AdminQuestionsUsecase,
	adminDatasetsUsecase *adminUsecase.AdminDatasetsUsecase,
	loggingService service.LoggingServiceInterface,
	wsService service.WebSocketServiceInterface,
	logger *zap.Logger,
) *AdminQuestionHandler {
	return &AdminQuestionHandler{
		Usecase:              usecase,
		AdminDatasetsUsecase: adminDatasetsUsecase,
		LoggingService:       loggingService,
		wsService:            wsService,
		logger:               logger,
	}
}

// -----------------------------------------------------
// Admin Question Handlers
//
// - CreateQuestion
// - GetQuestion
// - GetQuestionBySlug
// - UpdateQuestion
// - DeleteQuestion
// - ListQuestions
// - BackupQuestions
// - ExportQuestionsClean
// -----------------------------------------------------

func (qc *AdminQuestionHandler) CreateQuestion(c *gin.Context) {
	var req model.QuestionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	question, err := qc.Usecase.CreateQuestion(req)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	adminID := httputil.GetUserIDFromContext(c)
	if adminID != uuid.Nil {
		adminName := c.GetString("username")
		go func() {
			if err := qc.LoggingService.LogAdminAction(adminID, adminName, "question_create", "question", &question.ID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				keySlug:       question.Slug,
				keyTheme:      question.Theme,
				keyDifficulty: question.Difficulty,
			}, true, nil); err != nil {
				qc.logger.Warn("Failed to log admin action", zap.Error(err))
			}
		}()
	}

	httputil.Success(c, http.StatusCreated, question)
}

func (qc *AdminQuestionHandler) GetQuestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid question ID")
		return
	}

	question, err := qc.Usecase.GetQuestionByID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Question not found")
		return
	}

	httputil.Success(c, http.StatusOK, question)
}

func (qc *AdminQuestionHandler) GetQuestionBySlug(c *gin.Context) {
	slug := c.Param(keySlug)
	datasetIDStr := c.Query("dataset_id")

	var datasetID *uuid.UUID
	if datasetIDStr != "" {
		parsed, err := uuid.Parse(datasetIDStr)
		if err != nil {
			httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset_id format")
			return
		}
		datasetID = &parsed
	}

	question, err := qc.Usecase.GetQuestionBySlug(slug, datasetID)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Question not found")
		return
	}

	httputil.Success(c, http.StatusOK, question)
}

func (qc *AdminQuestionHandler) UpdateQuestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid question ID")
		return
	}

	var req model.QuestionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	question, err := qc.Usecase.UpdateQuestion(id, req)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	adminID := httputil.GetUserIDFromContext(c)
	if adminID != uuid.Nil {
		adminName := c.GetString("username")
		go func() {
			if err := qc.LoggingService.LogAdminAction(adminID, adminName, "question_update", "question", &id, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				keySlug:       question.Slug,
				keyTheme:      question.Theme,
				keyDifficulty: question.Difficulty,
			}, true, nil); err != nil {
				qc.logger.Warn("Failed to log admin action", zap.Error(err))
			}
		}()
	}

	httputil.Success(c, http.StatusOK, question)
}

func (qc *AdminQuestionHandler) DeleteQuestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid question ID")
		return
	}

	question, err := qc.Usecase.GetQuestionByID(id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Question not found")
		return
	}

	err = qc.Usecase.DeleteQuestion(id)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	adminID := httputil.GetUserIDFromContext(c)
	if adminID != uuid.Nil {
		adminName := c.GetString("username")
		go func() {
			if err := qc.LoggingService.LogAdminAction(adminID, adminName, "question_delete", "question", &id, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				keySlug:       question.Slug,
				keyTheme:      question.Theme,
				keyDifficulty: question.Difficulty,
			}, true, nil); err != nil {
				qc.logger.Warn("Failed to log admin action", zap.Error(err))
			}
		}()
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "Question deleted successfully", nil)
}

func (qc *AdminQuestionHandler) ListQuestions(c *gin.Context) {
	pagination := pagination.Parse(c, pagination.AdminConfig())

	datasetIDStr := c.Query("dataset_id")
	themeStr := c.Query(keyTheme)
	subthemeStr := c.Query("subtheme")
	difficultyStr := c.Query(keyDifficulty)
	qtypeStr := c.Query("qtype")
	tagsStr := c.Query("tags")
	searchQuery := c.Query("search")

	filters := model.QuestionFilters{}

	if datasetIDStr != "" {
		parsed, err := uuid.Parse(datasetIDStr)
		if err != nil {
			httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset_id format")
			return
		}
		filters.DatasetID = &parsed
	} else {
		defaultDataset, err := qc.AdminDatasetsUsecase.GetDefaultDataset("question")
		if err != nil {
			qc.logger.Warn("No default dataset found, listing all questions", zap.Error(err))
		} else {
			filters.DatasetID = &defaultDataset.ID
		}
	}

	if themeStr != "" {
		filters.Theme = &themeStr
	}

	if subthemeStr != "" {
		filters.Subtheme = &subthemeStr
	}

	if difficultyStr != "" {
		filters.Difficulty = &difficultyStr
	}

	if qtypeStr != "" {
		filters.QType = &qtypeStr
	}

	if tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
		filters.Tags = tags
	}

	if searchQuery != "" {
		filters.SearchQuery = &searchQuery
	}

	var questions []*model.Question
	var total int64
	var err error

	if searchQuery != "" && len(filters.Tags) == 0 && filters.Theme == nil && filters.Subtheme == nil && filters.Difficulty == nil && filters.QType == nil {
		questions, total, err = qc.Usecase.SearchQuestions(searchQuery, filters.DatasetID, pagination.Limit, pagination.Offset)
	} else {
		questions, total, err = qc.Usecase.ListQuestionsWithFilters(filters, pagination.Limit, pagination.Offset)
	}

	if err != nil {
		qc.logger.Error("Failed to list questions", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	pagination.WithTotal(total)
	httputil.SuccessList(c, questions, httputil.ParamsToPagination(pagination.TotalCount, pagination.Limit, pagination.Offset))
}

func (qc *AdminQuestionHandler) BackupQuestions(c *gin.Context) {
	var datasetID *uuid.UUID
	if datasetIDStr := c.Query("dataset_id"); datasetIDStr != "" {
		parsed, err := uuid.Parse(datasetIDStr)
		if err != nil {
			httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset_id format")
			return
		}
		datasetID = &parsed
	}

	questions, err := qc.Usecase.BackupQuestions(datasetID)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	adminID := httputil.GetUserIDFromContext(c)
	if adminID != uuid.Nil {
		adminName := c.GetString("username")
		go func() {
			if err := qc.LoggingService.LogAdminAction(adminID, adminName, "questions_export", "question", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				"count": len(questions),
			}, true, nil); err != nil {
				qc.logger.Warn("Failed to log admin action", zap.Error(err))
			}
		}()
	}

	httputil.Success(c, http.StatusOK, questions)
}

func (qc *AdminQuestionHandler) ExportQuestionsClean(c *gin.Context) {
	questions, err := qc.Usecase.ExportQuestionsClean()
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	adminID := httputil.GetUserIDFromContext(c)
	if adminID != uuid.Nil {
		adminName := c.GetString("username")
		go func() {
			if err := qc.LoggingService.LogAdminAction(adminID, adminName, "questions_export_clean", "question", nil, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				"count": len(questions),
			}, true, nil); err != nil {
				qc.logger.Warn("Failed to log admin action", zap.Error(err))
			}
		}()
	}

	httputil.Success(c, http.StatusOK, questions)
}
