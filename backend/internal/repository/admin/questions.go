// backend/internal/repository/admin/questions.go

package admin

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AdminQuestionRepositoryInterface interface {
	Create(question *model.Question) error
	Update(question *model.Question) error
	Delete(id uuid.UUID) error
	FindOrCreateTheme(slug string) (*model.Theme, error)
	FindOrCreateThemeWithTransaction(slug string) (*model.Theme, error)
	UpdateTheme(slug string, i18n datatypes.JSON) error
	DeleteQuestionSubthemes(questionID uuid.UUID) error
	DeleteQuestionTags(questionID uuid.UUID) error
	ImportQuestions(url string, jobID uuid.UUID, datasetID uuid.UUID, expectedChecksum string, importsRepo ImportJobRepositoryInterface, logger *zap.Logger) (int, int, int, []string, error)
}

type AdminQuestionRepository struct {
	DB *gorm.DB
}

func NewAdminQuestionRepository(db *gorm.DB) *AdminQuestionRepository {
	return &AdminQuestionRepository{
		DB: db,
	}
}

func (r *AdminQuestionRepository) Create(question *model.Question) error {
	return r.DB.Create(question).Error
}

func (r *AdminQuestionRepository) Update(question *model.Question) error {
	return r.DB.Save(question).Error
}

func (r *AdminQuestionRepository) Delete(id uuid.UUID) error {
	return r.DB.Delete(&model.Question{}, id).Error
}

func (r *AdminQuestionRepository) FindOrCreateTheme(slug string) (*model.Theme, error) {
	var theme model.Theme
	theme.Slug = slug
	if err := r.DB.Where("slug = ?", slug).FirstOrCreate(&theme).Error; err != nil {
		return nil, err
	}
	return &theme, nil
}

func (r *AdminQuestionRepository) FindOrCreateThemeWithTransaction(slug string) (*model.Theme, error) {
	tx := r.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var theme model.Theme
	theme.Slug = slug
	if err := tx.Where("slug = ?", slug).FirstOrCreate(&theme).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return &theme, nil
}

func (r *AdminQuestionRepository) UpdateTheme(slug string, i18n datatypes.JSON) error {
	return r.DB.Model(&model.Theme{}).Where("slug = ?", slug).Update("i18n", i18n).Error
}

func (r *AdminQuestionRepository) DeleteQuestionSubthemes(questionID uuid.UUID) error {
	return r.DB.Exec("DELETE FROM question_subthemes WHERE question_id = ?", questionID).Error
}

func (r *AdminQuestionRepository) DeleteQuestionTags(questionID uuid.UUID) error {
	return r.DB.Exec("DELETE FROM question_tags WHERE question_id = ?", questionID).Error
}

func (r *AdminQuestionRepository) ImportQuestions(url string, jobID uuid.UUID, datasetID uuid.UUID, expectedChecksum string, importsRepo ImportJobRepositoryInterface, logger *zap.Logger) (int, int, int, []string, error) {
	resp := httputil.FetchURLWithChecksum(url, expectedChecksum)
	if resp.Error != nil {
		if cm, ok := httputil.IsChecksumMismatch(resp.Error); ok {
			logger.Error("Questions file checksum verification failed",
				zap.String("url", url),
				zap.String("expected", cm.Expected),
				zap.String("actual", cm.Actual))
			return 0, 0, 0, []string{fmt.Sprintf("checksum verification failed for questions file: the file may have been modified or corrupted (expected: %s, got: %s)", cm.Expected, cm.Actual)}, resp.Error
		}
		logger.Error("Failed to fetch questions file", zap.String("url", url), zap.Error(resp.Error))
		return 0, 0, 0, []string{resp.Error.Error()}, resp.Error
	}

	tx := r.DB.Begin()
	if tx.Error != nil {
		return 0, 0, 0, []string{tx.Error.Error()}, tx.Error
	}

	added := 0
	updated := 0
	skipped := 0
	errs := []string{}

	scanner := bufio.NewScanner(strings.NewReader(resp.Body))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		type TempQuestion struct {
			Kind             string                        `json:"kind"`
			Version          string                        `json:"version"`
			Slug             string                        `json:"slug"`
			Theme            model.Theme                   `json:"theme"`
			Subthemes        []model.Theme                 `json:"subthemes"`
			Tags             []model.Theme                 `json:"tags"`
			QType            string                        `json:"qtype"`
			Difficulty       string                        `json:"difficulty"`
			EstimatedSeconds int                           `json:"estimated_seconds"`
			Points           float64                       `json:"points"`
			ShuffleAnswers   bool                          `json:"shuffle_answers"`
			I18n             map[string]model.QuestionI18n `json:"i18n"`
			Answers          []model.Answer                `json:"answers"`
			Sources          []string                      `json:"sources"`
		}

		var temp TempQuestion
		if err := json.Unmarshal([]byte(line), &temp); err != nil {
			errMsg := fmt.Sprintf("line %d: failed to parse: %v", lineNum, err)
			errs = append(errs, errMsg)
			logger.Warn("Failed to parse question line",
				zap.Int("line", lineNum),
				zap.Error(err),
			)
			_ = importsRepo.SaveImportLog(&model.ImportQuestionLog{JobID: jobID, Line: lineNum, Slug: "", Action: "error", Message: err.Error()})
			continue
		}

		if temp.QType == "true_false" {
			temp.QType = "true-false"
		}

		i18nJSON, err := json.Marshal(temp.I18n)
		if err != nil {
			errMsg := fmt.Sprintf("%s: failed to marshal i18n: %v", temp.Slug, err)
			errs = append(errs, errMsg)
			logger.Error("Failed to marshal i18n",
				zap.String("slug", temp.Slug),
				zap.Error(err),
			)
			_ = importsRepo.SaveImportLog(&model.ImportQuestionLog{JobID: jobID, Line: lineNum, Slug: temp.Slug, Action: "error", Message: err.Error()})
			skipped++
			continue
		}

		answersJSON, err := json.Marshal(temp.Answers)
		if err != nil {
			errMsg := fmt.Sprintf("%s: failed to marshal answers: %v", temp.Slug, err)
			errs = append(errs, errMsg)
			logger.Error("Failed to marshal answers",
				zap.String("slug", temp.Slug),
				zap.Error(err),
			)
			_ = importsRepo.SaveImportLog(&model.ImportQuestionLog{JobID: jobID, Line: lineNum, Slug: temp.Slug, Action: "error", Message: err.Error()})
			skipped++
			continue
		}

		sourcesJSON, err := json.Marshal(temp.Sources)
		if err != nil {
			errMsg := fmt.Sprintf("%s: failed to marshal sources: %v", temp.Slug, err)
			errs = append(errs, errMsg)
			logger.Error("Failed to marshal sources",
				zap.String("slug", temp.Slug),
				zap.Error(err),
			)
			_ = importsRepo.SaveImportLog(&model.ImportQuestionLog{JobID: jobID, Line: lineNum, Slug: temp.Slug, Action: "error", Message: err.Error()})
			skipped++
			continue
		}

		question := model.Question{
			Kind:             temp.Kind,
			Version:          temp.Version,
			Slug:             temp.Slug,
			QType:            temp.QType,
			Difficulty:       temp.Difficulty,
			EstimatedSeconds: temp.EstimatedSeconds,
			ShuffleAnswers:   temp.ShuffleAnswers,
			I18n:             datatypes.JSON(i18nJSON),
			Answers:          datatypes.JSON(answersJSON),
			Sources:          datatypes.JSON(sourcesJSON),
			DatasetID:        &datasetID,
		}

		themeObj, err := r.findOrCreateThemeWithTx(tx, temp.Theme.Slug, logger)
		if err != nil {
			errMsg := fmt.Sprintf("%s: failed to create theme: %v", temp.Slug, err)
			errs = append(errs, errMsg)
			logger.Error("Failed to get/create theme",
				zap.String("slug", temp.Slug),
				zap.String("theme", temp.Theme.Slug),
				zap.Error(err),
			)
			_ = importsRepo.SaveImportLog(&model.ImportQuestionLog{JobID: jobID, Line: lineNum, Slug: temp.Slug, Action: "error", Message: err.Error()})
			skipped++
			continue
		}
		question.ThemeID = themeObj.ID

		var subthemes []model.Theme
		for _, st := range temp.Subthemes {
			stObj, err := r.findOrCreateThemeWithTx(tx, st.Slug, logger)
			if err != nil {
				logger.Warn("Failed to create subtheme",
					zap.String("question", temp.Slug),
					zap.String("subtheme", st.Slug),
					zap.Error(err),
				)
				continue
			}
			subthemes = append(subthemes, *stObj)
		}
		question.Subthemes = subthemes

		var tags []model.Theme
		for _, tag := range temp.Tags {
			tagObj, err := r.findOrCreateThemeWithTx(tx, tag.Slug, logger)
			if err != nil {
				logger.Warn("Failed to create tag",
					zap.String("question", temp.Slug),
					zap.String("tag", tag.Slug),
					zap.Error(err),
				)
				continue
			}
			tags = append(tags, *tagObj)
		}
		question.Tags = tags

		var existingQuestion model.Question
		err = tx.Where("slug = ? AND dataset_id = ?", question.Slug, datasetID).First(&existingQuestion).Error

		switch {
		case err == nil:
			if existingQuestion.Version != question.Version {
				existingQuestion.Kind = question.Kind
				existingQuestion.Version = question.Version
				existingQuestion.QType = question.QType
				existingQuestion.Difficulty = question.Difficulty
				existingQuestion.EstimatedSeconds = question.EstimatedSeconds
				existingQuestion.ShuffleAnswers = question.ShuffleAnswers
				existingQuestion.I18n = question.I18n
				existingQuestion.Answers = question.Answers
				existingQuestion.Sources = question.Sources
				existingQuestion.ThemeID = question.ThemeID
				existingQuestion.DatasetID = &datasetID
				if err := tx.Save(&existingQuestion).Error; err != nil {
					errMsg := fmt.Sprintf("%s: failed to update: %v", question.Slug, err)
					errs = append(errs, errMsg)
					logger.Error("Failed to update question", zap.String("slug", question.Slug), zap.Error(err))
					_ = importsRepo.SaveImportLog(&model.ImportQuestionLog{JobID: jobID, Line: lineNum, Slug: question.Slug, Action: "error", Message: err.Error()})
					skipped++
				} else {
					if err := tx.Model(&existingQuestion).Association("Subthemes").Replace(question.Subthemes); err != nil {
						logger.Warn("Failed to update subthemes", zap.String("slug", question.Slug), zap.Error(err))
					}
					if err := tx.Model(&existingQuestion).Association("Tags").Replace(question.Tags); err != nil {
						logger.Warn("Failed to update tags", zap.String("slug", question.Slug), zap.Error(err))
					}
					logger.Debug("Question updated", zap.String("slug", question.Slug), zap.String("old_version", existingQuestion.Version), zap.String("new_version", question.Version))
					_ = importsRepo.SaveImportLog(&model.ImportQuestionLog{JobID: jobID, Line: lineNum, Slug: question.Slug, Action: "updated", Message: fmt.Sprintf("version %s -> %s", existingQuestion.Version, question.Version)})
					updated++
				}
			} else {
				logger.Debug("Question unchanged, skipping", zap.String("slug", question.Slug), zap.String("version", question.Version))
				_ = importsRepo.SaveImportLog(&model.ImportQuestionLog{JobID: jobID, Line: lineNum, Slug: question.Slug, Action: "skipped", Message: "unchanged"})
				skipped++
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			if err := tx.Create(&question).Error; err != nil {
				errMsg := fmt.Sprintf("%s: failed to create: %v", question.Slug, err)
				errs = append(errs, errMsg)
				logger.Error("Failed to create question", zap.String("slug", question.Slug), zap.Error(err))
				_ = importsRepo.SaveImportLog(&model.ImportQuestionLog{JobID: jobID, Line: lineNum, Slug: question.Slug, Action: "error", Message: err.Error()})
				skipped++
			} else {
				logger.Debug("Question created", zap.String("slug", question.Slug), zap.String("theme", temp.Theme.Slug), zap.Int("subthemes", len(question.Subthemes)), zap.Int("tags", len(question.Tags)))
				_ = importsRepo.SaveImportLog(&model.ImportQuestionLog{JobID: jobID, Line: lineNum, Slug: question.Slug, Action: "created", Message: "new question"})
				added++
			}
		default:
			errMsg := fmt.Sprintf("%s: database error: %v", question.Slug, err)
			errs = append(errs, errMsg)
			logger.Error("Database error", zap.String("slug", question.Slug), zap.Error(err))
			_ = importsRepo.SaveImportLog(&model.ImportQuestionLog{JobID: jobID, Line: lineNum, Slug: question.Slug, Action: "error", Message: err.Error()})
			skipped++
		}
	}

	if err := scanner.Err(); err != nil {
		errMsg := fmt.Sprintf("error reading NDJSON: %v", err)
		errs = append(errs, errMsg)
		logger.Error("NDJSON scanner error", zap.Error(err))
	}

	logger.Info("Questions import summary",
		zap.Int("added", added),
		zap.Int("updated", updated),
		zap.Int("skipped", skipped),
		zap.Int("errors", len(errs)),
		zap.Int("total_lines", lineNum),
	)

	if len(errs) > 0 {
		tx.Rollback()
		return added, updated, skipped, errs, fmt.Errorf("import completed with errors")
	}

	if err := tx.Commit().Error; err != nil {
		return added, updated, skipped, errs, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return added, updated, skipped, errs, nil
}

func (r *AdminQuestionRepository) findOrCreateThemeWithTx(tx *gorm.DB, slug string, logger *zap.Logger) (*model.Theme, error) {
	var theme model.Theme
	result := tx.Where("slug = ?", slug).First(&theme)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		theme = model.Theme{
			ID:   uuid.New(),
			Slug: slug,
		}
		if err := tx.Create(&theme).Error; err != nil {
			return nil, err
		}
		logger.Debug("Theme auto-created", zap.String("slug", slug))
	} else if result.Error != nil {
		return nil, result.Error
	}

	return &theme, nil
}
