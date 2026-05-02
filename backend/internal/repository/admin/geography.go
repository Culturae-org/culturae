// backend/internal/repository/admin/geography.go

package admin

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AdminGeographyRepositoryInterface interface {
	CreateDataset(dataset *model.GeographyDataset) error
	GetDatasetByID(id uuid.UUID) (*model.GeographyDataset, error)
	GetDatasetBySlug(slug string) (*model.GeographyDataset, error)
	GetDatasetByManifestURL(manifestURL string) (*model.GeographyDataset, error)
	ListDatasets(activeOnly bool) ([]model.GeographyDataset, error)
	UpdateDataset(dataset *model.GeographyDataset) error
	UpdateDatasetByID(id uuid.UUID, updates map[string]interface{}) error
	DeleteDataset(id uuid.UUID) error
	GetDefaultDataset() (*model.GeographyDataset, error)
	SetDefaultDataset(id uuid.UUID) error
	CountCountries(datasetID uuid.UUID) (int64, error)
	CountContinents(datasetID uuid.UUID) (int64, error)
	CountRegions(datasetID uuid.UUID) (int64, error)
	GetCountryBySlug(slug string, datasetID uuid.UUID) (*model.Country, error)
	ListCountries(datasetID uuid.UUID, limit, offset int) ([]model.Country, int64, error)
	GetContinentBySlug(slug string, datasetID uuid.UUID) (*model.Continent, error)
	GetRegionBySlug(slug string, datasetID uuid.UUID) (*model.Region, error)
	CreateCountry(country *model.Country) error
	UpdateCountry(country *model.Country) error
	DeleteCountriesByDataset(datasetID uuid.UUID) error
	CreateContinent(continent *model.Continent) error
	UpdateContinent(continent *model.Continent) error
	DeleteContinentsByDataset(datasetID uuid.UUID) error
	CreateRegion(region *model.Region) error
	UpdateRegion(region *model.Region) error
	DeleteRegionsByDataset(datasetID uuid.UUID) error
	ImportContinents(url string, datasetID uuid.UUID, jobID uuid.UUID, expectedChecksum string, importsRepo interface {
		SaveImportLog(*model.ImportQuestionLog) error
	}, logger *zap.Logger) (int, error)
	ImportRegions(url string, datasetID uuid.UUID, jobID uuid.UUID, expectedChecksum string, importsRepo interface {
		SaveImportLog(*model.ImportQuestionLog) error
	}, logger *zap.Logger) (int, error)
	ImportCountries(url string, datasetID uuid.UUID, jobID uuid.UUID, expectedChecksum string, importsRepo interface {
		SaveImportLog(*model.ImportQuestionLog) error
	}, logger *zap.Logger) (int, error)
	CountAllDatasets() (int64, error)
}

type AdminGeographyRepository struct {
	DB *gorm.DB
}

func NewAdminGeographyRepository(db *gorm.DB) *AdminGeographyRepository {
	return &AdminGeographyRepository{
		DB: db,
	}
}

func (r *AdminGeographyRepository) CreateDataset(dataset *model.GeographyDataset) error {
	return r.DB.Create(dataset).Error
}

func (r *AdminGeographyRepository) GetDatasetByID(id uuid.UUID) (*model.GeographyDataset, error) {
	var dataset model.GeographyDataset
	err := r.DB.Preload("ImportJob").First(&dataset, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrDatasetNotFound
		}
		return nil, err
	}
	return &dataset, nil
}

func (r *AdminGeographyRepository) GetDatasetBySlug(slug string) (*model.GeographyDataset, error) {
	var dataset model.GeographyDataset
	err := r.DB.Preload("ImportJob").Where("slug = ?", slug).First(&dataset).Error
	if err != nil {
		return nil, err
	}
	return &dataset, nil
}

func (r *AdminGeographyRepository) GetDatasetByManifestURL(manifestURL string) (*model.GeographyDataset, error) {
	var dataset model.GeographyDataset
	err := r.DB.Where("manifest_url = ?", manifestURL).First(&dataset).Error
	if err != nil {
		return nil, err
	}
	return &dataset, nil
}

func (r *AdminGeographyRepository) ListDatasets(activeOnly bool) ([]model.GeographyDataset, error) {
	var datasets []model.GeographyDataset
	query := r.DB.Preload("ImportJob")

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("created_at DESC").Find(&datasets).Error
	return datasets, err
}

func (r *AdminGeographyRepository) UpdateDataset(dataset *model.GeographyDataset) error {
	return r.DB.Save(dataset).Error
}

func (r *AdminGeographyRepository) UpdateDatasetByID(id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	return r.DB.Model(&model.GeographyDataset{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AdminGeographyRepository) DeleteDataset(id uuid.UUID) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("dataset_id = ?", id).Delete(&model.Country{}).Error; err != nil {
			return err
		}
		if err := tx.Where("dataset_id = ?", id).Delete(&model.Continent{}).Error; err != nil {
			return err
		}
		if err := tx.Where("dataset_id = ?", id).Delete(&model.Region{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.GeographyDataset{}, "id = ?", id).Error
	})
}

func (r *AdminGeographyRepository) GetDefaultDataset() (*model.GeographyDataset, error) {
	var dataset model.GeographyDataset
	err := r.DB.Preload("ImportJob").Where("is_default = ? AND is_active = ?", true, true).First(&dataset).Error
	if err != nil {
		return nil, err
	}
	return &dataset, nil
}

func (r *AdminGeographyRepository) SetDefaultDataset(id uuid.UUID) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.GeographyDataset{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(&model.GeographyDataset{}).Where("id = ?", id).Update("is_default", true).Error
	})
}

func (r *AdminGeographyRepository) CreateCountry(country *model.Country) error {
	return r.DB.Create(country).Error
}

func (r *AdminGeographyRepository) UpdateCountry(country *model.Country) error {
	return r.DB.Save(country).Error
}

func (r *AdminGeographyRepository) DeleteCountriesByDataset(datasetID uuid.UUID) error {
	return r.DB.Where("dataset_id = ?", datasetID).Delete(&model.Country{}).Error
}

func (r *AdminGeographyRepository) CreateContinent(continent *model.Continent) error {
	return r.DB.Create(continent).Error
}

func (r *AdminGeographyRepository) UpdateContinent(continent *model.Continent) error {
	return r.DB.Save(continent).Error
}

func (r *AdminGeographyRepository) DeleteContinentsByDataset(datasetID uuid.UUID) error {
	return r.DB.Where("dataset_id = ?", datasetID).Delete(&model.Continent{}).Error
}

func (r *AdminGeographyRepository) CreateRegion(region *model.Region) error {
	return r.DB.Create(region).Error
}

func (r *AdminGeographyRepository) UpdateRegion(region *model.Region) error {
	return r.DB.Save(region).Error
}

func (r *AdminGeographyRepository) DeleteRegionsByDataset(datasetID uuid.UUID) error {
	return r.DB.Where("dataset_id = ?", datasetID).Delete(&model.Region{}).Error
}

func (r *AdminGeographyRepository) GetDB() *gorm.DB {
	return r.DB
}

func (r *AdminGeographyRepository) CountCountries(datasetID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB.Model(&model.Country{}).Where("dataset_id = ?", datasetID).Count(&count).Error
	return count, err
}

func (r *AdminGeographyRepository) CountContinents(datasetID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB.Model(&model.Continent{}).Where("dataset_id = ?", datasetID).Count(&count).Error
	return count, err
}

func (r *AdminGeographyRepository) CountRegions(datasetID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB.Model(&model.Region{}).Where("dataset_id = ?", datasetID).Count(&count).Error
	return count, err
}

func (r *AdminGeographyRepository) GetCountryBySlug(slug string, datasetID uuid.UUID) (*model.Country, error) {
	var country model.Country
	if err := r.DB.Where("slug = ? AND dataset_id = ?", slug, datasetID).First(&country).Error; err != nil {
		return nil, err
	}
	return &country, nil
}

func (r *AdminGeographyRepository) ListCountries(datasetID uuid.UUID, limit, offset int) ([]model.Country, int64, error) {
	var countries []model.Country
	var total int64
	query := r.DB.Where("dataset_id = ?", datasetID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Limit(limit).Offset(offset).Find(&countries).Error; err != nil {
		return nil, 0, err
	}
	return countries, total, nil
}

func (r *AdminGeographyRepository) GetContinentBySlug(slug string, datasetID uuid.UUID) (*model.Continent, error) {
	var continent model.Continent
	if err := r.DB.Where("slug = ? AND dataset_id = ?", slug, datasetID).First(&continent).Error; err != nil {
		return nil, err
	}
	return &continent, nil
}

func (r *AdminGeographyRepository) GetRegionBySlug(slug string, datasetID uuid.UUID) (*model.Region, error) {
	var region model.Region
	if err := r.DB.Where("slug = ? AND dataset_id = ?", slug, datasetID).First(&region).Error; err != nil {
		return nil, err
	}
	return &region, nil
}

func (r *AdminGeographyRepository) ImportContinents(url string, datasetID uuid.UUID, jobID uuid.UUID, expectedChecksum string, importsRepo interface {
	SaveImportLog(*model.ImportQuestionLog) error
}, logger *zap.Logger) (int, error) {
	tx := r.DB.Begin()
	if tx.Error != nil {
		return 0, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	resp := httputil.FetchURLWithChecksum(url, expectedChecksum)
	if resp.Error != nil {
		tx.Rollback()
		if cm, ok := httputil.IsChecksumMismatch(resp.Error); ok {
			logger.Error("Continents file checksum verification failed",
				zap.String("url", url),
				zap.String("expected", cm.Expected),
				zap.String("actual", cm.Actual))
			return 0, fmt.Errorf("checksum verification failed for continents file: the file may have been modified or corrupted (expected: %s, got: %s)", cm.Expected, cm.Actual)
		}
		return 0, fmt.Errorf("failed to fetch continents: %w", resp.Error)
	}

	var continents []model.Continent
	scanner := bufio.NewScanner(strings.NewReader(resp.Body))
	lineNum := 0
	added := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		if line == "" {
			continue
		}

		var raw model.ContinentRaw
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			logger.Warn("Failed to parse continent", zap.Int("line", lineNum), zap.Error(err))
			_ = importsRepo.SaveImportLog(&model.ImportQuestionLog{JobID: jobID, Line: lineNum, Slug: "", Action: actionError, Message: err.Error()})
			continue
		}

		nameJSON, _ := json.Marshal(raw.Name)
		countriesJSON, _ := json.Marshal(raw.Countries)

		continent := model.Continent{
			ID:         uuid.New(),
			DatasetID:  datasetID,
			Slug:       raw.Slug,
			Name:       datatypes.JSON(nameJSON),
			Countries:  datatypes.JSON(countriesJSON),
			AreaKm2:    raw.AreaKm2,
			Population: raw.Population,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		continents = append(continents, continent)
		_ = importsRepo.SaveImportLog(&model.ImportQuestionLog{JobID: jobID, Line: lineNum, Slug: raw.Slug, Action: actionCreated, Message: "new continent"})
		added++
	}

	if err := scanner.Err(); err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("scanner error: %w", err)
	}

	if len(continents) > 0 {
		if err := tx.CreateInBatches(continents, 100).Error; err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to insert continents: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return added, nil
}

func (r *AdminGeographyRepository) ImportRegions(url string, datasetID uuid.UUID, jobID uuid.UUID, expectedChecksum string, importsRepo interface {
	SaveImportLog(*model.ImportQuestionLog) error
}, logger *zap.Logger) (int, error) {
	tx := r.DB.Begin()
	if tx.Error != nil {
		return 0, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	resp := httputil.FetchURLWithChecksum(url, expectedChecksum)
	if resp.Error != nil {
		tx.Rollback()
		if cm, ok := httputil.IsChecksumMismatch(resp.Error); ok {
			logger.Error("Regions file checksum verification failed",
				zap.String("url", url),
				zap.String("expected", cm.Expected),
				zap.String("actual", cm.Actual))
			return 0, fmt.Errorf("checksum verification failed for regions file: the file may have been modified or corrupted (expected: %s, got: %s)", cm.Expected, cm.Actual)
		}
		return 0, fmt.Errorf("failed to fetch regions: %w", resp.Error)
	}

	var regions []model.Region
	scanner := bufio.NewScanner(strings.NewReader(resp.Body))
	lineNum := 0
	added := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		if line == "" {
			continue
		}

		var raw model.RegionRaw
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			logger.Warn("Failed to parse region", zap.Int("line", lineNum), zap.Error(err))
			_ = importsRepo.SaveImportLog(&model.ImportQuestionLog{JobID: jobID, Line: lineNum, Slug: "", Action: actionError, Message: err.Error()})
			continue
		}

		nameJSON, _ := json.Marshal(raw.Name)
		countriesJSON, _ := json.Marshal(raw.Countries)

		region := model.Region{
			ID:        uuid.New(),
			DatasetID: datasetID,
			Slug:      raw.Slug,
			Name:      datatypes.JSON(nameJSON),
			Continent: raw.Continent,
			Countries: datatypes.JSON(countriesJSON),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		regions = append(regions, region)
		_ = importsRepo.SaveImportLog(&model.ImportQuestionLog{JobID: jobID, Line: lineNum, Slug: raw.Slug, Action: actionCreated, Message: "new region"})
		added++
	}

	if err := scanner.Err(); err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("scanner error: %w", err)
	}

	if len(regions) > 0 {
		if err := tx.CreateInBatches(regions, 100).Error; err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to insert regions: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return added, nil
}

func (r *AdminGeographyRepository) ImportCountries(url string, datasetID uuid.UUID, jobID uuid.UUID, expectedChecksum string, importsRepo interface {
	SaveImportLog(*model.ImportQuestionLog) error
}, logger *zap.Logger) (int, error) {
	tx := r.DB.Begin()
	if tx.Error != nil {
		return 0, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	resp := httputil.FetchURLWithChecksum(url, expectedChecksum)
	if resp.Error != nil {
		tx.Rollback()
		if cm, ok := httputil.IsChecksumMismatch(resp.Error); ok {
			logger.Error("Countries file checksum verification failed",
				zap.String("url", url),
				zap.String("expected", cm.Expected),
				zap.String("actual", cm.Actual))
			return 0, fmt.Errorf("checksum verification failed for countries file: the file may have been modified or corrupted (expected: %s, got: %s)", cm.Expected, cm.Actual)
		}
		return 0, fmt.Errorf("failed to fetch countries: %w", resp.Error)
	}

	var countries []model.Country
	scanner := bufio.NewScanner(strings.NewReader(resp.Body))
	lineNum := 0
	added := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		if line == "" {
			continue
		}

		var raw model.CountryRaw
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			logger.Warn("Failed to parse country", zap.Int("line", lineNum), zap.Error(err))
			_ = importsRepo.SaveImportLog(&model.ImportQuestionLog{JobID: jobID, Line: lineNum, Slug: "", Action: actionError, Message: err.Error()})
			continue
		}

		nameJSON, _ := json.Marshal(raw.Name)
		officialJSON, _ := json.Marshal(raw.Official)
		capitalJSON, _ := json.Marshal(raw.Capital)
		currencyJSON, _ := json.Marshal(raw.Currency)
		languagesJSON, _ := json.Marshal(raw.Languages)
		neighborsJSON, _ := json.Marshal(raw.Neighbors)

		country := model.Country{
			ID:           uuid.New(),
			DatasetID:    datasetID,
			Slug:         raw.Slug,
			ISOAlpha2:    raw.ISOAlpha2,
			ISOAlpha3:    raw.ISOAlpha3,
			ISONumeric:   raw.ISONumeric,
			Name:         datatypes.JSON(nameJSON),
			OfficialName: datatypes.JSON(officialJSON),
			Capital:      datatypes.JSON(capitalJSON),
			Continent:    raw.Continent,
			Region:       raw.Region,
			Latitude:     raw.Coords.Lat,
			Longitude:    raw.Coords.Lng,
			Flag:         raw.Flag,
			Population:   raw.Population,
			AreaKm2:      int64(raw.AreaKm2),
			Currency:     datatypes.JSON(currencyJSON),
			Languages:    datatypes.JSON(languagesJSON),
			Neighbors:    datatypes.JSON(neighborsJSON),
			TLD:          raw.TLD,
			PhoneCode:    raw.PhoneCode,
			DrivingSide:  raw.DrivingSide,
			Independent:  raw.Independent,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		countries = append(countries, country)
		_ = importsRepo.SaveImportLog(&model.ImportQuestionLog{JobID: jobID, Line: lineNum, Slug: raw.Slug, Action: actionCreated, Message: "new country"})
		added++
	}

	if err := scanner.Err(); err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("scanner error: %w", err)
	}

	if len(countries) > 0 {
		if err := tx.CreateInBatches(countries, 100).Error; err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to insert countries: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return added, nil
}

func (r *AdminGeographyRepository) CountAllDatasets() (int64, error) {
	var count int64
	err := r.DB.Model(&model.GeographyDataset{}).Count(&count).Error
	return count, err
}
