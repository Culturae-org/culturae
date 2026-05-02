// backend/internal/usecase/admin/datasets.go

package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/repository"
	adminRepo "github.com/Culturae-org/culturae/internal/repository/admin"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AdminDatasetsUsecase struct {
	questionDatasetRepo repository.QuestionDatasetRepositoryInterface
	questionRepo        repository.QuestionRepositoryInterface
	adminQuestionRepo   adminRepo.AdminQuestionRepositoryInterface
	geographyRepo       adminRepo.AdminGeographyRepositoryInterface
	adminLogsRepo       adminRepo.AdminLogsRepositoryInterface
	logger              *zap.Logger
}

func NewAdminDatasetsUsecase(
	questionDatasetRepo repository.QuestionDatasetRepositoryInterface,
	questionRepo repository.QuestionRepositoryInterface,
	adminQuestionRepo adminRepo.AdminQuestionRepositoryInterface,
	geographyRepo adminRepo.AdminGeographyRepositoryInterface,
	adminLogsRepo adminRepo.AdminLogsRepositoryInterface,
	logger *zap.Logger,
) *AdminDatasetsUsecase {
	return &AdminDatasetsUsecase{
		questionDatasetRepo: questionDatasetRepo,
		questionRepo:        questionRepo,
		adminQuestionRepo:   adminQuestionRepo,
		geographyRepo:       geographyRepo,
		adminLogsRepo:       adminLogsRepo,
		logger:              logger,
	}
}

// -----------------------------------------------
// Admin Datasets Usecase Methods
//
// - ListAllDatasets
// - GetDataset
// - DeleteDataset
// - SetDefaultDataset
// - GetDefaultDataset
// - GetDatasetStatistics
// - ListQuestionDatasets
// - GetQuestionDataset
// - GetQuestionDatasetBySlug
// - GetQuestionDatasetByManifestURL
// - CreateQuestionDataset
// - UpdateQuestionDataset
// - UpdateQuestionDatasetStatistics
// - GetQuestionDatasetQuestions
// - CheckAllDatasetsForUpdates
// - CheckForUpdates
// - CheckAllForUpdates
//
// -----------------------------------------------

func (u *AdminDatasetsUsecase) ListAllDatasets(datasetType *string) ([]model.UnifiedDataset, error) {
	var unified []model.UnifiedDataset

	if datasetType == nil || *datasetType == DatasetTypeQuestions {
		questionDatasets, err := u.questionDatasetRepo.ListDatasets(false)
		if err != nil {
			u.logger.Error("Failed to list question datasets", zap.Error(err))
		} else {
			for _, qd := range questionDatasets {
				unified = append(unified, model.UnifiedDataset{
					ID:                     qd.ID,
					Type:                   DatasetTypeQuestions,
					Slug:                   qd.Slug,
					Name:                   qd.Name,
					Description:            qd.Description,
					Version:                qd.Version,
					Source:                 qd.Source,
					ManifestURL:            qd.ManifestURL,
					IsActive:               qd.IsActive,
					IsDefault:              qd.IsDefault,
					ImportedAt:             qd.ImportedAt,
					UpdateAvailable:        qd.LatestAvailableVersion != "" && qd.LatestAvailableVersion != qd.Version,
					LatestAvailableVersion: qd.LatestAvailableVersion,
					Stats: map[string]interface{}{
						keyQuestionCount: qd.QuestionCount,
						keyThemeCount:    qd.ThemeCount,
					},
				})
			}
		}
	}

	if datasetType == nil || *datasetType == DatasetTypeGeography {
		geoDatasets, err := u.geographyRepo.ListDatasets(false)
		if err != nil {
			u.logger.Error("Failed to list geography datasets", zap.Error(err))
		} else {
			for _, gd := range geoDatasets {
				unified = append(unified, model.UnifiedDataset{
					ID:                     gd.ID,
					Type:                   DatasetTypeGeography,
					Slug:                   gd.Slug,
					Name:                   gd.Name,
					Description:            gd.Description,
					Version:                gd.Version,
					Source:                 gd.Source,
					ManifestURL:            gd.ManifestURL,
					IsActive:               gd.IsActive,
					IsDefault:              gd.IsDefault,
					ImportedAt:             gd.ImportedAt,
					UpdateAvailable:        gd.LatestAvailableVersion != "" && gd.LatestAvailableVersion != gd.Version,
					LatestAvailableVersion: gd.LatestAvailableVersion,
					Stats: map[string]interface{}{
						keyCountryCount:   gd.CountryCount,
						keyContinentCount: gd.ContinentCount,
						keyRegionCount:    gd.RegionCount,
						keyFlagCount:      gd.FlagCount,
					},
				})
			}
		}
	}

	u.logger.Info("Listed all datasets",
		zap.Int("total_count", len(unified)),
		zap.Stringp("type_filter", datasetType),
	)

	return unified, nil
}

func (u *AdminDatasetsUsecase) GetDataset(datasetType string, id uuid.UUID) (*model.UnifiedDataset, error) {
	switch datasetType {
	case DatasetTypeQuestions:
		qd, err := u.questionDatasetRepo.GetDatasetByID(id)
		if err != nil {
			return nil, err
		}
		return &model.UnifiedDataset{
			ID:          qd.ID,
			Type:        DatasetTypeQuestions,
			Slug:        qd.Slug,
			Name:        qd.Name,
			Description: qd.Description,
			Version:     qd.Version,
			Source:      qd.Source,
			ManifestURL: qd.ManifestURL,
			IsActive:    qd.IsActive,
			IsDefault:   qd.IsDefault,
			ImportedAt:  qd.ImportedAt,
			Stats: map[string]interface{}{
				keyQuestionCount: qd.QuestionCount,
				keyThemeCount:    qd.ThemeCount,
			},
		}, nil

	case DatasetTypeGeography:
		gd, err := u.geographyRepo.GetDatasetByID(id)
		if err != nil {
			return nil, err
		}
		return &model.UnifiedDataset{
			ID:          gd.ID,
			Type:        DatasetTypeGeography,
			Slug:        gd.Slug,
			Name:        gd.Name,
			Description: gd.Description,
			Version:     gd.Version,
			Source:      gd.Source,
			ManifestURL: gd.ManifestURL,
			IsActive:    gd.IsActive,
			IsDefault:   gd.IsDefault,
			ImportedAt:  gd.ImportedAt,
			Stats: map[string]interface{}{
				keyCountryCount:   gd.CountryCount,
				keyContinentCount: gd.ContinentCount,
				keyRegionCount:    gd.RegionCount,
				keyFlagCount:      gd.FlagCount,
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported dataset type: %s (supported: questions, geography)", datasetType)
	}
}

func (u *AdminDatasetsUsecase) DeleteDataset(datasetType string, id uuid.UUID, force bool) error {
	u.logger.Info("Deleting dataset",
		zap.String("type", datasetType),
		zap.String("id", id.String()),
		zap.Bool("force", force),
	)

	switch datasetType {
	case DatasetTypeQuestions:
		dataset, err := u.questionDatasetRepo.GetDatasetByID(id)
		if err != nil {
			if errors.Is(err, model.ErrDatasetNotFound) {
				u.logger.Info("Dataset already deleted, skipping deletion", zap.String("id", id.String()))
				return nil
			}
			return err
		}

		count, err := u.questionDatasetRepo.CountOtherActiveDatasets(id)
		if err != nil {
			return err
		}

		if count == 0 {
			return fmt.Errorf("cannot delete the only dataset of this type")
		}

		if dataset.IsDefault {
			return fmt.Errorf("cannot delete the default dataset")
		}

		questions, _, err := u.questionRepo.ListQuestionsByDataset(&id, 0, 0)
		if err != nil {
			return err
		}
		for _, q := range questions {
			if err := u.adminQuestionRepo.DeleteQuestionSubthemes(q.ID); err != nil {
				return err
			}
			if err := u.adminQuestionRepo.DeleteQuestionTags(q.ID); err != nil {
				return err
			}
		}

		questionsToDelete, _, err := u.questionRepo.ListQuestionsByDataset(&id, 0, 0)
		if err != nil {
			return err
		}
		for _, q := range questionsToDelete {
			if err := u.adminQuestionRepo.Delete(q.ID); err != nil {
				return err
			}
		}

		return u.questionDatasetRepo.DeleteDataset(id)

	case DatasetTypeGeography:
		dataset, err := u.geographyRepo.GetDatasetByID(id)
		if err != nil {
			if errors.Is(err, model.ErrDatasetNotFound) {
				u.logger.Info("Geography dataset already deleted, skipping deletion", zap.String("id", id.String()))
				return nil
			}
			return err
		}

		count, err := u.geographyRepo.CountCountries(id)
		if err != nil {
			return err
		}

		if count == 0 && false {
			return fmt.Errorf("cannot delete the only dataset of this type")
		}

		if dataset.IsDefault {
			return fmt.Errorf("cannot delete the default dataset")
		}

		if err := u.geographyRepo.DeleteCountriesByDataset(id); err != nil {
			return err
		}
		if err := u.geographyRepo.DeleteContinentsByDataset(id); err != nil {
			return err
		}
		if err := u.geographyRepo.DeleteRegionsByDataset(id); err != nil {
			return err
		}
		return u.geographyRepo.DeleteDataset(id)

	default:
		return fmt.Errorf("unsupported dataset type: %s (supported: questions, geography)", datasetType)
	}
}

func (u *AdminDatasetsUsecase) SetDefaultDataset(datasetType string, id uuid.UUID) error {
	u.logger.Info("Setting default dataset",
		zap.String("type", datasetType),
		zap.String("id", id.String()),
	)

	switch datasetType {
	case DatasetTypeQuestions:
		return u.questionDatasetRepo.SetDefaultDataset(id)

	case DatasetTypeGeography:
		return u.geographyRepo.SetDefaultDataset(id)

	default:
		return fmt.Errorf("unsupported dataset type: %s (supported: questions, geography)", datasetType)
	}
}

func (u *AdminDatasetsUsecase) GetDefaultDataset(datasetType string) (*model.UnifiedDataset, error) {
	switch datasetType {
	case DatasetTypeQuestions:
		qd, err := u.questionDatasetRepo.GetDefaultDataset()
		if err != nil {
			return nil, err
		}
		return &model.UnifiedDataset{
			ID:          qd.ID,
			Type:        DatasetTypeQuestions,
			Slug:        qd.Slug,
			Name:        qd.Name,
			Description: qd.Description,
			Version:     qd.Version,
			Source:      qd.Source,
			ManifestURL: qd.ManifestURL,
			IsActive:    qd.IsActive,
			IsDefault:   qd.IsDefault,
			ImportedAt:  qd.ImportedAt,
		}, nil

	case DatasetTypeGeography:
		gd, err := u.geographyRepo.GetDefaultDataset()
		if err != nil {
			return nil, err
		}
		return &model.UnifiedDataset{
			ID:          gd.ID,
			Type:        DatasetTypeGeography,
			Slug:        gd.Slug,
			Name:        gd.Name,
			Description: gd.Description,
			Version:     gd.Version,
			Source:      gd.Source,
			ManifestURL: gd.ManifestURL,
			IsActive:    gd.IsActive,
			IsDefault:   gd.IsDefault,
			ImportedAt:  gd.ImportedAt,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported dataset type: %s (supported: questions, geography)", datasetType)
	}
}

func (u *AdminDatasetsUsecase) GetDatasetStatistics(datasetType string, id uuid.UUID) (map[string]interface{}, error) {
	u.logger.Info("Getting dataset statistics",
		zap.String("type", datasetType),
		zap.String("id", id.String()),
	)

	switch datasetType {
	case DatasetTypeQuestions:
		dataset, err := u.questionDatasetRepo.GetDatasetByID(id)
		if err != nil {
			return nil, err
		}

		totalQuestions, err := u.questionRepo.CountByDataset(id, "")
		if err != nil {
			totalQuestions = 0
		}

		byDifficulty, _ := u.questionRepo.GetQuestionCountByDifficulty(id)
		byTheme, _ := u.questionRepo.GetQuestionCountByTheme(id)

		themeCount, _ := u.questionDatasetRepo.GetThemeCountByDataset(id)

		return map[string]interface{}{
			"dataset": map[string]interface{}{
				"id":          dataset.ID,
				keySlug:      dataset.Slug,
				keyName:        dataset.Name,
				keyVersion:     dataset.Version,
				keyIsActive:   dataset.IsActive,
				keyIsDefault:  dataset.IsDefault,
				keyImportedAt: dataset.ImportedAt,
			},
			"questions": map[string]interface{}{
				"total":         totalQuestions,
				"by_difficulty": byDifficulty,
				"by_theme":      byTheme,
			},
			"themes": map[string]interface{}{
				"total": themeCount,
			},
		}, nil

	case DatasetTypeGeography:
		dataset, err := u.geographyRepo.GetDatasetByID(id)
		if err != nil {
			return nil, err
		}

		countryCount, _ := u.geographyRepo.CountCountries(id)
		continentCount, _ := u.geographyRepo.CountContinents(id)
		regionCount, _ := u.geographyRepo.CountRegions(id)

		return map[string]interface{}{
			"dataset_id":      id,
			keyName:            dataset.Name,
			keyVersion:         dataset.Version,
			keyCountryCount:   countryCount,
			keyContinentCount: continentCount,
			keyRegionCount:    regionCount,
			keyFlagCount:      dataset.FlagCount,
			keyIsActive:       dataset.IsActive,
			keyIsDefault:      dataset.IsDefault,
			keyImportedAt:     dataset.ImportedAt,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported dataset type: %s (supported: questions, geography)", datasetType)
	}
}

func (u *AdminDatasetsUsecase) ListQuestionDatasets(activeOnly bool) ([]model.QuestionDataset, error) {
	return u.questionDatasetRepo.ListDatasets(activeOnly)
}

func (u *AdminDatasetsUsecase) GetQuestionDataset(id uuid.UUID) (*model.QuestionDataset, error) {
	return u.questionDatasetRepo.GetDatasetByID(id)
}

func (u *AdminDatasetsUsecase) GetQuestionDatasetBySlug(slug string) (*model.QuestionDataset, error) {
	return u.questionDatasetRepo.GetDatasetBySlug(slug)
}

func (u *AdminDatasetsUsecase) GetQuestionDatasetByManifestURL(manifestURL string) (*model.QuestionDataset, error) {
	return u.questionDatasetRepo.GetDatasetByManifestURL(manifestURL)
}

func (u *AdminDatasetsUsecase) CreateQuestionDataset(req *model.CreateDatasetRequest) (*model.QuestionDataset, error) {
	if !model.ValidateSlug(req.Slug) {
		return nil, fmt.Errorf("invalid slug format: must contain only lowercase letters, numbers, and hyphens")
	}

	exists, err := u.questionDatasetRepo.ExistsBySlug(req.Slug, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("dataset with slug '%s' already exists", req.Slug)
	}

	dataset := &model.QuestionDataset{
		ID:          uuid.New(),
		Slug:        req.Slug,
		Name:        req.Name,
		Description: req.Description,
		Version:     req.Version,
		Source:      req.Source,
		IsActive:    true,
		IsDefault:   req.IsDefault,
		ImportedAt:  time.Now(),
	}

	if err := u.questionDatasetRepo.CreateDataset(dataset); err != nil {
		u.logger.Error("Failed to create question dataset", zap.Error(err))
		return nil, err
	}

	if req.IsDefault {
		if err := u.questionDatasetRepo.SetDefaultDataset(dataset.ID); err != nil {
			u.logger.Warn("Failed to set default dataset", zap.Error(err))
		}
	}

	u.logger.Info("Question dataset created",
		zap.String(keySlug,dataset.Slug),
		zap.String(keyName, dataset.Name),
		zap.String(keyVersion, dataset.Version),
		zap.Bool(keyIsDefault, dataset.IsDefault),
	)

	return dataset, nil
}

func (u *AdminDatasetsUsecase) UpdateQuestionDataset(id uuid.UUID, req *model.UpdateDatasetRequest) (*model.QuestionDataset, error) {
	_, err := u.questionDatasetRepo.GetDatasetByID(id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.Name != nil {
		updates[keyName] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsActive != nil {
		updates[keyIsActive] = *req.IsActive
	}
	if req.IsDefault != nil && *req.IsDefault {
		if err := u.questionDatasetRepo.SetDefaultDataset(id); err != nil {
			u.logger.Warn("Failed to unset previous defaults", zap.Error(err))
		}
		updates[keyIsDefault] = true
	}

	if len(updates) > 0 {
		if err := u.questionDatasetRepo.UpdateDataset(id, updates); err != nil {
			u.logger.Error("Failed to update question dataset", zap.Error(err))
			return nil, err
		}
	}

	return u.questionDatasetRepo.GetDatasetByID(id)
}

func (u *AdminDatasetsUsecase) UpdateQuestionDatasetStatistics(datasetID uuid.UUID) error {
	questionCount, err := u.questionRepo.CountByDataset(datasetID, "")
	if err != nil {
		return err
	}

	themeCount, err := u.questionDatasetRepo.GetThemeCountByDataset(datasetID)
	if err != nil {
		return err
	}

	return u.questionDatasetRepo.UpdateDataset(datasetID, map[string]interface{}{
		keyQuestionCount: questionCount,
		keyThemeCount:    themeCount,
	})
}

func (u *AdminDatasetsUsecase) GetQuestionDatasetQuestions(datasetID uuid.UUID, limit, offset int) ([]*model.Question, int64, error) {
	return u.questionRepo.ListQuestionsByDataset(&datasetID, limit, offset)
}

type DatasetWithUpdates struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Type           string    `json:"type"`
	Version        string    `json:"version"`
	ManifestURL    string    `json:"manifest_url"`
	HasUpdate      bool      `json:"has_update"`
	CurrentVersion string    `json:"current_version"`
	LatestVersion  string    `json:"latest_version"`
}

func (u *AdminDatasetsUsecase) CheckAllDatasetsForUpdates() ([]DatasetWithUpdates, error) {
	var results []DatasetWithUpdates

	questionDatasets, err := u.questionDatasetRepo.ListDatasets(false)
	if err != nil {
		u.logger.Error("Failed to fetch question datasets", zap.Error(err))
		return nil, err
	}

	for _, qd := range questionDatasets {
		if qd.ManifestURL == "" {
			continue
		}
		updateInfo, err := u.CheckForUpdates(qd.ManifestURL, DatasetTypeQuestions, false)
		if err != nil {
			u.logger.Warn("Failed to check updates for dataset", zap.String("id", qd.ID.String()), zap.Error(err))
			continue
		}

		results = append(results, DatasetWithUpdates{
			ID:             qd.ID,
			Name:           qd.Name,
			Slug:           qd.Slug,
			Type:           DatasetTypeQuestions,
			Version:        qd.Version,
			ManifestURL:    qd.ManifestURL,
			HasUpdate:      updateInfo.HasUpdate,
			CurrentVersion: updateInfo.CurrentVersion,
			LatestVersion:  updateInfo.LatestVersion,
		})
	}

	geographyDatasets, err := u.geographyRepo.ListDatasets(false)
	if err != nil {
		u.logger.Error("Failed to fetch geography datasets", zap.Error(err))
		return nil, err
	}

	for _, gd := range geographyDatasets {
		if gd.ManifestURL == "" {
			continue
		}
		updateInfo, err := u.CheckForUpdates(gd.ManifestURL, DatasetTypeGeography, false)
		if err != nil {
			u.logger.Warn("Failed to check updates for dataset", zap.String("id", gd.ID.String()), zap.Error(err))
			continue
		}

		results = append(results, DatasetWithUpdates{
			ID:             gd.ID,
			Name:           gd.Name,
			Slug:           gd.Slug,
			Type:           DatasetTypeGeography,
			Version:        gd.Version,
			ManifestURL:    gd.ManifestURL,
			HasUpdate:      updateInfo.HasUpdate,
			CurrentVersion: updateInfo.CurrentVersion,
			LatestVersion:  updateInfo.LatestVersion,
		})
	}

	return results, nil
}

func (u *AdminDatasetsUsecase) CheckForUpdates(manifestURL string, datasetType string, isAutomatic bool) (*model.DatasetUpdateInfo, error) {
	if isAutomatic {
		u.logger.Info("Checking for updates from manifest",
			zap.String("manifest_url", manifestURL),
			zap.String("dataset_type", datasetType),
		)
	}

	manifestResp := httputil.FetchURL(manifestURL)
	if manifestResp.Error != nil {
		u.logger.Error("Failed to fetch manifest", zap.Error(manifestResp.Error))
		return nil, fmt.Errorf("failed to fetch manifest: %w", manifestResp.Error)
	}

	var manifest model.DatasetManifest
	if err := json.Unmarshal([]byte(manifestResp.Body), &manifest); err != nil {
		u.logger.Error("Failed to parse manifest", zap.Error(err))
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	slug := fmt.Sprintf("%s-v%s", manifest.Dataset, manifest.Version)
	var hasUpdate bool
	currentVersion := ""

	switch datasetType {
	case DatasetTypeQuestions:
		existing, err := u.questionDatasetRepo.GetDatasetBySlug(slug)
		if err == nil && existing != nil {
			hasUpdate = false
			currentVersion = manifest.Version
			if isAutomatic {
				u.logger.Info("Dataset version already exists",
					zap.String(keySlug,slug),
					zap.String(keyVersion, manifest.Version),
				)
			}
		} else {
			allDatasets, err := u.questionDatasetRepo.ListDatasets(false)
			if err == nil && len(allDatasets) > 0 {
				baseSlug := manifest.Dataset + "-v"
				var existingDatasets []model.QuestionDataset
				for _, ds := range allDatasets {
					if len(ds.Slug) > len(baseSlug) && ds.Slug[:len(baseSlug)] == baseSlug && ds.ManifestURL != "" {
						existingDatasets = append(existingDatasets, ds)
					}
				}
				if len(existingDatasets) > 0 {
					latestExisting := existingDatasets[0]
					for _, ds := range existingDatasets {
						if ds.Version > latestExisting.Version {
							latestExisting = ds
						}
					}
					currentVersion = latestExisting.Version
					if manifest.Version > latestExisting.Version {
						hasUpdate = true
						if isAutomatic {
							u.logger.Info("Newer version available",
								zap.String("current_version", latestExisting.Version),
								zap.String("latest_version", manifest.Version),
							)
						}
					} else {
						hasUpdate = false
					}
				} else {
					hasUpdate = true
					if isAutomatic {
						u.logger.Info("New dataset available", zap.String("dataset", manifest.Dataset))
					}
				}
			} else {
				hasUpdate = true
				if isAutomatic {
					u.logger.Info("New dataset available", zap.String("dataset", manifest.Dataset))
				}
			}
		}

	case DatasetTypeGeography:
		existing, err := u.geographyRepo.GetDatasetBySlug(slug)
		if err == nil && existing != nil {
			hasUpdate = false
			currentVersion = manifest.Version
			if isAutomatic {
				u.logger.Info("Geography dataset version already exists",
					zap.String(keySlug,slug),
					zap.String(keyVersion, manifest.Version),
				)
			}
		} else {
			allDatasets, err := u.geographyRepo.ListDatasets(false)
			if err == nil && len(allDatasets) > 0 {
				baseSlug := manifest.Dataset + "-v"
				var existingDatasets []model.GeographyDataset
				for _, ds := range allDatasets {
					if len(ds.Slug) > len(baseSlug) && ds.Slug[:len(baseSlug)] == baseSlug && ds.ManifestURL != "" {
						existingDatasets = append(existingDatasets, ds)
					}
				}
				if len(existingDatasets) > 0 {
					latestExisting := existingDatasets[0]
					for _, ds := range existingDatasets {
						if ds.Version > latestExisting.Version {
							latestExisting = ds
						}
					}
					currentVersion = latestExisting.Version
					if manifest.Version > latestExisting.Version {
						hasUpdate = true
						if isAutomatic {
							u.logger.Info("Newer geography version available",
								zap.String("current_version", latestExisting.Version),
								zap.String("latest_version", manifest.Version),
							)
						}
					} else {
						hasUpdate = false
					}
				} else {
					hasUpdate = true
					if isAutomatic {
						u.logger.Info("New geography dataset available", zap.String("dataset", manifest.Dataset))
					}
				}
			} else {
				hasUpdate = true
				if isAutomatic {
					u.logger.Info("New geography dataset available", zap.String("dataset", manifest.Dataset))
				}
			}
		}

	default:
		return nil, fmt.Errorf("unsupported dataset type: %s", datasetType)
	}

	updateInfo := &model.DatasetUpdateInfo{
		HasUpdate:      hasUpdate,
		CurrentVersion: currentVersion,
		LatestVersion:  manifest.Version,
		UpdatedAt:      manifest.UpdatedAt,
		Manifest:       &manifest,
	}

	if hasUpdate && isAutomatic {
		u.logger.Info("New version available",
			zap.String("dataset", manifest.Dataset),
			zap.String(keyVersion, manifest.Version),
		)
	}

	return updateInfo, nil
}

func (u *AdminDatasetsUsecase) CheckAllForUpdates() ([]model.DatasetUpdateResult, error) {
	u.logger.Info("Starting automatic check for dataset updates")
	var results []model.DatasetUpdateResult
	now := time.Now()

	questionDatasets, err := u.questionDatasetRepo.ListDatasets(false)
	if err != nil {
		u.logger.Error("Failed to query question datasets for update check", zap.Error(err))
	} else {
		for _, qd := range questionDatasets {
			if qd.ManifestURL == "" {
				continue
			}
			updateInfo, err := u.CheckForUpdates(qd.ManifestURL, DatasetTypeQuestions, true)
			if err != nil {
				u.logger.Warn("Failed to check updates for question dataset",
					zap.String("dataset", qd.Name),
					zap.Error(err),
				)
				continue
			}

			updates := map[string]interface{}{"update_checked_at": now}
			if updateInfo.HasUpdate {
				updates["latest_available_version"] = updateInfo.LatestVersion
				results = append(results, model.DatasetUpdateResult{
					DatasetID:   qd.ID,
					DatasetType: DatasetTypeQuestions,
					DatasetName: qd.Name,
					OldVersion:  qd.Version,
					NewVersion:  updateInfo.LatestVersion,
				})
			}
			if err := u.questionDatasetRepo.UpdateDataset(qd.ID, updates); err != nil {
				u.logger.Warn("Failed to update question dataset", zap.Error(err))
			}
		}
	}

	geoDatasets, err := u.geographyRepo.ListDatasets(false)
	if err != nil {
		u.logger.Error("Failed to query geography datasets for update check", zap.Error(err))
	} else {
		for _, gd := range geoDatasets {
			if gd.ManifestURL == "" {
				continue
			}
			updateInfo, err := u.CheckForUpdates(gd.ManifestURL, DatasetTypeGeography, true)
			if err != nil {
				u.logger.Warn("Failed to check updates for geography dataset",
					zap.String("dataset", gd.Name),
					zap.Error(err),
				)
				continue
			}

			updates := map[string]interface{}{"update_checked_at": now}
			if updateInfo.HasUpdate {
				updates["latest_available_version"] = updateInfo.LatestVersion
				results = append(results, model.DatasetUpdateResult{
					DatasetID:   gd.ID,
					DatasetType: DatasetTypeGeography,
					DatasetName: gd.Name,
					OldVersion:  gd.Version,
					NewVersion:  updateInfo.LatestVersion,
				})
			}
			if err := u.geographyRepo.UpdateDataset(&gd); err != nil {
				u.logger.Warn("Failed to update geography dataset", zap.Error(err))
			}
		}
	}

	u.logger.Info("Dataset update check completed",
		zap.Int("updates_found", len(results)),
	)
	return results, nil
}
