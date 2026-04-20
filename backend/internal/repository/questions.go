// backend/internal/repository/questions.go

package repository

import (
	"github.com/Culturae-org/culturae/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuestionRepositoryInterface interface {
	GetByID(id uuid.UUID) (*model.Question, error)
	GetBySlug(slug string) (*model.Question, error)
	GetBySlugAndDataset(slug string, datasetID uuid.UUID) (*model.Question, error)
	ListQuestionsByDataset(datasetID *uuid.UUID, limit, offset int) ([]*model.Question, int64, error)
	ListQuestionsWithFilters(filters model.QuestionFilters, limit, offset int) ([]*model.Question, int64, error)
	SearchQuestions(searchQuery string, datasetID *uuid.UUID, limit, offset int) ([]*model.Question, int64, error)
	Count() (int64, error)
	CountByDataset(datasetID uuid.UUID, qtype string) (int64, error)
	GetTotalCount() (int64, error)
	Exists(slug string) bool
	GetAll() ([]model.Question, error)
	GetAllByDataset(datasetID *uuid.UUID) ([]model.Question, error)
	GetQuestionCountByDifficulty(datasetID uuid.UUID) (map[string]int64, error)
	GetQuestionCountByTheme(datasetID uuid.UUID) ([]map[string]interface{}, error)
	ListRandomQuestionsByDataset(datasetID *uuid.UUID, count int, qtype string) ([]*model.Question, error)
	FindGeographyQuestion(slug string) (*model.Question, error)
	IncrementQuestionStats(questionID uuid.UUID, isCorrect bool, timeSpent int) error
}

type QuestionRepository struct {
	DB *gorm.DB
}

func NewQuestionRepository(
	db *gorm.DB,
) *QuestionRepository {
	return &QuestionRepository{
		DB: db,
	}
}

func (r *QuestionRepository) Create(question *model.Question) error {
	return r.DB.Create(question).Error
}

func (r *QuestionRepository) GetByID(id uuid.UUID) (*model.Question, error) {
	var question model.Question
	err := r.DB.Preload("Theme").Preload("Subthemes").Preload("Tags").First(&question, id).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (r *QuestionRepository) GetBySlug(slug string) (*model.Question, error) {
	var question model.Question
	err := r.DB.Preload("Theme").Preload("Subthemes").Preload("Tags").Where("slug = ?", slug).First(&question).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (r *QuestionRepository) GetBySlugAndDataset(slug string, datasetID uuid.UUID) (*model.Question, error) {
	var question model.Question
	err := r.DB.Preload("Theme").Preload("Subthemes").Preload("Tags").
		Where("slug = ? AND dataset_id = ?", slug, datasetID).
		First(&question).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (r *QuestionRepository) Update(question *model.Question) error {
	return r.DB.Save(question).Error
}

func (r *QuestionRepository) Delete(id uuid.UUID) error {
	return r.DB.Delete(&model.Question{}, id).Error
}

func (r *QuestionRepository) ListQuestionsByDataset(datasetID *uuid.UUID, limit, offset int) ([]*model.Question, int64, error) {
	var questions []*model.Question
	var total int64

	query := r.DB.Preload("Theme").Preload("Subthemes").Preload("Tags")

	if datasetID != nil {
		query = query.Where("dataset_id = ?", datasetID)
	}

	if err := query.Model(&model.Question{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(limit).Offset(offset).Find(&questions).Error
	return questions, total, err
}

func (r *QuestionRepository) ListQuestionsWithFilters(filters model.QuestionFilters, limit, offset int) ([]*model.Question, int64, error) {
	var questions []*model.Question
	var total int64

	query := r.DB.Preload("Theme").Preload("Subthemes").Preload("Tags")

	if filters.DatasetID != nil {
		query = query.Where("dataset_id = ?", filters.DatasetID)
	}

	if filters.Difficulty != nil {
		query = query.Where("difficulty = ?", *filters.Difficulty)
	}

	if filters.QType != nil {
		query = query.Where("q_type = ?", *filters.QType)
	}

	if filters.Theme != nil {
		query = query.Joins("JOIN themes ON questions.theme_id = themes.id").
			Where("themes.slug = ?", *filters.Theme)
	}

	if filters.Subtheme != nil {
		query = query.Joins("JOIN question_subthemes qs ON questions.id = qs.question_id").
			Joins("JOIN subthemes s ON qs.subtheme_id = s.id").
			Where("s.slug = ?", *filters.Subtheme)
	}

	if len(filters.Tags) > 0 {
		query = query.Joins("JOIN question_tags qt ON questions.id = qt.question_id").
			Joins("JOIN tags t ON qt.tag_id = t.id").
			Where("t.slug IN ?", filters.Tags)
	}

	if filters.SearchQuery != nil && *filters.SearchQuery != "" {
		searchPattern := "%" + *filters.SearchQuery + "%"
		query = query.Where(
			r.DB.Where("CAST(questions.id AS TEXT) ILIKE ?", searchPattern).
				Or("questions.slug ILIKE ?", searchPattern).
				Or("questions.question ILIKE ?", searchPattern).
				Or("questions.answer ILIKE ?", searchPattern),
		)
	}

	if err := query.Model(&model.Question{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&questions).Error
	return questions, total, err
}

func (r *QuestionRepository) SearchQuestions(searchQuery string, datasetID *uuid.UUID, limit, offset int) ([]*model.Question, int64, error) {
	var questions []*model.Question
	var total int64

	query := r.DB.Preload("Theme").Preload("Subthemes").Preload("Tags")

	if datasetID != nil {
		query = query.Where("dataset_id = ?", datasetID)
	}

	if searchQuery != "" {
		searchPattern := "%" + searchQuery + "%"
		query = query.Where(
			r.DB.Where("CAST(id AS TEXT) ILIKE ?", searchPattern).
				Or("slug ILIKE ?", searchPattern).
				Or("id IN (?)",
					r.DB.Table("questions").
						Select("questions.id").
						Joins("JOIN themes ON questions.theme_id = themes.id").
						Where("themes.slug ILIKE ?", searchPattern),
				),
		)
	}

	if err := query.Model(&model.Question{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&questions).Error
	return questions, total, err
}

func (r *QuestionRepository) Count() (int64, error) {
	var count int64
	err := r.DB.Model(&model.Question{}).Count(&count).Error
	return count, err
}

func (r *QuestionRepository) CountByDataset(datasetID uuid.UUID, qtype string) (int64, error) {
	var count int64
	q := r.DB.Model(&model.Question{}).Where("dataset_id = ?", datasetID)
	q = applyQTypeFilter(q, qtype)
	err := q.Count(&count).Error
	return count, err
}

func (r *QuestionRepository) GetTotalCount() (int64, error) {
	return r.Count()
}

func (r *QuestionRepository) Exists(slug string) bool {
	var count int64
	r.DB.Model(&model.Question{}).Where("slug = ?", slug).Count(&count)
	return count > 0
}

func (r *QuestionRepository) GetAll() ([]model.Question, error) {
	var questions []model.Question
	err := r.DB.Preload("Theme").Preload("Subthemes").Preload("Tags").Find(&questions).Error
	return questions, err
}

func (r *QuestionRepository) GetAllByDataset(datasetID *uuid.UUID) ([]model.Question, error) {
	var questions []model.Question
	query := r.DB.Preload("Theme").Preload("Subthemes").Preload("Tags")
	if datasetID != nil {
		query = query.Where("dataset_id = ?", datasetID)
	}
	err := query.Find(&questions).Error
	return questions, err
}

func (r *QuestionRepository) GetQuestionCountByDifficulty(datasetID uuid.UUID) (map[string]int64, error) {
	type Result struct {
		Difficulty string
		Count      int64
	}

	var results []Result
	err := r.DB.Model(&model.Question{}).
		Select("difficulty, COUNT(*) as count").
		Where("dataset_id = ?", datasetID).
		Group("difficulty").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Difficulty] = r.Count
	}

	return counts, nil
}

func (r *QuestionRepository) GetQuestionCountByTheme(datasetID uuid.UUID) ([]map[string]interface{}, error) {
	type Result struct {
		ThemeSlug string
		Count     int64
	}

	var results []Result
	err := r.DB.Table("questions").
		Select("themes.slug as theme_slug, COUNT(questions.id) as count").
		Joins("JOIN themes ON questions.theme_id = themes.id").
		Where("questions.dataset_id = ?", datasetID).
		Group("themes.slug").
		Order("count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var output []map[string]interface{}
	for _, r := range results {
		output = append(output, map[string]interface{}{
			"theme": r.ThemeSlug,
			"count": r.Count,
		})
	}

	return output, nil
}

func (r *QuestionRepository) ListRandomQuestionsByDataset(datasetID *uuid.UUID, count int, qtype string) ([]*model.Question, error) {
	var questions []*model.Question

	query := r.DB.Preload("Theme").Preload("Subthemes").Preload("Tags")

	if datasetID != nil {
		query = query.Where("dataset_id = ?", datasetID)
	}
	query = applyQTypeFilter(query, qtype)

	err := query.Order("RANDOM()").Limit(count).Find(&questions).Error
	return questions, err
}

func applyQTypeFilter(q *gorm.DB, qtype string) *gorm.DB {
	mcqTypes := []string{model.QTypeMCQ, "single_choice"}
	switch qtype {
	case "":
		return q
	case "mcq_2":
		return q.Where(
			"(q_type IN ? AND jsonb_array_length(answers) = 2) OR q_type = ?",
			mcqTypes, "true-false",
		)
	case "mcq_4":
		return q.Where("q_type IN ? AND jsonb_array_length(answers) = 4", mcqTypes)
	case "true_false":
		return q.Where("q_type = ?", "true-false")
	case "single_choice_2":
		return q.Where("q_type IN ?", mcqTypes)
	case "mcq_2_mix":
		return q.Where("q_type IN ? OR q_type = ?", mcqTypes, "true-false")
	case model.QTypeMCQ:
		return q.Where("q_type IN ? OR q_type = ?", mcqTypes, "true-false")
	default:
		return q.Where("q_type = ?", qtype)
	}
}

func (r *QuestionRepository) FindGeographyQuestion(slug string) (*model.Question, error) {
	var question model.Question
	err := r.DB.Where("slug = ?", slug).First(&question).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (r *QuestionRepository) IncrementQuestionStats(questionID uuid.UUID, isCorrect bool, timeSpent int) error {
	correctIncrement := 0
	if isCorrect {
		correctIncrement = 1
	}
	return r.DB.Model(&model.Question{}).Where("id = ?", questionID).Updates(map[string]interface{}{
		"times_played":  gorm.Expr("times_played + 1"),
		"times_correct": gorm.Expr("times_correct + ?", correctIncrement),
		"avg_time_ms":   gorm.Expr("COALESCE(avg_time_ms, 0) * times_played + ? / (times_played + 1)", timeSpent),
	}).Error
}
