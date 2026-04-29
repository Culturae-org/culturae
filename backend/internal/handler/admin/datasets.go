// backend/internal/handler/admin/datasets.go

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

const (
	datasetTypeQuestions = "questions"
	datasetTypeGeography = "geography"
)

type AdminDatasetsHandler struct {
	AdminDatasetsUsecase  *adminUsecase.AdminDatasetsUsecase
	AdminImportsUsecase   *adminUsecase.AdminImportsUsecase
	AdminQuestionsUsecase *adminUsecase.AdminQuestionsUsecase
	AdminGeographyUsecase *adminUsecase.AdminGeographyUsecase
	LoggingService        service.LoggingServiceInterface
	wsService             service.WebSocketServiceInterface
	logger                *zap.Logger
}

func NewAdminDatasetsHandler(
	adminDatasetsUsecase *adminUsecase.AdminDatasetsUsecase,
	adminImportsUsecase *adminUsecase.AdminImportsUsecase,
	adminQuestionsUsecase *adminUsecase.AdminQuestionsUsecase,
	adminGeographyUsecase *adminUsecase.AdminGeographyUsecase,
	loggingService service.LoggingServiceInterface,
	wsService service.WebSocketServiceInterface,
	logger *zap.Logger,
) *AdminDatasetsHandler {
	return &AdminDatasetsHandler{
		AdminDatasetsUsecase:  adminDatasetsUsecase,
		AdminImportsUsecase:   adminImportsUsecase,
		AdminQuestionsUsecase: adminQuestionsUsecase,
		AdminGeographyUsecase: adminGeographyUsecase,
		LoggingService:        loggingService,
		wsService:             wsService,
		logger:                logger,
	}
}

// -----------------------------------------------------
// Admin Datasets Handlers
//
// - ListDatasets
// - GetDataset
// - GetDatasetBySlug
// - CreateDataset
// - UpdateDataset
// - DeleteDataset
// - SetDefaultDataset
// - GetDefaultDataset
// - CheckForUpdates
// - ImportDataset
// - UpdateDatasetStatistics
// - GetDatasetStatistics
// - GetDatasetQuestions
// - GetHistory
// -----------------------------------------------------

func (dc *AdminDatasetsHandler) ListDatasets(c *gin.Context) {
	pag := pagination.Parse(c, pagination.Config{
		DefaultLimit: 10,
		MaxLimit:     50,
	})

	activeOnlyPtr := httputil.QueryBool(c, "active_only")
	activeOnly := activeOnlyPtr != nil && *activeOnlyPtr

	defaultOnlyPtr := httputil.QueryBool(c, "default_only")
	defaultOnly := defaultOnlyPtr != nil && *defaultOnlyPtr

	datasetType := c.Query("dataset_type")

	var allDatasets []model.UnifiedDataset

	switch datasetType {
	case datasetTypeGeography:
		geography, err := dc.AdminGeographyUsecase.ListGeographyDatasets(activeOnly)
		if err != nil {
			dc.logger.Error("Failed to list geography datasets", zap.Error(err))
			httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to list datasets")
			return
		}
		for _, gd := range geography {
			allDatasets = append(allDatasets, model.UnifiedDataset{
				ID:                     gd.ID,
				Type:                   datasetTypeGeography,
				Slug:                   gd.Slug,
				Name:                   gd.Name,
				Description:            gd.Description,
				Version:                gd.Version,
				Source:                 gd.Source,
				ManifestURL:            gd.ManifestURL,
				IsActive:               gd.IsActive,
				IsDefault:              gd.IsDefault,
				ImportedAt:             gd.ImportedAt,
				LatestAvailableVersion: gd.LatestAvailableVersion,
				QuestionCount:          gd.CountryCount,
				ThemeCount:             gd.ContinentCount,
				RegionCount:            gd.RegionCount,
			})
		}

	case "questions":
		questions, err := dc.AdminDatasetsUsecase.ListQuestionDatasets(activeOnly)
		if err != nil {
			dc.logger.Error("Failed to list question datasets", zap.Error(err))
			httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to list datasets")
			return
		}
		for _, qd := range questions {
			allDatasets = append(allDatasets, model.UnifiedDataset{
				ID:                     qd.ID,
				Type:                   "questions",
				Slug:                   qd.Slug,
				Name:                   qd.Name,
				Description:            qd.Description,
				Version:                qd.Version,
				Source:                 qd.Source,
				ManifestURL:            qd.ManifestURL,
				IsActive:               qd.IsActive,
				IsDefault:              qd.IsDefault,
				ImportedAt:             qd.ImportedAt,
				LatestAvailableVersion: qd.LatestAvailableVersion,
				QuestionCount:          qd.QuestionCount,
				ThemeCount:             qd.ThemeCount,
			})
		}

	case "", "all":
		questions, qErr := dc.AdminDatasetsUsecase.ListQuestionDatasets(activeOnly)
		geography, gErr := dc.AdminGeographyUsecase.ListGeographyDatasets(activeOnly)

		if qErr != nil || gErr != nil {
			dc.logger.Error("Failed to list datasets", zap.Error(qErr), zap.Error(gErr))
			httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to list datasets")
			return
		}

		for _, qd := range questions {
			allDatasets = append(allDatasets, model.UnifiedDataset{
				ID:                     qd.ID,
				Type:                   "questions",
				Slug:                   qd.Slug,
				Name:                   qd.Name,
				Description:            qd.Description,
				Version:                qd.Version,
				Source:                 qd.Source,
				ManifestURL:            qd.ManifestURL,
				IsActive:               qd.IsActive,
				IsDefault:              qd.IsDefault,
				ImportedAt:             qd.ImportedAt,
				LatestAvailableVersion: qd.LatestAvailableVersion,
				QuestionCount:          qd.QuestionCount,
				ThemeCount:             qd.ThemeCount,
			})
		}

		for _, gd := range geography {
			allDatasets = append(allDatasets, model.UnifiedDataset{
				ID:                     gd.ID,
				Type:                   datasetTypeGeography,
				Slug:                   gd.Slug,
				Name:                   gd.Name,
				Description:            gd.Description,
				Version:                gd.Version,
				Source:                 gd.Source,
				ManifestURL:            gd.ManifestURL,
				IsActive:               gd.IsActive,
				IsDefault:              gd.IsDefault,
				ImportedAt:             gd.ImportedAt,
				LatestAvailableVersion: gd.LatestAvailableVersion,
				CountryCount:           gd.CountryCount,
				ContinentCount:         gd.ContinentCount,
				RegionCount:            gd.RegionCount,
			})
		}

	default:
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, "Invalid dataset_type (supported: questions, geography, all)")
		return
	}

	if defaultOnly {
		var filtered []model.UnifiedDataset
		for _, ds := range allDatasets {
			if ds.IsDefault {
				filtered = append(filtered, ds)
			}
		}
		allDatasets = filtered
	}

	total := len(allDatasets)

	start := pag.Offset
	end := start + pag.Limit
	if end > total {
		end = total
	}

	var paginatedDatasets []model.UnifiedDataset
	if start < total {
		paginatedDatasets = allDatasets[start:end]
	} else {
		paginatedDatasets = []model.UnifiedDataset{}
	}

	pag.WithTotal(int64(total))
	httputil.SuccessList(c, paginatedDatasets, httputil.ParamsToPagination(pag.TotalCount, pag.Limit, pag.Offset))
}

func (dc *AdminDatasetsHandler) GetDataset(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	dataset, err := dc.AdminDatasetsUsecase.GetQuestionDataset(id)
	if err != nil {
		dc.logger.Error("Failed to get dataset", zap.String("id", idParam), zap.Error(err))
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Dataset not found")
		return
	}

	httputil.Success(c, http.StatusOK, dataset)
}

func (dc *AdminDatasetsHandler) GetDatasetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	qDataset, qErr := dc.AdminDatasetsUsecase.GetQuestionDatasetBySlug(slug)
	if qErr == nil {
		httputil.Success(c, http.StatusOK, qDataset)
		return
	}

	gDataset, gErr := dc.AdminGeographyUsecase.GetGeographyDatasetBySlug(slug)
	if gErr == nil {
		httputil.Success(c, http.StatusOK, gDataset)
		return
	}

	dc.logger.Error("Failed to get dataset by slug", zap.String("slug", slug), zap.Error(qErr), zap.Error(gErr))
	httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Dataset not found")
}

func (dc *AdminDatasetsHandler) CreateDataset(c *gin.Context) {
	var req model.CreateDatasetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	dataset, err := dc.AdminDatasetsUsecase.CreateQuestionDataset(&req)
	if err != nil {
		dc.logger.Error("Failed to create dataset", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	if adminUUID != uuid.Nil {
		go func() {
			httputil.LogAdminAction(dc.LoggingService, adminUUID, adminName, "dataset_create", "dataset", &dataset.ID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				"name":    dataset.Name,
				"slug":    dataset.Slug,
				"version": dataset.Version,
			}, true, nil)
		}()
	}

	httputil.Success(c, http.StatusCreated, dataset)
}

func (dc *AdminDatasetsHandler) UpdateDataset(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	var req model.UpdateDatasetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	dataset, err := dc.AdminDatasetsUsecase.UpdateQuestionDataset(id, &req)
	if err != nil {
		dc.logger.Error("Failed to update dataset", zap.String("id", idParam), zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	if adminUUID != uuid.Nil {
		go func() {
			httputil.LogAdminAction(dc.LoggingService, adminUUID, adminName, "dataset_update", "dataset", &id, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				"name":    dataset.Name,
				"slug":    dataset.Slug,
				"version": dataset.Version,
			}, true, nil)
		}()
	}

	httputil.Success(c, http.StatusOK, dataset)
}

func (dc *AdminDatasetsHandler) DeleteDataset(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	force := c.Query("force") == "true"

	dc.logger.Info("Deleting dataset",
		zap.String("type", "questions"),
		zap.String("id", idParam),
		zap.Bool("force", force),
	)

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	realIP := httputil.GetRealIP(c)
	userAgent := httputil.GetUserAgent(c)

	if err := dc.AdminDatasetsUsecase.DeleteDataset("questions", id, force); err != nil {
		dc.logger.Error("Failed to delete dataset", zap.String("id", idParam), zap.Error(err))

		if adminUUID != uuid.Nil {
			errorMsg := err.Error()
			go func() {
				httputil.LogAdminAction(dc.LoggingService, adminUUID, adminName, "dataset_delete", "dataset", &id, realIP, userAgent, map[string]interface{}{
					"dataset_id": id,
					"error":      errorMsg,
				}, false, &errorMsg)
			}()
		}

		if strings.Contains(err.Error(), "cannot delete the only dataset") {
			httputil.Error(c, http.StatusConflict, httputil.ErrCodeConflict, "cannot delete the only dataset of this type")
			return
		}
		if strings.Contains(err.Error(), "cannot delete the default dataset") {
			httputil.Error(c, http.StatusConflict, httputil.ErrCodeConflict, "cannot delete the default dataset")
			return
		}
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	dataset, _ := dc.AdminDatasetsUsecase.GetQuestionDataset(id)

	if adminUUID != uuid.Nil {
		go func() {
			details := map[string]interface{}{
				"dataset_id": id,
			}
			if dataset != nil {
				details["name"] = dataset.Name
				details["slug"] = dataset.Slug
			}
			httputil.LogAdminAction(dc.LoggingService, adminUUID, adminName, "dataset_delete", "dataset", &id, realIP, userAgent, details, true, nil)
		}()
	}

	dc.logger.Info("Dataset deleted successfully", zap.String("id", idParam))
	c.Status(http.StatusNoContent)
}

func (dc *AdminDatasetsHandler) SetDefaultDataset(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	realIP := httputil.GetRealIP(c)
	userAgent := httputil.GetUserAgent(c)

	dataset, err := dc.AdminDatasetsUsecase.GetQuestionDataset(id)
	if err != nil {
		dc.logger.Error("Failed to get dataset for default setting logging", zap.String("id", idParam), zap.Error(err))

		if adminUUID != uuid.Nil {
			errorMsg := "Dataset not found"
			go func() {
				httputil.LogAdminAction(dc.LoggingService, adminUUID, adminName, "dataset_set_default", "dataset", &id, realIP, userAgent, map[string]interface{}{
					"dataset_id": id,
					"error":      errorMsg,
				}, false, &errorMsg)
			}()
		}

		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Dataset not found for default setting")
		return
	}

	if err := dc.AdminDatasetsUsecase.SetDefaultDataset("questions", id); err != nil {
		dc.logger.Error("Failed to set default dataset", zap.String("id", idParam), zap.Error(err))

		if adminUUID != uuid.Nil {
			errorMsg := err.Error()
			go func() {
				httputil.LogAdminAction(dc.LoggingService, adminUUID, adminName, "dataset_set_default", "dataset", &id, realIP, userAgent, map[string]interface{}{
					"dataset_id": id,
					"name":       dataset.Name,
					"slug":       dataset.Slug,
					"error":      errorMsg,
				}, false, &errorMsg)
			}()
		}

		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	if adminUUID != uuid.Nil {
		go func() {
			httputil.LogAdminAction(dc.LoggingService, adminUUID, adminName, "dataset_set_default", "dataset", &id, realIP, userAgent, map[string]interface{}{
				"name": dataset.Name,
				"slug": dataset.Slug,
			}, true, nil)
		}()
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "Default dataset updated", nil)
}

func (dc *AdminDatasetsHandler) GetDefaultDataset(c *gin.Context) {
	dataset, err := dc.AdminDatasetsUsecase.GetDefaultDataset("questions")
	if err != nil {
		dc.logger.Error("Failed to get default dataset", zap.Error(err))
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "No default dataset found")
		return
	}

	httputil.Success(c, http.StatusOK, dataset)
}

func (dc *AdminDatasetsHandler) CheckForUpdates(c *gin.Context) {
	updatesInfo, err := dc.AdminDatasetsUsecase.CheckAllDatasetsForUpdates()
	if err != nil {
		dc.logger.Error("Failed to check for updates", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, updatesInfo)
}

func (dc *AdminDatasetsHandler) ImportDataset(c *gin.Context) {
	var req model.ImportDatasetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	dc.logger.Info("Starting dataset import from manifest",
		zap.String("manifest_url", req.ManifestURL),
		zap.String("dataset_type", req.DatasetType),
		zap.Bool("set_as_default", req.SetAsDefault),
		zap.Bool("force", req.Force),
	)

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	realIP := httputil.GetRealIP(c)
	userAgent := httputil.GetUserAgent(c)

	result, err := dc.AdminImportsUsecase.ImportFromManifest(req.ManifestURL)
	if err != nil {
		dc.logger.Error("Failed to import dataset from manifest", zap.Error(err))

		if adminUUID != uuid.Nil {
			errorMsg := err.Error()
			go func() {
				httputil.LogAdminAction(dc.LoggingService, adminUUID, adminName, "dataset_import", "dataset", nil, realIP, userAgent, map[string]interface{}{
					"manifest_url": req.ManifestURL,
					"error":        errorMsg,
					"error_type":   "import_failed",
				}, false, &errorMsg)
			}()
		}

		if strings.Contains(err.Error(), "already imported") || strings.Contains(err.Error(), "already exists") {
			httputil.Error(c, http.StatusConflict, httputil.ErrCodeConflict, err.Error())
			return
		}
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	if req.SetAsDefault {
		dataset, err := dc.AdminDatasetsUsecase.GetQuestionDatasetByManifestURL(req.ManifestURL)
		if err != nil {
			dc.logger.Error("Failed to find imported dataset to set as default", zap.Error(err))

			if adminUUID != uuid.Nil {
				errorMsg := err.Error()
				go func() {
					httputil.LogAdminAction(dc.LoggingService, adminUUID, adminName, "dataset_import", "dataset", nil, realIP, userAgent, map[string]interface{}{
						"manifest_url": req.ManifestURL,
						"error":        errorMsg,
						"error_type":   "find_imported_dataset_failed",
					}, false, &errorMsg)
				}()
			}

			httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to set dataset as default")
			return
		}

		if err := dc.AdminDatasetsUsecase.SetDefaultDataset("questions", dataset.ID); err != nil {
			dc.logger.Error("Failed to set dataset as default", zap.Error(err))

			if adminUUID != uuid.Nil {
				errorMsg := err.Error()
				go func() {
					httputil.LogAdminAction(dc.LoggingService, adminUUID, adminName, "dataset_import", "dataset", nil, realIP, userAgent, map[string]interface{}{
						"manifest_url": req.ManifestURL,
						"dataset_id":   dataset.ID,
						"error":        errorMsg,
						"error_type":   "set_default_failed",
					}, false, &errorMsg)
				}()
			}

			httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to set dataset as default")
			return
		}

		dc.logger.Info("Dataset set as default", zap.String("dataset_id", dataset.ID.String()))
	}

	if adminUUID != uuid.Nil {
		go func() {
			logData := map[string]interface{}{
				"manifest_url":   req.ManifestURL,
				"set_as_default": req.SetAsDefault,
			}

			if req.DatasetType == "geography" {
				if geoResult, ok := result.(*model.GeographyImportResult); ok {
					logData["countries_added"] = geoResult.CountriesAdded
					logData["continents_added"] = geoResult.ContinentsAdded
					logData["regions_added"] = geoResult.RegionsAdded
				}
			} else {
				if qResult, ok := result.(*model.ImportResult); ok {
					logData["questions_added"] = qResult.QuestionsAdded
					logData["questions_updated"] = qResult.QuestionsUpdated
					logData["questions_skipped"] = qResult.QuestionsSkipped
				}
			}

			httputil.LogAdminAction(dc.LoggingService, adminUUID, adminName, "dataset_import", "dataset", nil, realIP, userAgent, logData, true, nil)
		}()
	}

	notificationData := map[string]interface{}{
		"dataset_type": req.DatasetType,
		"admin_name":   c.GetString("username"),
	}

	if req.DatasetType == "geography" {
		if geoResult, ok := result.(*model.GeographyImportResult); ok {
			notificationData["countries_added"] = geoResult.CountriesAdded
			notificationData["continents_added"] = geoResult.ContinentsAdded
			notificationData["regions_added"] = geoResult.RegionsAdded
		}
	} else {
		if qResult, ok := result.(*model.ImportResult); ok {
			notificationData["questions_added"] = qResult.QuestionsAdded
			notificationData["questions_updated"] = qResult.QuestionsUpdated
			notificationData["questions_skipped"] = qResult.QuestionsSkipped
		}
	}

	dc.wsService.BroadcastAdminNotification(service.AdminNotification{
		Event:      "dataset_imported",
		Data:       notificationData,
		EntityType: "dataset",
		ActionURL:  "/questions",
	})

	httputil.Success(c, http.StatusOK, result)
}

func (dc *AdminDatasetsHandler) UpdateDatasetStatistics(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	adminUUID, adminName := httputil.GetUserIDFromContext(c), c.GetString("username")
	if err := dc.AdminDatasetsUsecase.UpdateQuestionDatasetStatistics(id); err != nil {
		dc.logger.Error("Failed to update dataset statistics", zap.String("id", idParam), zap.Error(err))
		errMsg := err.Error()
		httputil.LogAdminAction(dc.LoggingService, adminUUID, adminName, "dataset_update_stats", "dataset", &id, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{"dataset_id": id}, false, &errMsg)
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.LogAdminAction(dc.LoggingService, adminUUID, adminName, "dataset_update_stats", "dataset", &id, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{"dataset_id": id}, true, nil)
	httputil.SuccessWithMessage(c, http.StatusOK, "Statistics updated", nil)
}

func (dc *AdminDatasetsHandler) GetDatasetStatistics(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	stats, err := dc.AdminDatasetsUsecase.GetDatasetStatistics("questions", id)
	if err != nil {
		dc.logger.Error("Failed to get dataset statistics", zap.String("id", idParam), zap.Error(err))
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, stats)
}

func (dc *AdminDatasetsHandler) GetDatasetQuestions(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	pag := pagination.Parse(c, pagination.Config{
		DefaultLimit: 50,
		MaxLimit:     200,
	})

	questions, total, err := dc.AdminDatasetsUsecase.GetQuestionDatasetQuestions(id, pag.Limit, pag.Offset)
	if err != nil {
		dc.logger.Error("Failed to get dataset questions", zap.String("id", idParam), zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	pag.WithTotal(total)
	httputil.SuccessList(c, questions, httputil.ParamsToPagination(pag.TotalCount, pag.Limit, pag.Offset))
}

func (dc *AdminDatasetsHandler) GetHistory(c *gin.Context) {
	pag := pagination.Parse(c, pagination.Config{
		DefaultLimit: 20,
		MaxLimit:     100,
	})

	typeFilter := c.Query("type")
	var typeFilterPtr *string
	if typeFilter != "" {
		typeFilterPtr = &typeFilter
	}

	events, total, err := dc.AdminDatasetsUsecase.ListHistory(pag.Limit, pag.Offset, typeFilterPtr)
	if err != nil {
		dc.logger.Error("Failed to list history", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to list history")
		return
	}

	pag.WithTotal(total)
	httputil.SuccessList(c, events, httputil.ParamsToPagination(pag.TotalCount, pag.Limit, pag.Offset))
}
