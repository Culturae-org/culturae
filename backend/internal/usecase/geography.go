// backend/internal/usecase/geography.go

package usecase

import (
	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type GeographyUsecase struct {
	repo        repository.GeographyRepositoryInterface
	logger      *zap.Logger
	flagStorage repository.FlagStorageRepositoryInterface
}

func NewGeographyUsecase(
	repo repository.GeographyRepositoryInterface,
	logger *zap.Logger,
	flagStorage repository.FlagStorageRepositoryInterface,
) *GeographyUsecase {
	return &GeographyUsecase{
		repo:        repo,
		logger:      logger,
		flagStorage: flagStorage,
	}
}

// -----------------------------------------------
// Geography Usecase Methods
//
// - GetDefaultDataset
// - ListCountries
// - ListCountriesWithFilters
// - GetCountryBySlug
// - ListCountriesByContinent
// - ListCountriesByRegion
// - SearchCountries
// - ListContinents
// - GetContinentBySlug
// - ListRegions
// - ListRegionsByContinent
// - GetRegionBySlug
//
// -----------------------------------------------

func (u *GeographyUsecase) GetDefaultDataset() (*model.GeographyDataset, error) {
	return u.repo.GetDefaultDataset()
}

func (u *GeographyUsecase) ListCountries(datasetID uuid.UUID, limit, offset int) ([]model.Country, int64, error) {
	return u.repo.ListCountries(datasetID, limit, offset)
}

func (u *GeographyUsecase) ListCountriesWithFilters(datasetID uuid.UUID, filters model.CountryFilters, limit, offset int) ([]model.Country, int64, error) {
	return u.repo.ListCountriesWithFilters(datasetID, filters, limit, offset)
}

func (u *GeographyUsecase) GetCountryBySlug(slug string, datasetID uuid.UUID) (*model.Country, error) {
	return u.repo.GetCountryBySlug(slug, datasetID)
}

func (u *GeographyUsecase) ListCountriesByContinent(datasetID uuid.UUID, continent string) ([]model.Country, error) {
	return u.repo.ListCountriesByContinent(datasetID, continent)
}

func (u *GeographyUsecase) ListCountriesByRegion(datasetID uuid.UUID, region string) ([]model.Country, error) {
	return u.repo.ListCountriesByRegion(datasetID, region)
}

func (u *GeographyUsecase) SearchCountries(datasetID uuid.UUID, query string, limit, offset int) ([]model.Country, int64, error) {
	return u.repo.SearchCountries(datasetID, query, limit, offset)
}

func (u *GeographyUsecase) ListContinents(datasetID uuid.UUID) ([]model.Continent, error) {
	return u.repo.ListContinents(datasetID)
}

func (u *GeographyUsecase) GetContinentBySlug(slug string, datasetID uuid.UUID) (*model.Continent, error) {
	return u.repo.GetContinentBySlug(slug, datasetID)
}

func (u *GeographyUsecase) ListRegions(datasetID uuid.UUID) ([]model.Region, error) {
	return u.repo.ListRegions(datasetID)
}

func (u *GeographyUsecase) ListRegionsByContinent(datasetID uuid.UUID, continent string) ([]model.Region, error) {
	return u.repo.ListRegionsByContinent(datasetID, continent)
}

func (u *GeographyUsecase) GetRegionBySlug(slug string, datasetID uuid.UUID) (*model.Region, error) {
	return u.repo.GetRegionBySlug(slug, datasetID)
}

func (u *GeographyUsecase) GetFlagURL(countryCode string) (string, error) {
	return u.flagStorage.GetFlagURL(countryCode)
}

func (u *GeographyUsecase) GetFlagFile(countryCode string) ([]byte, string, error) {
	return u.flagStorage.GetFlagFile(countryCode)
}

func (u *GeographyUsecase) GetFlagFilePNGNative(countryCode string, format string) ([]byte, error) {
	return u.flagStorage.GetFlagFilePNGNative(countryCode, format)
}
