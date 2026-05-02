// backend/internal/usecase/admin/geography.go

package admin

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/repository"
	admin "github.com/Culturae-org/culturae/internal/repository/admin"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

type AdminGeographyUsecase struct {
	repo        admin.AdminGeographyRepositoryInterface
	publicRepo  repository.GeographyRepositoryInterface
	logger      *zap.Logger
	flagStorage repository.FlagStorageRepositoryInterface
	importsRepo admin.ImportJobRepositoryInterface
	datasetsUC  *AdminDatasetsUsecase
}

func NewAdminGeographyUsecase(
	repo admin.AdminGeographyRepositoryInterface,
	publicRepo repository.GeographyRepositoryInterface,
	logger *zap.Logger,
	flagStorage repository.FlagStorageRepositoryInterface,
	importsRepo admin.ImportJobRepositoryInterface,
) *AdminGeographyUsecase {
	return &AdminGeographyUsecase{
		repo:        repo,
		publicRepo:  publicRepo,
		logger:      logger,
		flagStorage: flagStorage,
		importsRepo: importsRepo,
		datasetsUC:  nil,
	}
}

// -----------------------------------------------
// Admin Geography Usecase Methods
//
// - ListGeographyDatasets
// - GetGeographyDataset
// - GetGeographyDatasetBySlug
// - DeleteGeographyDataset
// - SetDefaultGeographyDataset
// - GetDefaultGeographyDataset
// - UpdateDatasetStatistics
// - GetGeographyDatasetStatistics
// - GetGeographyDatasetByManifestURL
// - ImportGeographyFromManifest
// - UpdateCountry
// - UpdateContinent
// - UpdateRegion
//
// -----------------------------------------------

func (u *AdminGeographyUsecase) SetAdminDatasetsUsecase(datasetsUC *AdminDatasetsUsecase) {
	u.datasetsUC = datasetsUC
}

func (u *AdminGeographyUsecase) GetAdminDatasetsUsecase() *AdminDatasetsUsecase {
	return u.datasetsUC
}

func (u *AdminGeographyUsecase) ListGeographyDatasets(activeOnly bool) ([]model.GeographyDataset, error) {
	return u.repo.ListDatasets(activeOnly)
}

func (u *AdminGeographyUsecase) GetGeographyDataset(id uuid.UUID) (*model.GeographyDataset, error) {
	return u.repo.GetDatasetByID(id)
}

func (u *AdminGeographyUsecase) GetGeographyDatasetBySlug(slug string) (*model.GeographyDataset, error) {
	return u.repo.GetDatasetBySlug(slug)
}

func (u *AdminGeographyUsecase) DeleteGeographyDataset(id uuid.UUID, force bool) error {
	dataset, err := u.repo.GetDatasetByID(id)
	if err != nil {
		return err
	}

	datasets, err := u.repo.ListDatasets(true)
	if err != nil {
		return err
	}

	count := int64(0)
	for _, d := range datasets {
		if d.ID != id {
			count++
		}
	}

	if count == 0 {
		return fmt.Errorf("cannot delete the only dataset of this type")
	}

	if dataset.IsDefault {
		return fmt.Errorf("cannot delete the default dataset")
	}

	return u.repo.DeleteDataset(id)
}

func (u *AdminGeographyUsecase) SetDefaultGeographyDataset(id uuid.UUID) error {
	return u.repo.SetDefaultDataset(id)
}

func (u *AdminGeographyUsecase) GetDefaultGeographyDataset() (*model.GeographyDataset, error) {
	return u.repo.GetDefaultDataset()
}

func (u *AdminGeographyUsecase) UpdateDatasetStatistics(datasetID uuid.UUID) error {
	countryCount, err := u.repo.CountCountries(datasetID)
	if err != nil {
		return err
	}

	continentCount, err := u.repo.CountContinents(datasetID)
	if err != nil {
		return err
	}

	regionCount, err := u.repo.CountRegions(datasetID)
	if err != nil {
		return err
	}

	return u.repo.UpdateDatasetByID(datasetID, map[string]interface{}{
		keyCountryCount:   countryCount,
		keyContinentCount: continentCount,
		keyRegionCount:    regionCount,
	})
}

func (u *AdminGeographyUsecase) GetGeographyDatasetStatistics(datasetID uuid.UUID) (map[string]interface{}, error) {
	dataset, err := u.repo.GetDatasetByID(datasetID)
	if err != nil {
		return nil, err
	}

	countryCount, _ := u.repo.CountCountries(datasetID)
	continentCount, _ := u.repo.CountContinents(datasetID)
	regionCount, _ := u.repo.CountRegions(datasetID)

	return map[string]interface{}{
		"dataset_id":         datasetID,
		keyName:               dataset.Name,
		keyVersion:            dataset.Version,
		keyCountryCount:      countryCount,
		keyContinentCount:    continentCount,
		keyRegionCount:       regionCount,
		keyFlagCount:         dataset.FlagCount,
		"flag_png512_count":  dataset.FlagPNG512Count,
		"flag_png1024_count": dataset.FlagPNG1024Count,
		keyIsActive:          dataset.IsActive,
		keyIsDefault:         dataset.IsDefault,
		keyImportedAt:        dataset.ImportedAt,
	}, nil
}

func (u *AdminGeographyUsecase) GetGeographyDatasetByManifestURL(manifestURL string) (*model.GeographyDataset, error) {
	return u.repo.GetDatasetByManifestURL(manifestURL)
}

func (u *AdminGeographyUsecase) ImportGeographyFromManifest(manifestURL string) (*model.GeographyImportResult, error) {
	startTime := time.Now()
	result := &model.GeographyImportResult{
		Success: false,
		Errors:  []string{},
	}

	u.logger.Info("Starting Geography import from manifest",
		zap.String("manifest_url", manifestURL),
		zap.Time("started_at", startTime),
	)

	manifestResp := httputil.FetchURL(manifestURL)
	if manifestResp.Error != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", manifestResp.Error)
	}

	var manifest model.DatasetManifest
	if err := json.Unmarshal([]byte(manifestResp.Body), &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	u.logger.Info("Manifest loaded successfully",
		zap.String("dataset", manifest.Dataset),
		zap.String(keyVersion, manifest.Version),
		zap.String("type", manifest.Type),
		zap.Int("expected_countries", manifest.Counts.Countries),
		zap.Int("expected_continents", manifest.Counts.Continents),
		zap.Int("expected_regions", manifest.Counts.Regions),
	)

	datasetSlug := fmt.Sprintf("%s-v%s", manifest.Dataset, manifest.Version)
	datasetName := fmt.Sprintf("%s v%s", manifest.Dataset, manifest.Version)

	_, err := u.repo.GetDatasetBySlug(datasetSlug)
	if err == nil {
		return nil, fmt.Errorf("dataset with slug '%s' already exists", datasetSlug)
	}

	job := model.ImportJob{
		ID:          uuid.New(),
		ManifestURL: manifestURL,
		Dataset:     manifest.Dataset,
		Version:     manifest.Version,
		StartedAt:   startTime,
	}

	if err := u.importsRepo.CreateImportJob(&job); err != nil {
		return nil, fmt.Errorf("failed to create import job: %w", err)
	}

	existingCount, err := u.repo.CountAllDatasets()
	if err != nil {
		return nil, fmt.Errorf("failed to count existing datasets: %w", err)
	}

	isDefault := existingCount == 0

	manifestJSON, _ := json.Marshal(manifest)
	dataset := model.GeographyDataset{
		ID:           uuid.New(),
		Slug:         datasetSlug,
		Name:         datasetName,
		Description:  fmt.Sprintf("Imported from manifest on %s", startTime.Format("2006-01-02")),
		Version:      manifest.Version,
		ManifestURL:  manifestURL,
		ManifestData: datatypes.JSON(manifestJSON),
		Source:       "manifest",
		ImportJobID:  &job.ID,
		ImportedAt:   startTime,
		IsActive:     true,
		IsDefault:    isDefault,
	}

	if err := u.repo.CreateDataset(&dataset); err != nil {
		return nil, fmt.Errorf("failed to create dataset: %w", err)
	}

	u.logger.Info("Dataset created",
		zap.String("slug", dataset.Slug),
		zap.String("id", dataset.ID.String()),
	)

	baseURL := manifestURL[:strings.LastIndex(manifestURL, "/")+1]

	if contains(manifest.Includes, "continents") {
		continentsURL := baseURL + "continents.ndjson"
		expectedChecksum := manifest.Checksums["continents.ndjson"]
		added, err := u.repo.ImportContinents(continentsURL, dataset.ID, job.ID, expectedChecksum, u.importsRepo, u.logger)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("continents: %v", err))
			u.logger.Error("Failed to import continents", zap.Error(err))
		} else {
			result.ContinentsAdded = added
			u.logger.Info("Continents imported", zap.Int("added", added))
			if manifest.Counts.Continents > 0 && added != manifest.Counts.Continents {
				u.logger.Warn("Continents count mismatch", zap.Int("expected", manifest.Counts.Continents), zap.Int("actual", added))
			}
		}
	}

	if contains(manifest.Includes, "regions") {
		regionsURL := baseURL + "regions.ndjson"
		expectedChecksum := manifest.Checksums["regions.ndjson"]
		added, err := u.repo.ImportRegions(regionsURL, dataset.ID, job.ID, expectedChecksum, u.importsRepo, u.logger)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("regions: %v", err))
			u.logger.Error("Failed to import regions", zap.Error(err))
		} else {
			result.RegionsAdded = added
			u.logger.Info("Regions imported", zap.Int("added", added))
			if manifest.Counts.Regions > 0 && added != manifest.Counts.Regions {
				u.logger.Warn("Regions count mismatch", zap.Int("expected", manifest.Counts.Regions), zap.Int("actual", added))
			}
		}
	}

	if contains(manifest.Includes, "countries") {
		countriesURL := baseURL + "countries.ndjson"
		expectedChecksum := manifest.Checksums["countries.ndjson"]
		added, err := u.repo.ImportCountries(countriesURL, dataset.ID, job.ID, expectedChecksum, u.importsRepo, u.logger)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("countries: %v", err))
			u.logger.Error("Failed to import countries", zap.Error(err))
		} else {
			result.CountriesAdded = added
			u.logger.Info("Countries imported", zap.Int("added", added))
			if manifest.Counts.Countries > 0 && added != manifest.Counts.Countries {
				u.logger.Warn("Countries count mismatch", zap.Int("expected", manifest.Counts.Countries), zap.Int("actual", added))
			}
		}
	}

	flagsBaseURL := ""
	flagsPNG512BaseURL := ""
	flagsPNG1024BaseURL := ""
	if manifest.Assets != nil {
		if path, ok := manifest.Assets["flags-svg"]; ok && path != "" {
			flagsBaseURL = baseURL + path
		}
		if path, ok := manifest.Assets["flags-png-512"]; ok && path != "" {
			flagsPNG512BaseURL = baseURL + path
		}
		if path, ok := manifest.Assets["flags-png-1024"]; ok && path != "" {
			flagsPNG1024BaseURL = baseURL + path
		}
	}

	u.logger.Info("Flags base URLs configured",
		zap.String("flags_svg_url", flagsBaseURL),
		zap.String("flags_png512_url", flagsPNG512BaseURL),
		zap.String("flags_png1024_url", flagsPNG1024BaseURL),
		zap.Any("manifest_assets", manifest.Assets),
	)

	dataset.CountryCount = result.CountriesAdded
	dataset.ContinentCount = result.ContinentsAdded
	dataset.RegionCount = result.RegionsAdded
	dataset.FlagCount = manifest.Counts.Flags

	if err := u.repo.UpdateDatasetByID(dataset.ID, map[string]interface{}{
		keyCountryCount:   dataset.CountryCount,
		keyContinentCount: dataset.ContinentCount,
		keyRegionCount:    dataset.RegionCount,
		keyFlagCount:      dataset.FlagCount,
	}); err != nil {
		return nil, fmt.Errorf("failed to update dataset statistics: %w", err)
	}

	duration := time.Since(startTime)
	result.Success = len(result.Errors) == 0

	if result.Success {
		result.Message = fmt.Sprintf("Successfully imported %d countries, %d continents, %d regions in %v",
			result.CountriesAdded, result.ContinentsAdded, result.RegionsAdded, duration.Round(time.Millisecond))
	} else {
		result.Message = fmt.Sprintf("Import completed with %d errors", len(result.Errors))
	}

	finished := time.Now()
	jobUpdates := map[string]interface{}{
		"finished_at": finished,
		"success":     result.Success,
		"added":       result.CountriesAdded + result.ContinentsAdded + result.RegionsAdded,
		"updated":     result.CountriesUpdated + result.ContinentsUpdated + result.RegionsUpdated,
		"skipped":     result.CountriesSkipped,
		"errors":      len(result.Errors),
		keyMessage:     result.Message,
	}

	if err := u.importsRepo.UpdateImportJobStatus(job.ID, jobUpdates); err != nil {
		u.logger.Error("Failed to update import job", zap.Error(err))
	}

	u.logger.Info("Geography import completed",
		zap.Bool("success", result.Success),
		zap.Int("countries_added", result.CountriesAdded),
		zap.Int("continents_added", result.ContinentsAdded),
		zap.Int("regions_added", result.RegionsAdded),
		zap.Duration("duration", duration),
	)

	if u.flagStorage == nil {
		u.logger.Warn("MinIO service not configured, skipping flag import")
	} else if flagsBaseURL == "" && flagsPNG512BaseURL == "" && flagsPNG1024BaseURL == "" {
		u.logger.Warn("No flags base URLs configured, skipping flag import")
	} else {
		u.logger.Info("Starting async flag import",
			zap.String("flags_svg_url", flagsBaseURL),
			zap.String("flags_png512_url", flagsPNG512BaseURL),
			zap.String("flags_png1024_url", flagsPNG1024BaseURL),
			zap.String("dataset_id", dataset.ID.String()),
		)
		go u.importFlagsAsync(flagsBaseURL, flagsPNG512BaseURL, flagsPNG1024BaseURL, dataset.ID, job.ID)
	}

	return result, nil
}

func (u *AdminGeographyUsecase) importFlagsAsync(svgBaseURL, png512BaseURL, png1024BaseURL string, datasetID uuid.UUID, jobID uuid.UUID) {
	u.logger.Info("Starting async flag import",
		zap.String("svg_base_url", svgBaseURL),
		zap.String("png512_base_url", png512BaseURL),
		zap.String("png1024_base_url", png1024BaseURL),
		zap.String("dataset_id", datasetID.String()),
		zap.String("job_id", jobID.String()),
	)

	startTime := time.Now()

	now := time.Now()
	if err := u.importsRepo.UpdateImportJobStatus(jobID, map[string]interface{}{
		"flags_started_at": now,
	}); err != nil {
		u.logger.Error("Failed to update flags_started_at", zap.Error(err))
	}

	countries, _, err := u.publicRepo.ListCountries(datasetID, 1000, 0)
	if err != nil {
		u.logger.Error("Failed to get countries for flag import", zap.Error(err))
		return
	}

	var svgCount, png512Count, png1024Count int
	var mu sync.Mutex
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10)

	for _, country := range countries {
		if country.Slug == "" {
			continue
		}

		wg.Add(1)
		go func(countryCode string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if svgBaseURL != "" {
				flagURL := fmt.Sprintf("%s%s.svg", svgBaseURL, countryCode)
				if _, err := u.flagStorage.UploadFlagFromURL(countryCode, flagURL); err != nil {
					u.logger.Debug("Failed to import SVG flag",
						zap.String("country", countryCode),
						zap.Error(err),
					)
				} else {
					mu.Lock()
					svgCount++
					mu.Unlock()
				}
			}

			if png512BaseURL != "" {
				flagURL := fmt.Sprintf("%s%s.png", png512BaseURL, countryCode)
				if _, err := u.flagStorage.UploadFlagPNGFromURL(countryCode, "512", flagURL); err != nil {
					u.logger.Debug("Failed to import PNG 512 flag",
						zap.String("country", countryCode),
						zap.Error(err),
					)
				} else {
					mu.Lock()
					png512Count++
					mu.Unlock()
				}
			}

			if png1024BaseURL != "" {
				flagURL := fmt.Sprintf("%s%s.png", png1024BaseURL, countryCode)
				if _, err := u.flagStorage.UploadFlagPNGFromURL(countryCode, "1024", flagURL); err != nil {
					u.logger.Debug("Failed to import PNG 1024 flag",
						zap.String("country", countryCode),
						zap.Error(err),
					)
				} else {
					mu.Lock()
					png1024Count++
					mu.Unlock()
				}
			}
		}(country.Slug)
	}

	wg.Wait()

	duration := time.Since(startTime)
	u.logger.Info("Flag import completed",
		zap.Int("svg_success", svgCount),
		zap.Int("png512_success", png512Count),
		zap.Int("png1024_success", png1024Count),
		zap.Duration("duration", duration),
	)

	updates := map[string]interface{}{}
	if svgBaseURL != "" {
		updates[keyFlagCount] = svgCount
	}
	if png512BaseURL != "" {
		updates["flag_png512_count"] = png512Count
	}
	if png1024BaseURL != "" {
		updates["flag_png1024_count"] = png1024Count
	}

	if len(updates) > 0 {
		if err := u.repo.UpdateDatasetByID(datasetID, updates); err != nil {
			u.logger.Error("Failed to update flag counts", zap.Error(err))
		}
	}

	finished := time.Now()
	flagUpdates := map[string]interface{}{
		"flags_finished_at":   finished,
		"flags_svg_count":     svgCount,
		"flags_png512_count":  png512Count,
		"flags_png1024_count": png1024Count,
	}

	if err := u.importsRepo.UpdateImportJobStatus(jobID, flagUpdates); err != nil {
		u.logger.Error("Failed to update import job with flag counts", zap.Error(err))
	}
}

func (u *AdminGeographyUsecase) UpdateCountry(datasetID uuid.UUID, slug string, updates map[string]interface{}) (*model.Country, error) {
	country, err := u.repo.GetCountryBySlug(slug, datasetID)
	if err != nil {
		return nil, fmt.Errorf("country not found: %w", err)
	}

	if name, ok := updates[keyName].(map[string]interface{}); ok {
		jsonName, err := json.Marshal(name)
		if err == nil {
			country.Name = datatypes.JSON(jsonName)
		}
	}
	if capital, ok := updates["capital"].(map[string]interface{}); ok {
		jsonCapital, err := json.Marshal(capital)
		if err == nil {
			country.Capital = datatypes.JSON(jsonCapital)
		}
	}
	if continent, ok := updates["continent"].(string); ok {
		country.Continent = continent
	}
	if region, ok := updates["region"].(string); ok {
		country.Region = region
	}
	if latitude, ok := updates["latitude"].(float64); ok {
		country.Latitude = latitude
	}
	if longitude, ok := updates["longitude"].(float64); ok {
		country.Longitude = longitude
	}
	if population, ok := updates["population"].(float64); ok {
		country.Population = int64(population)
	}
	if areaKm2, ok := updates["area_km2"].(float64); ok {
		country.AreaKm2 = int64(areaKm2)
	}
	if currency, ok := updates["currency"].(map[string]interface{}); ok {
		jsonCurrency, err := json.Marshal(currency)
		if err == nil {
			country.Currency = datatypes.JSON(jsonCurrency)
		}
	}
	if phoneCode, ok := updates["phone_code"].(string); ok {
		country.PhoneCode = phoneCode
	}
	if tld, ok := updates["tld"].(string); ok {
		country.TLD = tld
	}
	if drivingSide, ok := updates["driving_side"].(string); ok {
		country.DrivingSide = drivingSide
	}
	if flag, ok := updates["flag"].(string); ok {
		country.Flag = flag
	}

	if err := u.repo.UpdateCountry(country); err != nil {
		return nil, fmt.Errorf("failed to update country: %w", err)
	}

	return country, nil
}

func (u *AdminGeographyUsecase) UpdateContinent(datasetID uuid.UUID, slug string, updates map[string]interface{}) (*model.Continent, error) {
	continent, err := u.repo.GetContinentBySlug(slug, datasetID)
	if err != nil {
		return nil, fmt.Errorf("continent not found: %w", err)
	}

	if name, ok := updates[keyName].(map[string]interface{}); ok {
		jsonName, err := json.Marshal(name)
		if err == nil {
			continent.Name = datatypes.JSON(jsonName)
		}
	}
	if population, ok := updates["population"].(float64); ok {
		continent.Population = int64(population)
	}
	if areaKm2, ok := updates["area_km2"].(float64); ok {
		continent.AreaKm2 = int64(areaKm2)
	}

	if err := u.repo.UpdateContinent(continent); err != nil {
		return nil, fmt.Errorf("failed to update continent: %w", err)
	}

	return continent, nil
}

func (u *AdminGeographyUsecase) UpdateRegion(datasetID uuid.UUID, slug string, updates map[string]interface{}) (*model.Region, error) {
	region, err := u.repo.GetRegionBySlug(slug, datasetID)
	if err != nil {
		return nil, fmt.Errorf("region not found: %w", err)
	}

	if name, ok := updates[keyName].(map[string]interface{}); ok {
		jsonName, err := json.Marshal(name)
		if err == nil {
			region.Name = datatypes.JSON(jsonName)
		}
	}
	if continent, ok := updates["continent"].(string); ok {
		region.Continent = continent
	}

	if err := u.repo.UpdateRegion(region); err != nil {
		return nil, fmt.Errorf("failed to update region: %w", err)
	}

	return region, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
