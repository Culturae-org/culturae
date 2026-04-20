// backend/internal/usecase/admin/imports.go

package admin

import (
	"encoding/json"
	"fmt"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/crypto"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/identifier"
	"github.com/Culturae-org/culturae/internal/repository"
	admin "github.com/Culturae-org/culturae/internal/repository/admin"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	DatasetTypeQuestions = "questions"
	DatasetTypeGeography = "geography"
)

type AdminImportsUsecase struct {
	logger      *zap.Logger
	importsRepo admin.ImportJobRepositoryInterface
	Repo        repository.UserRepositoryInterface
	SessionRepo repository.SessionRepositoryInterface

	adminQuestionsUsecase *AdminQuestionsUsecase
	adminGeographyUsecase *AdminGeographyUsecase
}

func NewAdminImportsUsecase(
	logger *zap.Logger,
	importsRepo admin.ImportJobRepositoryInterface,
	repo repository.UserRepositoryInterface,
	sessionRepo repository.SessionRepositoryInterface,
	adminQuestions *AdminQuestionsUsecase,
	adminGeography *AdminGeographyUsecase,
) *AdminImportsUsecase {
	return &AdminImportsUsecase{
		logger:                logger,
		importsRepo:           importsRepo,
		Repo:                  repo,
		SessionRepo:           sessionRepo,
		adminQuestionsUsecase: adminQuestions,
		adminGeographyUsecase: adminGeography,
	}
}

// -----------------------------------------------
// Admin Imports Usecase Methods
//
// - ImportFromManifest
// - ListImportJobs
// - GetImportJob
// - GetImportJobLogs
// - GetImportJobLogsByAction
// - GetImportStats
// - GetAllUsersForExport
// - HashPassword
// - CreateUserFromImport
//
// -----------------------------------------------

func (uc *AdminImportsUsecase) ImportFromManifest(manifestURL string) (interface{}, error) {
	uc.logger.Info("Fetching manifest for import", zap.String("url", manifestURL))

	manifestResp := httputil.FetchURL(manifestURL)
	if manifestResp.Error != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", manifestResp.Error)
	}

	var manifest model.DatasetManifest
	if err := json.Unmarshal([]byte(manifestResp.Body), &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	if err := uc.validateManifest(&manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	uc.logger.Info("Manifest validated successfully",
		zap.String("type", manifest.Type),
		zap.String("dataset", manifest.Dataset),
		zap.String("version", manifest.Version),
	)

	switch manifest.Type {
	case DatasetTypeQuestions:
		if uc.adminQuestionsUsecase == nil {
			return nil, fmt.Errorf("questions import handler not configured")
		}
		return uc.adminQuestionsUsecase.ImportFromManifest(manifestURL)

	case DatasetTypeGeography:
		if uc.adminGeographyUsecase == nil {
			return nil, fmt.Errorf("geography import handler not configured")
		}
		return uc.adminGeographyUsecase.ImportGeographyFromManifest(manifestURL)

	default:
		return nil, fmt.Errorf("unsupported dataset type: %s (supported: questions, geography)", manifest.Type)
	}
}

func (uc *AdminImportsUsecase) validateManifest(manifest *model.DatasetManifest) error {
	if manifest.Type == "" {
		return fmt.Errorf("manifest.type is required")
	}
	if manifest.Dataset == "" {
		return fmt.Errorf("manifest.dataset is required")
	}
	if manifest.Version == "" {
		return fmt.Errorf("manifest.version is required")
	}

	if manifest.SchemaVersion != "" {
		if err := uc.validateSchemaVersion(manifest.SchemaVersion, manifest.Type); err != nil {
			return err
		}
	}

	switch manifest.Type {
	case DatasetTypeQuestions:
		if !contains(manifest.Includes, "questions") {
			return fmt.Errorf("questions manifest must include 'questions' in includes array")
		}
	case DatasetTypeGeography:
		if !contains(manifest.Includes, "countries") {
			return fmt.Errorf("geography manifest must include 'countries' in includes array")
		}
	default:
		return fmt.Errorf("unsupported manifest type: %s", manifest.Type)
	}

	return nil
}

func (uc *AdminImportsUsecase) validateSchemaVersion(schemaVersion string, datasetType string) error {
	parts := strings.Split(schemaVersion, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid schema_version format: %s (expected prefix/major.minor.patch)", schemaVersion)
	}

	prefix := parts[0]
	versionPart := parts[1]

	versionParts := strings.Split(versionPart, ".")
	if len(versionParts) < 1 {
		return fmt.Errorf("invalid schema_version format: %s", schemaVersion)
	}

	majorVersion, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return fmt.Errorf("invalid major version in schema_version: %s", schemaVersion)
	}

	switch datasetType {
	case DatasetTypeQuestions:
		if prefix != "qcm" {
			return fmt.Errorf("invalid schema_version prefix for questions dataset: %s (expected 'qcm')", prefix)
		}
		if majorVersion != 1 {
			return fmt.Errorf("unsupported schema_version major version for questions: %d (expected 1)", majorVersion)
		}
	case DatasetTypeGeography:
		if prefix != "reference" {
			return fmt.Errorf("invalid schema_version prefix for geography dataset: %s (expected 'reference')", prefix)
		}
		if majorVersion != 1 {
			return fmt.Errorf("unsupported schema_version major version for geography: %d (expected 1)", majorVersion)
		}
	default:
		return fmt.Errorf("cannot validate schema_version for unsupported dataset type: %s", datasetType)
	}

	return nil
}

func (uc *AdminImportsUsecase) ListImportJobs(limit, offset int, datasetType *string) ([]model.ImportJob, int64, error) {
	return uc.importsRepo.ListImportJobs(limit, offset, datasetType)
}

func (uc *AdminImportsUsecase) GetImportJob(jobID uuid.UUID) (*model.ImportJob, error) {
	return uc.importsRepo.GetImportJob(jobID)
}

func (uc *AdminImportsUsecase) GetImportJobLogs(jobID uuid.UUID, limit, offset int) ([]model.ImportQuestionLog, int64, error) {
	return uc.importsRepo.GetImportJobLogs(jobID, limit, offset)
}

func (uc *AdminImportsUsecase) GetImportJobLogsByAction(jobID uuid.UUID, action string, limit, offset int) ([]model.ImportQuestionLog, int64, error) {
	return uc.importsRepo.GetImportJobLogsByAction(jobID, action, limit, offset)
}

func (uc *AdminImportsUsecase) GetImportStats() (map[string]interface{}, error) {
	return uc.importsRepo.GetImportStats()
}

func (uc *AdminImportsUsecase) GetAllUsersForExport() ([]model.User, error) {
	return uc.Repo.GetAllUsers()
}

func (uc *AdminImportsUsecase) HashPassword(password string) (string, error) {
	return crypto.HashPassword(password, model.GetArgonParams())
}

func (uc *AdminImportsUsecase) CreateUserFromImport(user *model.User, preserveDates bool, createdAt, updatedAt time.Time) (*model.User, error) {
	if user.PublicID == "" {
		publicID, err := identifier.GenerateUniquePublicID(uc.Repo)
		if err != nil {
			return nil, err
		}
		user.PublicID = publicID
	}

	if err := uc.Repo.CreateWithoutHash(user); err != nil {
		return nil, err
	}

	if preserveDates && !createdAt.IsZero() {
		if err := uc.Repo.UpdateDates(user.ID, createdAt, updatedAt); err != nil {
			_ = err
			return user, nil
		}
	}

	return user, nil
}
