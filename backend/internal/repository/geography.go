// backend/internal/repository/geography.go

package repository

import (
	"io"

	"github.com/Culturae-org/culturae/internal/infrastructure/storage"
	"github.com/Culturae-org/culturae/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GeographyRepositoryInterface interface {
	GetDefaultDataset() (*model.GeographyDataset, error)
	GetDatasetByID(id uuid.UUID) (*model.GeographyDataset, error)
	GetCountryByID(id uuid.UUID) (*model.Country, error)
	GetCountryBySlug(slug string, datasetID uuid.UUID) (*model.Country, error)
	ListCountries(datasetID uuid.UUID, limit, offset int) ([]model.Country, int64, error)
	ListCountriesWithFilters(datasetID uuid.UUID, filters model.CountryFilters, limit, offset int) ([]model.Country, int64, error)
	ListCountriesByContinent(datasetID uuid.UUID, continent string) ([]model.Country, error)
	ListCountriesByRegion(datasetID uuid.UUID, region string) ([]model.Country, error)
	SearchCountries(datasetID uuid.UUID, query string, limit, offset int) ([]model.Country, int64, error)
	CountCountries(datasetID uuid.UUID) (int64, error)
	GetContinentBySlug(slug string, datasetID uuid.UUID) (*model.Continent, error)
	ListContinents(datasetID uuid.UUID) ([]model.Continent, error)
	CountContinents(datasetID uuid.UUID) (int64, error)
	GetRegionBySlug(slug string, datasetID uuid.UUID) (*model.Region, error)
	ListRegions(datasetID uuid.UUID) ([]model.Region, error)
	ListRegionsByContinent(datasetID uuid.UUID, continent string) ([]model.Region, error)
	CountRegions(datasetID uuid.UUID) (int64, error)
	ListRandomCountries(datasetID uuid.UUID, count int) ([]model.Country, error)
	ListRandomCountriesFiltered(datasetID uuid.UUID, count int, independentOnly bool, continent string) ([]model.Country, error)
	ListAllIndependentCountries(datasetID uuid.UUID, continent string) ([]model.Country, error)
	GetDB() *gorm.DB
}

type GeographyRepository struct {
	DB *gorm.DB
}

func NewGeographyRepository(
	db *gorm.DB,
) *GeographyRepository {
	return &GeographyRepository{
		DB: db,
	}
}

func (r *GeographyRepository) GetDefaultDataset() (*model.GeographyDataset, error) {
	var dataset model.GeographyDataset
	err := r.DB.Preload("ImportJob").Where("is_default = ? AND is_active = ?", true, true).First(&dataset).Error
	if err != nil {
		return nil, err
	}
	return &dataset, nil
}

func (r *GeographyRepository) GetDatasetByID(id uuid.UUID) (*model.GeographyDataset, error) {
	var dataset model.GeographyDataset
	if err := r.DB.Preload("ImportJob").First(&dataset, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &dataset, nil
}

func (r *GeographyRepository) GetCountryByID(id uuid.UUID) (*model.Country, error) {
	var country model.Country
	err := r.DB.First(&country, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &country, nil
}

func (r *GeographyRepository) GetCountryBySlug(slug string, datasetID uuid.UUID) (*model.Country, error) {
	var country model.Country
	err := r.DB.Where("slug = ? AND dataset_id = ?", slug, datasetID).First(&country).Error
	if err != nil {
		return nil, err
	}
	return &country, nil
}

func (r *GeographyRepository) ListCountries(datasetID uuid.UUID, limit, offset int) ([]model.Country, int64, error) {
	var countries []model.Country
	var total int64

	query := r.DB.Model(&model.Country{}).Where("dataset_id = ?", datasetID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("slug ASC").Limit(limit).Offset(offset).Find(&countries).Error
	return countries, total, err
}

func (r *GeographyRepository) ListCountriesWithFilters(datasetID uuid.UUID, filters model.CountryFilters, limit, offset int) ([]model.Country, int64, error) {
	var countries []model.Country
	var total int64

	query := r.DB.Model(&model.Country{}).Where("dataset_id = ?", datasetID)

	if filters.Continent != "" {
		query = query.Where("continent = ?", filters.Continent)
	}
	if filters.Region != "" {
		query = query.Where("region = ?", filters.Region)
	}
	if filters.PopulationMin != nil {
		query = query.Where("population >= ?", *filters.PopulationMin)
	}
	if filters.PopulationMax != nil {
		query = query.Where("population <= ?", *filters.PopulationMax)
	}
	if filters.AreaMin != nil {
		query = query.Where("area_km2 >= ?", *filters.AreaMin)
	}
	if filters.AreaMax != nil {
		query = query.Where("area_km2 <= ?", *filters.AreaMax)
	}
	if filters.Independent != nil {
		query = query.Where("independent = ?", *filters.Independent)
	}
	if filters.DrivingSide != "" {
		query = query.Where("driving_side = ?", filters.DrivingSide)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("slug ASC").Limit(limit).Offset(offset).Find(&countries).Error
	return countries, total, err
}

func (r *GeographyRepository) ListCountriesByContinent(datasetID uuid.UUID, continent string) ([]model.Country, error) {
	var countries []model.Country
	err := r.DB.Where("dataset_id = ? AND continent = ?", datasetID, continent).Order("slug ASC").Find(&countries).Error
	return countries, err
}

func (r *GeographyRepository) ListCountriesByRegion(datasetID uuid.UUID, region string) ([]model.Country, error) {
	var countries []model.Country
	err := r.DB.Where("dataset_id = ? AND region = ?", datasetID, region).Order("slug ASC").Find(&countries).Error
	return countries, err
}

func (r *GeographyRepository) SearchCountries(datasetID uuid.UUID, query string, limit, offset int) ([]model.Country, int64, error) {
	var countries []model.Country
	var total int64

	searchQuery := "%" + query + "%"
	dbQuery := r.DB.Model(&model.Country{}).Where("dataset_id = ?", datasetID).
		Where("slug ILIKE ? OR iso_alpha2 ILIKE ? OR iso_alpha3 ILIKE ? OR name::text ILIKE ?",
			searchQuery, searchQuery, searchQuery, searchQuery)

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := dbQuery.Order("slug ASC").Limit(limit).Offset(offset).Find(&countries).Error
	return countries, total, err
}

func (r *GeographyRepository) ListRandomCountries(datasetID uuid.UUID, count int) ([]model.Country, error) {
	var countries []model.Country
	err := r.DB.Where("dataset_id = ?", datasetID).Order("RANDOM()").Limit(count).Find(&countries).Error
	return countries, err
}

func (r *GeographyRepository) ListRandomCountriesFiltered(datasetID uuid.UUID, count int, independentOnly bool, continent string) ([]model.Country, error) {
	var countries []model.Country
	query := r.DB.Where("dataset_id = ?", datasetID)
	if independentOnly {
		query = query.Where("independent = ?", true)
	}
	if continent != "" {
		query = query.Where("continent = ?", continent)
	}
	err := query.Order("RANDOM()").Limit(count).Find(&countries).Error
	return countries, err
}

func (r *GeographyRepository) ListAllIndependentCountries(datasetID uuid.UUID, continent string) ([]model.Country, error) {
	var countries []model.Country
	query := r.DB.Where("dataset_id = ? AND independent = ?", datasetID, true)
	if continent != "" {
		query = query.Where("continent = ?", continent)
	}
	err := query.Order("slug ASC").Find(&countries).Error
	return countries, err
}

func (r *GeographyRepository) GetContinentBySlug(slug string, datasetID uuid.UUID) (*model.Continent, error) {
	var continent model.Continent
	err := r.DB.Where("slug = ? AND dataset_id = ?", slug, datasetID).First(&continent).Error
	if err != nil {
		return nil, err
	}
	return &continent, nil
}

func (r *GeographyRepository) ListContinents(datasetID uuid.UUID) ([]model.Continent, error) {
	var continents []model.Continent
	err := r.DB.Where("dataset_id = ?", datasetID).Order("slug ASC").Find(&continents).Error
	return continents, err
}

func (r *GeographyRepository) CountContinents(datasetID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB.Model(&model.Continent{}).Where("dataset_id = ?", datasetID).Count(&count).Error
	return count, err
}

func (r *GeographyRepository) GetRegionBySlug(slug string, datasetID uuid.UUID) (*model.Region, error) {
	var region model.Region
	err := r.DB.Where("slug = ? AND dataset_id = ?", slug, datasetID).First(&region).Error
	if err != nil {
		return nil, err
	}
	return &region, nil
}

func (r *GeographyRepository) ListRegions(datasetID uuid.UUID) ([]model.Region, error) {
	var regions []model.Region
	err := r.DB.Where("dataset_id = ?", datasetID).Order("slug ASC").Find(&regions).Error
	return regions, err
}

func (r *GeographyRepository) ListRegionsByContinent(datasetID uuid.UUID, continent string) ([]model.Region, error) {
	var regions []model.Region
	err := r.DB.Where("dataset_id = ? AND continent = ?", datasetID, continent).Order("slug ASC").Find(&regions).Error
	return regions, err
}

func (r *GeographyRepository) CountRegions(datasetID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB.Model(&model.Region{}).Where("dataset_id = ?", datasetID).Count(&count).Error
	return count, err
}

func (r *GeographyRepository) CountCountries(datasetID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB.Model(&model.Country{}).Where("dataset_id = ?", datasetID).Count(&count).Error
	return count, err
}

func (r *GeographyRepository) GetDB() *gorm.DB {
	return r.DB
}

type FlagStorageRepositoryInterface interface {
	GetFlagURL(countryCode string) (string, error)
	GetFlagFile(countryCode string) ([]byte, string, error)
	GetFlagFilePNGNative(countryCode string, format string) ([]byte, error)
	UploadFlag(countryCode string, svgContent []byte) (string, error)
	UploadFlagFromURL(countryCode string, url string) (string, error)
	UploadFlagPNGFromURL(countryCode string, format string, url string) (string, error)
	DeleteFlag(countryCode string) error
	FlagExists(countryCode string) (bool, error)
}

type FlagStorageAdapter struct {
	minioClient storage.MinIOClientInterface
	logger      *zap.Logger
}

func NewFlagStorageAdapter(minioClient storage.MinIOClientInterface, logger *zap.Logger) *FlagStorageAdapter {
	return &FlagStorageAdapter{
		minioClient: minioClient,
		logger:      logger,
	}
}

func (f *FlagStorageAdapter) GetFlagURL(countryCode string) (string, error) {
	return f.minioClient.GetFlagURL(countryCode)
}

func (f *FlagStorageAdapter) GetFlagFile(countryCode string) ([]byte, string, error) {
	reader, err := f.minioClient.GetFlagFile(countryCode)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		if err := reader.Close(); err != nil {
			f.logger.Error("Error closing flag file reader", zap.Error(err))
		}
	}()
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", err
	}
	return content, "image/svg+xml", nil
}

func (f *FlagStorageAdapter) GetFlagFilePNGNative(countryCode string, format string) ([]byte, error) {
	reader, err := f.minioClient.GetFlagPNGFile(countryCode, format)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := reader.Close(); err != nil {
			f.logger.Error("Error closing PNG flag file reader", zap.Error(err))
		}
	}()
	return io.ReadAll(reader)
}

func (f *FlagStorageAdapter) UploadFlag(countryCode string, svgContent []byte) (string, error) {
	return f.minioClient.UploadFlag(countryCode, svgContent)
}

func (f *FlagStorageAdapter) UploadFlagFromURL(countryCode string, url string) (string, error) {
	return f.minioClient.UploadFlagFromURL(countryCode, url)
}

func (f *FlagStorageAdapter) UploadFlagPNGFromURL(countryCode string, format string, url string) (string, error) {
	return f.minioClient.UploadFlagPNGFromURL(countryCode, format, url)
}

func (f *FlagStorageAdapter) DeleteFlag(countryCode string) error {
	return f.minioClient.DeleteFlag(countryCode)
}

func (f *FlagStorageAdapter) FlagExists(countryCode string) (bool, error) {
	return f.minioClient.FlagExists(countryCode)
}
