// backend/internal/game/validator/validator.go

package validator

import (
	"github.com/Culturae-org/culturae/internal/game/types"
	"github.com/Culturae-org/culturae/internal/model"
)

type QuestionValidator interface {
	Validate(answer types.Answer, question *model.Question) ValidationResult
}

type ValidationResult struct {
	IsCorrect bool
	Score     int
	Feedback  map[string]interface{}
}

func NormalizeString(s string) string {
	return normalizeString(s)
}

func LevenshteinDistance(a, b string) int {
	return levenshteinDistance(a, b)
}

func GetValidator(questionType string) QuestionValidator {
	switch questionType {
	case model.QTypeMCQ, "single_choice":
		return &MCQValidator{}
	case "true_false", "true-false":
		return &TrueFalseValidator{}
	case model.QTypeTextInput:
		return &TextInputValidator{
			TypoTolerance: 2,
			CaseSensitive: false,
		}
default:
		return &MCQValidator{}
	}
}
