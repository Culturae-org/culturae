// backend/internal/usecase/admin/questions.go

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
	"github.com/Culturae-org/culturae/internal/repository"
	admin "github.com/Culturae-org/culturae/internal/repository/admin"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

type AdminQuestionsUsecase struct {
	questionRepo        repository.QuestionRepositoryInterface
	adminQuestionRepo   admin.AdminQuestionRepositoryInterface
	questionDatasetRepo admin.AdminDatasetRepositoryInterface
	logger              *zap.Logger
	importsRepo         admin.ImportJobRepositoryInterface
}

func NewAdminQuestionsUsecase(
	questionRepo repository.QuestionRepositoryInterface,
	adminQuestionRepo admin.AdminQuestionRepositoryInterface,
	questionDatasetRepo admin.AdminDatasetRepositoryInterface,
	logger *zap.Logger,
	importsRepo admin.ImportJobRepositoryInterface,
) *AdminQuestionsUsecase {
	return &AdminQuestionsUsecase{
		questionRepo:        questionRepo,
		adminQuestionRepo:   adminQuestionRepo,
		questionDatasetRepo: questionDatasetRepo,
		logger:              logger,
		importsRepo:         importsRepo,
	}
}

// -----------------------------------------------
// Admin Questions Usecase Methods
//
// - ImportFromManifest
// - CreateQuestion
// - GetQuestionByID
// - GetQuestionBySlug
// - UpdateQuestion
// - DeleteQuestion
// - ListQuestionsByDataset
// - ListQuestionsWithFilters
// - SearchQuestions
// - GetTotalQuestions
// - BackupQuestions
// - ExportQuestionsClean
//
// -----------------------------------------------

func (u *AdminQuestionsUsecase) ImportFromManifest(manifestURL string) (*model.ImportResult, error) {
	startTime := time.Now()
	result := &model.ImportResult{
		Success: false,
		Errors:  []string{},
	}

	u.logger.Info("Starting manifest import",
		zap.String("manifest_url", manifestURL),
		zap.Time("started_at", startTime),
	)

	u.logger.Info("Fetching manifest", zap.String("url", manifestURL))
	manifestResp := httputil.FetchURL(manifestURL)
	if manifestResp.Error != nil {
		u.logger.Error("Failed to fetch manifest", zap.Error(manifestResp.Error))
		return nil, fmt.Errorf("failed to fetch manifest: %w", manifestResp.Error)
	}

	u.logger.Debug("Manifest response received",
		zap.Int("content_length", len(manifestResp.Body)),
	)

	var manifest model.DatasetManifest
	if err := json.Unmarshal([]byte(manifestResp.Body), &manifest); err != nil {
		u.logger.Error("Failed to parse manifest JSON", zap.Error(err))
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	u.logger.Info("Manifest loaded successfully",
		zap.String("dataset", manifest.Dataset),
		zap.String("version", manifest.Version),
		zap.String("schema_version", manifest.SchemaVersion),
		zap.Int("expected_questions", manifest.Counts.Questions),
		zap.Int("expected_themes", manifest.Counts.Themes),
		zap.Int("expected_subthemes", manifest.Counts.Subthemes),
		zap.Int("expected_tags", manifest.Counts.Tags),
	)

	datasetSlug := fmt.Sprintf("%s-v%s", manifest.Dataset, manifest.Version)
	datasetName := fmt.Sprintf("%s v%s", manifest.Dataset, manifest.Version)

	_, err := u.questionDatasetRepo.GetDatasetBySlug(datasetSlug)
	if err == nil {
		return nil, fmt.Errorf("dataset version %s already imported (slug: %s). Delete the existing dataset first to reimport", manifest.Version, datasetSlug)
	}

	job := model.ImportJob{
		ID:          uuid.New(),
		ManifestURL: manifestURL,
		Dataset:     manifest.Dataset,
		Version:     manifest.Version,
		StartedAt:   startTime,
	}
	if err := u.importsRepo.CreateImportJob(&job); err != nil {
		u.logger.Error("Failed to create import job record", zap.Error(err))
		return nil, fmt.Errorf("failed to create import job: %w", err)
	}

	datasets, err := u.questionDatasetRepo.ListDatasets(true)
	if err != nil {
		u.logger.Error("Failed to list existing datasets", zap.Error(err))
		return nil, fmt.Errorf("failed to list existing datasets: %w", err)
	}

	existingCount := 0
	for _, d := range datasets {
		if d.ManifestURL != "" {
			existingCount++
		}
	}

	isDefault := existingCount == 0

	manifestJSON, _ := json.Marshal(manifest)
	dataset := model.QuestionDataset{
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

	if err := u.questionDatasetRepo.CreateDataset(&dataset); err != nil {
		u.logger.Error("Failed to create dataset", zap.Error(err))
		return nil, fmt.Errorf("failed to create dataset: %w", err)
	}

	u.logger.Info("Dataset created",
		zap.String(keySlug, dataset.Slug),
		zap.String("id", dataset.ID.String()),
	)

	baseURL := manifestURL[:strings.LastIndex(manifestURL, "/")+1]
	u.logger.Debug("Base URL calculated", zap.String("base_url", baseURL))

	if contains(manifest.Includes, "themes") {
		themesURL := baseURL + "themes.ndjson"
		expectedChecksum := manifest.Checksums["themes.ndjson"]
		u.logger.Info("Importing themes", zap.String("url", themesURL))
		themesAdded, err := u.importEntitiesWithChecksum(themesURL, "themes", expectedChecksum)
		if err != nil {
			errMsg := fmt.Sprintf("themes: %v", err)
			u.logger.Error("Themes import failed", zap.Error(err))
			result.Errors = append(result.Errors, errMsg)
		} else {
			result.ThemesAdded = themesAdded
			u.logger.Info("Themes import completed", zap.Int("added", themesAdded))
			if manifest.Counts.Themes > 0 && themesAdded != manifest.Counts.Themes {
				u.logger.Warn("Themes count mismatch", zap.Int("expected", manifest.Counts.Themes), zap.Int("actual", themesAdded))
			}
		}
	}

	if contains(manifest.Includes, "subthemes") {
		subthemesURL := baseURL + "subthemes.ndjson"
		expectedChecksum := manifest.Checksums["subthemes.ndjson"]
		u.logger.Info("Importing subthemes", zap.String("url", subthemesURL))
		subthemesAdded, err := u.importEntitiesWithChecksum(subthemesURL, "subthemes", expectedChecksum)
		if err != nil {
			errMsg := fmt.Sprintf("subthemes: %v", err)
			u.logger.Error("Subthemes import failed", zap.Error(err))
			result.Errors = append(result.Errors, errMsg)
		} else {
			result.SubthemesAdded = subthemesAdded
			u.logger.Info("Subthemes import completed", zap.Int("added", subthemesAdded))
			if manifest.Counts.Subthemes > 0 && subthemesAdded != manifest.Counts.Subthemes {
				u.logger.Warn("Subthemes count mismatch", zap.Int("expected", manifest.Counts.Subthemes), zap.Int("actual", subthemesAdded))
			}
		}
	}

	if contains(manifest.Includes, "tags") {
		tagsURL := baseURL + "tags.ndjson"
		expectedChecksum := manifest.Checksums["tags.ndjson"]
		u.logger.Info("Importing tags", zap.String("url", tagsURL))
		tagsAdded, err := u.importEntitiesWithChecksum(tagsURL, "tags", expectedChecksum)
		if err != nil {
			errMsg := fmt.Sprintf("tags: %v", err)
			u.logger.Error("Tags import failed", zap.Error(err))
			result.Errors = append(result.Errors, errMsg)
		} else {
			result.TagsAdded = tagsAdded
			u.logger.Info("Tags import completed", zap.Int("added", tagsAdded))
			if manifest.Counts.Tags > 0 && tagsAdded != manifest.Counts.Tags {
				u.logger.Warn("Tags count mismatch", zap.Int("expected", manifest.Counts.Tags), zap.Int("actual", tagsAdded))
			}
		}
	}

	if contains(manifest.Includes, "questions") {
		questionsURL := baseURL + "questions.ndjson"
		expectedChecksum := manifest.Checksums["questions.ndjson"]
		u.logger.Info("Importing questions", zap.String("url", questionsURL))
		added, updated, skipped, errors, err := u.adminQuestionRepo.ImportQuestions(questionsURL, job.ID, dataset.ID, expectedChecksum, u.importsRepo, u.logger)
		if err != nil {
			errMsg := fmt.Sprintf("questions import failed: %v", err)
			u.logger.Error("Questions import failed", zap.Error(err))
			result.Errors = append(result.Errors, errMsg)
		}
		result.QuestionsAdded = added
		result.QuestionsUpdated = updated
		result.QuestionsSkipped = skipped
		result.Errors = append(result.Errors, errors...)

		if manifest.Counts.Questions > 0 && added != manifest.Counts.Questions {
			u.logger.Warn("Questions count mismatch", zap.Int("expected", manifest.Counts.Questions), zap.Int("actual", added))
		}

		errorThreshold := manifest.Counts.Questions / 10
		if len(errors) > errorThreshold && errorThreshold > 0 {
			errMsg := fmt.Sprintf("question import: too many errors (%d errors, threshold: %d)", len(errors), errorThreshold)
			u.logger.Warn("Question import exceeded error threshold",
				zap.Int("error_count", len(errors)),
				zap.Int("error_threshold", errorThreshold),
				zap.Int("expected_questions", manifest.Counts.Questions),
			)
			result.Errors = append(result.Errors, errMsg)
		}
	}

	duration := time.Since(startTime)
	result.Success = len(result.Errors) == 0

	if result.Success {
		result.Message = fmt.Sprintf("Successfully imported %d questions from %s", result.QuestionsAdded, manifest.Dataset)
		u.logger.Info("Import completed successfully",
			zap.String("dataset", manifest.Dataset),
			zap.Int("questions_added", result.QuestionsAdded),
			zap.Int("questions_updated", result.QuestionsUpdated),
			zap.Int("questions_skipped", result.QuestionsSkipped),
			zap.Int("themes_added", result.ThemesAdded),
			zap.Int("subthemes_added", result.SubthemesAdded),
			zap.Int("tags_added", result.TagsAdded),
			zap.Duration("duration", duration),
		)
	} else {
		result.Message = fmt.Sprintf("Import completed with %d errors", len(result.Errors))
		u.logger.Warn("Import completed with errors",
			zap.Int("error_count", len(result.Errors)),
			zap.Int("questions_added", result.QuestionsAdded),
			zap.Int("questions_updated", result.QuestionsUpdated),
			zap.Int("questions_skipped", result.QuestionsSkipped),
			zap.Duration("duration", duration),
			zap.Strings("errors", result.Errors),
		)
	}

	finished := time.Now()
	jobUpdates := map[string]interface{}{
		"finished_at": finished,
		"success":     result.Success,
		"added":       result.QuestionsAdded,
		"updated":     result.QuestionsUpdated,
		"skipped":     result.QuestionsSkipped,
		"errors":      len(result.Errors),
		keyMessage:     result.Message,
	}
	if err := u.importsRepo.UpdateImportJobStatus(job.ID, jobUpdates); err != nil {
		u.logger.Error("Failed to update import job summary", zap.Error(err))
		return nil, fmt.Errorf("failed to update import job: %w", err)
	}

	if err := u.questionDatasetRepo.UpdateDatasetStatistics(dataset.ID, result.QuestionsAdded, result.ThemesAdded+result.SubthemesAdded+result.TagsAdded); err != nil {
		u.logger.Error("Failed to update dataset statistics", zap.Error(err))
		return nil, fmt.Errorf("failed to update dataset statistics: %w", err)
	}

	return result, nil
}

func (u *AdminQuestionsUsecase) CreateQuestion(req model.QuestionCreateRequest) (*model.Question, error) {
	correctCount := 0
	for _, answer := range req.Answers {
		if answer.IsCorrect {
			correctCount++
		}
	}
	if correctCount != 1 {
		return nil, errors.New("exactly one answer must be correct")
	}

	if u.questionRepo.Exists(req.Slug) {
		return nil, errors.New("question with this slug already exists")
	}

	theme, err := u.adminQuestionRepo.FindOrCreateTheme(req.Theme.Slug)
	if err != nil {
		return nil, err
	}

	var subthemes []model.Theme
	for _, t := range req.Subthemes {
		th, err := u.adminQuestionRepo.FindOrCreateTheme(t.Slug)
		if err != nil {
			return nil, err
		}
		subthemes = append(subthemes, *th)
	}

	var tags []model.Theme
	for _, t := range req.Tags {
		th, err := u.adminQuestionRepo.FindOrCreateTheme(t.Slug)
		if err != nil {
			return nil, err
		}
		tags = append(tags, *th)
	}

	i18nJSON, err := json.Marshal(req.I18n)
	if err != nil {
		return nil, err
	}

	answersJSON, err := json.Marshal(req.Answers)
	if err != nil {
		return nil, err
	}

	sourcesJSON, err := json.Marshal(req.Sources)
	if err != nil {
		return nil, err
	}

	question := &model.Question{
		Slug:             req.Slug,
		Version:          req.Version,
		Kind:             req.Kind,
		QType:            req.QType,
		Difficulty:       req.Difficulty,
		EstimatedSeconds: req.EstimatedSeconds,
		ShuffleAnswers:   req.ShuffleAnswers,
		ThemeID:          theme.ID,
		Theme:            *theme,
		Subthemes:        subthemes,
		Tags:             tags,
		I18n:             datatypes.JSON(i18nJSON),
		Answers:          datatypes.JSON(answersJSON),
		Sources:          datatypes.JSON(sourcesJSON),
		DatasetID:        req.DatasetID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err = u.adminQuestionRepo.Create(question)
	if err != nil {
		return nil, err
	}

	return question, nil
}

func (u *AdminQuestionsUsecase) GetQuestionByID(id uuid.UUID) (*model.Question, error) {
	return u.questionRepo.GetByID(id)
}

func (u *AdminQuestionsUsecase) GetQuestionBySlug(slug string, datasetID *uuid.UUID) (*model.Question, error) {
	if datasetID != nil {
		return u.questionRepo.GetBySlugAndDataset(slug, *datasetID)
	}
	return u.questionRepo.GetBySlug(slug)
}

func (u *AdminQuestionsUsecase) UpdateQuestion(id uuid.UUID, req model.QuestionUpdateRequest) (*model.Question, error) {
	question, err := u.questionRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.I18n != nil {
		validLanguages := map[string]bool{"fr": true, "en": true, "es": true}
		for lang := range *req.I18n {
			if !validLanguages[lang] {
				return nil, errors.New("invalid language: " + lang)
			}
		}
	}

	if req.Answers != nil {
		if len(*req.Answers) > 4 {
			return nil, errors.New("cannot have more than 4 answers")
		}
		correctCount := 0
		for _, answer := range *req.Answers {
			if answer.IsCorrect {
				correctCount++
			}
		}
		if correctCount != 1 {
			return nil, errors.New("exactly one answer must be correct")
		}
	}

	if req.Slug != nil {
		if *req.Slug != question.Slug && u.questionRepo.Exists(*req.Slug) {
			return nil, errors.New("question with this slug already exists")
		}
		question.Slug = *req.Slug
	}
	if req.Version != nil {
		question.Version = *req.Version
	}
	if req.Kind != nil {
		question.Kind = *req.Kind
	}
	if req.QType != nil {
		question.QType = *req.QType
	}
	if req.Difficulty != nil {
		question.Difficulty = *req.Difficulty
	}
	if req.EstimatedSeconds != nil {
		question.EstimatedSeconds = *req.EstimatedSeconds
	}
	if req.ShuffleAnswers != nil {
		question.ShuffleAnswers = *req.ShuffleAnswers
	}
	if req.Theme != nil {
		theme, err := u.adminQuestionRepo.FindOrCreateTheme(req.Theme.Slug)
		if err != nil {
			return nil, err
		}
		question.ThemeID = theme.ID
		question.Theme = *theme
	}
	if req.Subthemes != nil {
		var subthemes []model.Theme
		for _, t := range *req.Subthemes {
			th, err := u.adminQuestionRepo.FindOrCreateTheme(t.Slug)
			if err != nil {
				return nil, err
			}
			subthemes = append(subthemes, *th)
		}
		question.Subthemes = subthemes
	}
	if req.Tags != nil {
		var tags []model.Theme
		for _, t := range *req.Tags {
			th, err := u.adminQuestionRepo.FindOrCreateTheme(t.Slug)
			if err != nil {
				return nil, err
			}
			tags = append(tags, *th)
		}
		question.Tags = tags
	}
	if req.I18n != nil {
		i18nJSON, err := json.Marshal(*req.I18n)
		if err != nil {
			return nil, err
		}
		question.I18n = datatypes.JSON(i18nJSON)
	}
	if req.Answers != nil {
		answersJSON, err := json.Marshal(*req.Answers)
		if err != nil {
			return nil, err
		}
		question.Answers = datatypes.JSON(answersJSON)
	}
	if req.Sources != nil {
		sourcesJSON, err := json.Marshal(*req.Sources)
		if err != nil {
			return nil, err
		}
		question.Sources = datatypes.JSON(sourcesJSON)
	}

	question.UpdatedAt = time.Now()

	err = u.adminQuestionRepo.Update(question)
	if err != nil {
		return nil, err
	}

	return question, nil
}

func (u *AdminQuestionsUsecase) DeleteQuestion(id uuid.UUID) error {
	if err := u.adminQuestionRepo.DeleteQuestionSubthemes(id); err != nil {
		return err
	}
	if err := u.adminQuestionRepo.DeleteQuestionTags(id); err != nil {
		return err
	}
	return u.adminQuestionRepo.Delete(id)
}

func (u *AdminQuestionsUsecase) ListQuestionsByDataset(datasetID *uuid.UUID, limit, offset int) ([]*model.Question, int64, error) {
	filters := model.QuestionFilters{DatasetID: datasetID}
	return u.ListQuestionsWithFilters(filters, limit, offset)
}

func (u *AdminQuestionsUsecase) ListQuestionsWithFilters(filters model.QuestionFilters, limit, offset int) ([]*model.Question, int64, error) {
	return u.questionRepo.ListQuestionsWithFilters(filters, limit, offset)
}

func (u *AdminQuestionsUsecase) SearchQuestions(searchQuery string, datasetID *uuid.UUID, limit, offset int) ([]*model.Question, int64, error) {
	return u.questionRepo.SearchQuestions(searchQuery, datasetID, limit, offset)
}

func (u *AdminQuestionsUsecase) GetTotalQuestions() (int64, error) {
	return u.questionRepo.Count()
}

func (u *AdminQuestionsUsecase) BackupQuestions(datasetID *uuid.UUID) ([]model.Question, error) {
	return u.questionRepo.GetAllByDataset(datasetID)
}

func (u *AdminQuestionsUsecase) ExportQuestionsClean() ([]model.Question, error) {
	return u.questionRepo.GetAll()
}

func (u *AdminQuestionsUsecase) importEntitiesWithChecksum(url string, entityType string, expectedChecksum string) (int, error) {
	resp := httputil.FetchURLWithChecksum(url, expectedChecksum)
	if resp.Error != nil {
		if cm, ok := httputil.IsChecksumMismatch(resp.Error); ok {
			u.logger.Error("Entities file checksum verification failed",
				zap.String("url", url),
				zap.String("type", entityType),
				zap.String("expected", cm.Expected),
				zap.String("actual", cm.Actual))
			return 0, fmt.Errorf("checksum verification failed for %s file: the file may have been modified or corrupted (expected: %s, got: %s)", entityType, cm.Expected, cm.Actual)
		}
		u.logger.Error("Failed to fetch entities file",
			zap.String("url", url),
			zap.String("type", entityType),
			zap.Error(resp.Error))
		return 0, resp.Error
	}

	added := 0
	skipped := 0
	scanner := bufio.NewScanner(strings.NewReader(resp.Body))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entity model.Theme
		if err := json.Unmarshal([]byte(line), &entity); err != nil {
			u.logger.Warn("Failed to parse entity line",
				zap.String("type", entityType),
				zap.Int("line", lineNum),
				zap.Error(err),
			)
			continue
		}

		existing, err := u.adminQuestionRepo.FindOrCreateTheme(entity.Slug)
		if err != nil {
			u.logger.Warn("Failed to find or create entity",
				zap.String("type", entityType),
				zap.String(keySlug, entity.Slug),
				zap.Error(err),
			)
			continue
		}

		if existing != nil && len(entity.I18n) > 0 {
			if err := u.adminQuestionRepo.UpdateTheme(entity.Slug, entity.I18n); err != nil {
				u.logger.Warn("Failed to update theme i18n",
					zap.String("type", entityType),
					zap.String(keySlug, entity.Slug),
					zap.Error(err),
				)
			}
		}

		added++
	}

	if err := scanner.Err(); err != nil {
		u.logger.Error("Error reading NDJSON",
			zap.String("type", entityType),
			zap.Error(err))
		return added, fmt.Errorf("error reading NDJSON: %w", err)
	}

	u.logger.Info("Entity import summary",
		zap.String("type", entityType),
		zap.Int("added", added),
		zap.Int("skipped", skipped),
		zap.Int("total_lines", lineNum),
	)
	return added, nil
}
