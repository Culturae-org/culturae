// backend/internal/handler/geography.go

package handler

import (
	"net/http"
	"strconv"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/usecase"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GeographyHandler struct {
	GeographyUsecase *usecase.GeographyUsecase
	logger           *zap.Logger
}

func NewGeographyHandler(
	geographyUsecase *usecase.GeographyUsecase,
	logger *zap.Logger,
) *GeographyHandler {
	return &GeographyHandler{
		GeographyUsecase: geographyUsecase,
		logger:           logger,
	}
}

// -----------------------------------------------------
// Geography Handlers
//
// - GetFlagPNG
// - GetFlag
// - GetFlagURL
// - GetCountries
// - GetContinents
// -----------------------------------------------------

func (gc *GeographyHandler) GetFlagPNG(c *gin.Context) {
	countryCode := c.Param("country_code")
	format := c.Param("format")
	if countryCode == "" {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeMissingField, "Country code is required")
		return
	}
	if format != "512" && format != "1024" {
		httputil.Error(c, http.StatusBadRequest, httputil.ErrCodeInvalidFormat, "Format must be 512 or 1024")
		return
	}

	content, err := gc.GeographyUsecase.GetFlagFilePNGNative(countryCode, format)
	if err != nil {
		gc.logger.Debug("Flag PNG not found", zap.String("country", countryCode), zap.String("format", format), zap.Error(err))
		httputil.Error(c, http.StatusNotFound, httputil.ErrCodeNotFound, "Flag not found")
		return
	}

	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "public, max-age=604800")
	c.Data(http.StatusOK, "image/png", content)
}

func (gc *GeographyHandler) GetFlag(c *gin.Context) {
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

func (gc *GeographyHandler) GetFlagURL(c *gin.Context) {
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

	httputil.Success(c, http.StatusOK, map[string]string{
		"country_code": countryCode,
		"url":          url,
	})
}

func (gc *GeographyHandler) GetCountries(c *gin.Context) {
	ds, err := gc.GeographyUsecase.GetDefaultDataset()
	if err != nil || ds == nil {
		httputil.Error(c, http.StatusServiceUnavailable, httputil.ErrCodeNotFound, "No geography dataset available")
		return
	}

	limit := 50
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "50")); err == nil && l > 0 && l <= 300 {
		limit = l
	}
	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	continent := c.Query("continent")
	region := c.Query("region")

	if continent != "" || region != "" {
		filters := model.CountryFilters{Continent: continent, Region: region}
		countries, total, err := gc.GeographyUsecase.ListCountriesWithFilters(ds.ID, filters, limit, offset)
		if err != nil {
			httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch countries")
			return
		}
		httputil.Success(c, http.StatusOK, gin.H{
			keyData:   countries,
			"total":    total,
			"limit":    limit,
			"offset":   offset,
			"has_more": offset+len(countries) < int(total),
		})
		return
	}

	countries, total, err := gc.GeographyUsecase.ListCountries(ds.ID, limit, offset)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch countries")
		return
	}
	httputil.Success(c, http.StatusOK, gin.H{
		keyData:   countries,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
		"has_more": offset+len(countries) < int(total),
	})
}

func (gc *GeographyHandler) GetContinents(c *gin.Context) {
	ds, err := gc.GeographyUsecase.GetDefaultDataset()
	if err != nil || ds == nil {
		httputil.Error(c, http.StatusServiceUnavailable, httputil.ErrCodeNotFound, "No geography dataset available")
		return
	}
	continents, err := gc.GeographyUsecase.ListContinents(ds.ID)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, httputil.ErrCodeInternal, "Failed to fetch continents")
		return
	}
	httputil.Success(c, http.StatusOK, gin.H{keyData: continents})
}
