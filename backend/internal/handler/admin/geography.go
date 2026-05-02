// backend/internal/handler/admin/geography.go

package admin

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/pagination"
	"github.com/Culturae-org/culturae/internal/service"
	"github.com/Culturae-org/culturae/internal/usecase"
	adminUsecase "github.com/Culturae-org/culturae/internal/usecase/admin"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AdminGeographyHandler struct {
	AdminGeographyUsecase *adminUsecase.AdminGeographyUsecase
	GeographyUsecase      *usecase.GeographyUsecase
	LoggingService        service.LoggingServiceInterface
	wsService             service.WebSocketServiceInterface
	logger                *zap.Logger
}

func NewAdminGeographyHandler(
	adminGeographyUsecase *adminUsecase.AdminGeographyUsecase,
	geographyUsecase *usecase.GeographyUsecase,
	loggingService service.LoggingServiceInterface,
	wsService service.WebSocketServiceInterface,
	logger *zap.Logger,
) *AdminGeographyHandler {
	return &AdminGeographyHandler{
		AdminGeographyUsecase: adminGeographyUsecase,
		GeographyUsecase:      geographyUsecase,
		LoggingService:        loggingService,
		wsService:             wsService,
		logger:                logger,
	}
}

// -----------------------------------------------------
// Admin Geography Handlers
//
// - ListGeographyDatasets
// - GetDefaultGeographyDataset
// - GetGeographyDataset
// - GetGeographyDatasetBySlug
// - DeleteGeographyDataset
// - SetDefaultGeographyDataset
// - GetGeographyDatasetStatistics
// - ImportGeographyDataset
// - ListCountries
// - GetCountry
// - UpdateCountry
// - ListCountriesByContinent
// - ListCountriesByRegion
// - SearchCountries
// - ListContinents
// - GetContinent
// - UpdateContinent
// - ListRegions
// - ListRegionsByContinent
// - GetRegion
// - UpdateRegion
// - GetFlag
// - GetFlagURL
// -----------------------------------------------------

func (gc *AdminGeographyHandler) ListGeographyDatasets(c *gin.Context) {
	activeOnly := httputil.QueryBool(c, "active_only")
	active := false
	if activeOnly != nil {
		active = *activeOnly
	}

	datasets, err := gc.AdminGeographyUsecase.ListGeographyDatasets(active)
	if err != nil {
		gc.logger.Error("Failed to list geography datasets", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to list datasets")
		return
	}

	httputil.Success(c, http.StatusOK, datasets)
}

func (gc *AdminGeographyHandler) GetDefaultGeographyDataset(c *gin.Context) {
	dataset, err := gc.AdminGeographyUsecase.GetDefaultGeographyDataset()
	if err != nil {
		gc.logger.Error("Failed to get default geography dataset", zap.Error(err))
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "No default geography dataset found")
		return
	}

	httputil.Success(c, http.StatusOK, dataset)
}

func (gc *AdminGeographyHandler) GetGeographyDataset(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	dataset, err := gc.AdminGeographyUsecase.GetGeographyDataset(id)
	if err != nil {
		gc.logger.Error("Failed to get geography dataset", zap.String("id", id.String()), zap.Error(err))
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Geography dataset not found")
		return
	}

	httputil.Success(c, http.StatusOK, dataset)
}

func (gc *AdminGeographyHandler) GetGeographyDatasetBySlug(c *gin.Context) {
	slug := c.Param(keySlug)

	dataset, err := gc.AdminGeographyUsecase.GetGeographyDatasetBySlug(slug)
	if err != nil {
		gc.logger.Error("Failed to get geography dataset by slug", zap.String(keySlug, slug), zap.Error(err))
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Geography dataset not found")
		return
	}

	httputil.Success(c, http.StatusOK, dataset)
}

func (gc *AdminGeographyHandler) DeleteGeographyDataset(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	force := c.Query("force") == "true"

	gc.logger.Info("Deleting geography dataset",
		zap.String("id", id.String()),
		zap.Bool("force", force),
	)

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	realIP := httputil.GetRealIP(c)
	userAgent := httputil.GetUserAgent(c)

	if err := gc.AdminGeographyUsecase.DeleteGeographyDataset(id, force); err != nil {
		gc.logger.Error("Failed to delete geography dataset", zap.String("id", id.String()), zap.Error(err))

		if adminUUID != uuid.Nil {
			errorMsg := err.Error()
			go func() {
				httputil.LogAdminAction(gc.LoggingService, adminUUID, adminName, "geography_dataset_delete", "geography_dataset", &id, realIP, userAgent, map[string]interface{}{
					keyDatasetID: id,
					keyError:      errorMsg,
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

	dataset, _ := gc.AdminGeographyUsecase.GetGeographyDataset(id)

	if adminUUID != uuid.Nil {
		go func() {
			details := map[string]interface{}{
				keyDatasetID: id,
			}
			if dataset != nil {
				details[keyName] = dataset.Name
				details[keySlug] = dataset.Slug
			}
			httputil.LogAdminAction(gc.LoggingService, adminUUID, adminName, "geography_dataset_delete", "geography_dataset", &id, realIP, userAgent, details, true, nil)
		}()
	}

	gc.logger.Info("Geography dataset deleted successfully", zap.String("id", id.String()))
	c.Status(http.StatusNoContent)
}

func (gc *AdminGeographyHandler) SetDefaultGeographyDataset(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	realIP := httputil.GetRealIP(c)
	userAgent := httputil.GetUserAgent(c)

	dataset, err := gc.AdminGeographyUsecase.GetGeographyDataset(id)
	if err != nil {
		gc.logger.Error("Failed to get geography dataset for default setting", zap.String("id", id.String()), zap.Error(err))

		if adminUUID != uuid.Nil {
			errorMsg := "Dataset not found"
			go func() {
				httputil.LogAdminAction(gc.LoggingService, adminUUID, adminName, "geography_dataset_set_default", "geography_dataset", &id, realIP, userAgent, map[string]interface{}{
					keyDatasetID: id,
					keyError:      errorMsg,
				}, false, &errorMsg)
			}()
		}

		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Geography dataset not found for default setting")
		return
	}

	if err := gc.AdminGeographyUsecase.SetDefaultGeographyDataset(id); err != nil {
		gc.logger.Error("Failed to set default geography dataset", zap.String("id", id.String()), zap.Error(err))

		if adminUUID != uuid.Nil {
			errorMsg := err.Error()
			go func() {
				httputil.LogAdminAction(gc.LoggingService, adminUUID, adminName, "geography_dataset_set_default", "geography_dataset", &id, realIP, userAgent, map[string]interface{}{
					keyDatasetID: id,
					keyName:       dataset.Name,
					keySlug:       dataset.Slug,
					keyError:      errorMsg,
				}, false, &errorMsg)
			}()
		}

		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	if adminUUID != uuid.Nil {
		go func() {
			httputil.LogAdminAction(gc.LoggingService, adminUUID, adminName, "geography_dataset_set_default", "geography_dataset", &id, realIP, userAgent, map[string]interface{}{
				keyName: dataset.Name,
				keySlug: dataset.Slug,
			}, true, nil)
		}()
	}

	httputil.SuccessWithMessage(c, http.StatusOK, "Default geography dataset updated", nil)
}

func (gc *AdminGeographyHandler) GetGeographyDatasetStatistics(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	stats, err := gc.AdminGeographyUsecase.GetGeographyDatasetStatistics(id)
	if err != nil {
		gc.logger.Error("Failed to get geography dataset statistics", zap.String("id", id.String()), zap.Error(err))
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, stats)
}

func (gc *AdminGeographyHandler) ImportGeographyDataset(c *gin.Context) {
	var req model.ImportDatasetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	req.DatasetType = datasetTypeGeography

	gc.logger.Info("Starting Geography import from manifest",
		zap.String(keyManifestURL, req.ManifestURL),
		zap.String(keyDatasetType, req.DatasetType),
		zap.Bool("set_as_default", req.SetAsDefault),
		zap.Bool("force", req.Force),
	)

	adminUUID := httputil.GetUserIDFromContext(c)
	adminName := c.GetString("username")
	realIP := httputil.GetRealIP(c)
	userAgent := httputil.GetUserAgent(c)

	result, err := gc.AdminGeographyUsecase.ImportGeographyFromManifest(req.ManifestURL)
	if err != nil {
		gc.logger.Error("Failed to import geography from manifest", zap.Error(err))

		if adminUUID != uuid.Nil {
			errorMsg := err.Error()
			go func() {
				httputil.LogAdminAction(gc.LoggingService, adminUUID, adminName, "geography_dataset_import", "geography_dataset", nil, realIP, userAgent, map[string]interface{}{
					keyManifestURL: req.ManifestURL,
					keyError:        errorMsg,
					keyErrorType:   "import_failed",
				}, false, &errorMsg)
			}()
		}

		if strings.Contains(err.Error(), "already exists") {
			httputil.Error(c, http.StatusConflict, httputil.ErrCodeConflict, err.Error())
			return
		}
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	if req.SetAsDefault {
		dataset, err := gc.AdminGeographyUsecase.GetGeographyDatasetByManifestURL(req.ManifestURL)
		if err != nil {
			gc.logger.Error("Failed to find imported geography dataset to set as default", zap.Error(err))

			if adminUUID != uuid.Nil {
				errorMsg := err.Error()
				go func() {
					httputil.LogAdminAction(gc.LoggingService, adminUUID, adminName, "geography_dataset_import", "geography_dataset", nil, realIP, userAgent, map[string]interface{}{
						keyManifestURL: req.ManifestURL,
						keyError:        errorMsg,
						keyErrorType:   "find_imported_dataset_failed",
					}, false, &errorMsg)
				}()
			}

			httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to set geography dataset as default")
			return
		}

		if err := gc.AdminGeographyUsecase.SetDefaultGeographyDataset(dataset.ID); err != nil {
			gc.logger.Error("Failed to set geography dataset as default", zap.Error(err))

			if adminUUID != uuid.Nil {
				errorMsg := err.Error()
				go func() {
					httputil.LogAdminAction(gc.LoggingService, adminUUID, adminName, "geography_dataset_import", "geography_dataset", nil, realIP, userAgent, map[string]interface{}{
						keyManifestURL: req.ManifestURL,
						keyDatasetID:   dataset.ID,
						keyError:        errorMsg,
						keyErrorType:   "set_default_failed",
					}, false, &errorMsg)
				}()
			}

			httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to set geography dataset as default")
			return
		}

		gc.logger.Info("Geography dataset set as default", zap.String(keyDatasetID, dataset.ID.String()))
	}

	if adminUUID != uuid.Nil {
		go func() {
			httputil.LogAdminAction(gc.LoggingService, adminUUID, adminName, "geography_dataset_import", "geography_dataset", nil, realIP, userAgent, map[string]interface{}{
				keyManifestURL:     req.ManifestURL,
				"countries_added":  result.CountriesAdded,
				"continents_added": result.ContinentsAdded,
				"regions_added":    result.RegionsAdded,
				"set_as_default":   req.SetAsDefault,
			}, true, nil)
		}()
	}

	gc.wsService.BroadcastAdminNotification(service.AdminNotification{
		Event: "geography_imported",
		Data: map[string]interface{}{
			keyDatasetType:     "geography",
			keyAdminName:       c.GetString("username"),
			"countries_added":  result.CountriesAdded,
			"continents_added": result.ContinentsAdded,
			"regions_added":    result.RegionsAdded,
		},
		EntityType: "dataset",
		ActionURL:  "/geography",
	})

	httputil.Success(c, http.StatusOK, result)
}

func (gc *AdminGeographyHandler) ListCountries(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	pageParams := pagination.Parse(c, pagination.Config{
		DefaultLimit: 50,
		MaxLimit:     300,
	})

	var filters model.CountryFilters
	if continent := httputil.QueryString(c, "continent"); continent != nil {
		filters.Continent = *continent
	}
	if region := httputil.QueryString(c, "region"); region != nil {
		filters.Region = *region
	}
	if drivingSide := httputil.QueryString(c, "driving_side"); drivingSide != nil {
		filters.DrivingSide = *drivingSide
	}

	if popMin := c.Query("population_min"); popMin != "" {
		if val, err := strconv.ParseInt(popMin, 10, 64); err == nil {
			filters.PopulationMin = &val
		}
	}
	if popMax := c.Query("population_max"); popMax != "" {
		if val, err := strconv.ParseInt(popMax, 10, 64); err == nil {
			filters.PopulationMax = &val
		}
	}
	if areaMin := c.Query("area_min"); areaMin != "" {
		if val, err := strconv.ParseInt(areaMin, 10, 64); err == nil {
			filters.AreaMin = &val
		}
	}
	if areaMax := c.Query("area_max"); areaMax != "" {
		if val, err := strconv.ParseInt(areaMax, 10, 64); err == nil {
			filters.AreaMax = &val
		}
	}
	if independent := httputil.QueryBool(c, "independent"); independent != nil {
		filters.Independent = independent
	}

	hasFilters := filters.Continent != "" || filters.Region != "" || filters.DrivingSide != "" ||
		filters.PopulationMin != nil || filters.PopulationMax != nil ||
		filters.AreaMin != nil || filters.AreaMax != nil || filters.Independent != nil

	var countries []model.Country
	var total int64

	if hasFilters {
		countries, total, err = gc.GeographyUsecase.ListCountriesWithFilters(id, filters, pageParams.Limit, pageParams.Offset)
	} else {
		countries, total, err = gc.GeographyUsecase.ListCountries(id, pageParams.Limit, pageParams.Offset)
	}

	if err != nil {
		gc.logger.Error("Failed to list countries", zap.String(keyDatasetID, id.String()), zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	pageParams.WithTotal(total)
	httputil.SuccessList(c, countries, httputil.ParamsToPagination(pageParams.TotalCount, pageParams.Limit, pageParams.Offset))
}

func (gc *AdminGeographyHandler) GetCountry(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	slug := c.Param(keySlug)

	country, err := gc.GeographyUsecase.GetCountryBySlug(slug, id)
	if err != nil {
		gc.logger.Error("Failed to get country", zap.String(keySlug, slug), zap.Error(err))
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Country not found")
		return
	}

	httputil.Success(c, http.StatusOK, country)
}

func (gc *AdminGeographyHandler) UpdateCountry(c *gin.Context) {
	datasetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	slug := c.Param(keySlug)

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	country, err := gc.AdminGeographyUsecase.UpdateCountry(datasetID, slug, updates)
	if err != nil {
		gc.logger.Error("Failed to update country", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	adminID := httputil.GetUserIDFromContext(c)
	if adminID != uuid.Nil {
		adminName := c.GetString("username")
		go func() {
			if err := gc.LoggingService.LogAdminAction(adminID, adminName, "geography_country_update", "country", &country.ID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				keySlug:    slug,
				"updates": updates,
			}, true, nil); err != nil {
				gc.logger.Warn("Failed to log admin action", zap.Error(err))
			}
		}()
	}

	httputil.Success(c, http.StatusOK, country)
}

func (gc *AdminGeographyHandler) ListCountriesByContinent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	continent := c.Param("continent")

	countries, err := gc.GeographyUsecase.ListCountriesByContinent(id, continent)
	if err != nil {
		gc.logger.Error("Failed to list countries by continent", zap.String("continent", continent), zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, countries)
}

func (gc *AdminGeographyHandler) ListCountriesByRegion(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	region := c.Param("region")

	countries, err := gc.GeographyUsecase.ListCountriesByRegion(id, region)
	if err != nil {
		gc.logger.Error("Failed to list countries by region", zap.String("region", region), zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, countries)
}

func (gc *AdminGeographyHandler) SearchCountries(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	query := c.Query("q")
	if query == "" {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeMissingField, "Query parameter 'q' is required")
		return
	}

	pageParams := pagination.Parse(c, pagination.Config{
		DefaultLimit: 50,
		MaxLimit:     100,
	})

	countries, total, err := gc.GeographyUsecase.SearchCountries(id, query, pageParams.Limit, pageParams.Offset)
	if err != nil {
		gc.logger.Error("Failed to search countries", zap.String("query", query), zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	pageParams.WithTotal(total)
	httputil.SuccessList(c, countries, httputil.ParamsToPagination(pageParams.TotalCount, pageParams.Limit, pageParams.Offset))
}

func (gc *AdminGeographyHandler) ListContinents(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	continents, err := gc.GeographyUsecase.ListContinents(id)
	if err != nil {
		gc.logger.Error("Failed to list continents", zap.String(keyDatasetID, id.String()), zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, continents)
}

func (gc *AdminGeographyHandler) GetContinent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	slug := c.Param(keySlug)

	continent, err := gc.GeographyUsecase.GetContinentBySlug(slug, id)
	if err != nil {
		gc.logger.Error("Failed to get continent", zap.String(keySlug, slug), zap.Error(err))
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Continent not found")
		return
	}

	httputil.Success(c, http.StatusOK, continent)
}

func (gc *AdminGeographyHandler) UpdateContinent(c *gin.Context) {
	datasetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	slug := c.Param(keySlug)

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	continent, err := gc.AdminGeographyUsecase.UpdateContinent(datasetID, slug, updates)
	if err != nil {
		gc.logger.Error("Failed to update continent", zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	adminID := httputil.GetUserIDFromContext(c)
	if adminID != uuid.Nil {
		adminName := c.GetString("username")
		go func() {
			if err := gc.LoggingService.LogAdminAction(adminID, adminName, "geography_continent_update", "continent", &continent.ID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				keySlug:    slug,
				"updates": updates,
			}, true, nil); err != nil {
				gc.logger.Warn("Failed to log admin action", zap.Error(err))
			}
		}()
	}

	httputil.Success(c, http.StatusOK, continent)
}

func (gc *AdminGeographyHandler) ListRegions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	regions, err := gc.GeographyUsecase.ListRegions(id)
	if err != nil {
		gc.logger.Error("Failed to list regions", zap.String(keyDatasetID, id.String()), zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, regions)
}

func (gc *AdminGeographyHandler) ListRegionsByContinent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	continent := c.Param("continent")

	regions, err := gc.GeographyUsecase.ListRegionsByContinent(id, continent)
	if err != nil {
		gc.logger.Error("Failed to list regions by continent", zap.String("continent", continent), zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	httputil.Success(c, http.StatusOK, regions)
}

func (gc *AdminGeographyHandler) GetRegion(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	slug := c.Param(keySlug)

	region, err := gc.GeographyUsecase.GetRegionBySlug(slug, id)
	if err != nil {
		gc.logger.Error("Failed to get region", zap.String(keySlug, slug), zap.Error(err))
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Region not found")
		return
	}

	httputil.Success(c, http.StatusOK, region)
}

func (gc *AdminGeographyHandler) UpdateRegion(c *gin.Context) {
	datasetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Invalid dataset ID")
		return
	}

	slug := c.Param(keySlug)

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeValidation, err.Error())
		return
	}

	region, err := gc.AdminGeographyUsecase.UpdateRegion(datasetID, slug, updates)
	if err != nil {
		gc.logger.Error("Failed to update region", zap.String(keySlug, slug), zap.Error(err))
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, err.Error())
		return
	}

	adminID := httputil.GetUserIDFromContext(c)
	if adminID != uuid.Nil {
		adminName := c.GetString("username")
		go func() {
			if err := gc.LoggingService.LogAdminAction(adminID, adminName, "geography_region_update", "region", &region.ID, httputil.GetRealIP(c), httputil.GetUserAgent(c), map[string]interface{}{
				keySlug:    slug,
				"updates": updates,
			}, true, nil); err != nil {
				gc.logger.Warn("Failed to log admin action", zap.Error(err))
			}
		}()
	}

	httputil.Success(c, http.StatusOK, region)
}

func (gc *AdminGeographyHandler) GetFlag(c *gin.Context) {
	countryCode := c.Param("country_code")
	if countryCode == "" {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeMissingField, "Country code is required")
		return
	}

	content, contentType, err := gc.GeographyUsecase.GetFlagFile(countryCode)
	if err != nil {
		gc.logger.Debug("Flag not found", zap.String("country", countryCode), zap.Error(err))
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Flag not found")
		return
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, contentType, content)
}

func (gc *AdminGeographyHandler) GetFlagURL(c *gin.Context) {
	countryCode := c.Param("country_code")
	if countryCode == "" {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeMissingField, "Country code is required")
		return
	}

	url, err := gc.GeographyUsecase.GetFlagURL(countryCode)
	if err != nil {
		gc.logger.Debug("Flag URL not found", zap.String("country", countryCode), zap.Error(err))
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Flag not found")
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{
		"country_code": countryCode,
		"url":          url,
	})
}
