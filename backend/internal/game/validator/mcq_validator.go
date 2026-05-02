// backend/internal/game/validator/mcq_validator.go

package validator

import (
	"encoding/json"

	"github.com/Culturae-org/culturae/internal/game/types"
	"github.com/Culturae-org/culturae/internal/model"
)

type MCQValidator struct{}

func (v *MCQValidator) Validate(answer types.Answer, question *model.Question) ValidationResult {
	var answers []model.Answer
	if err := json.Unmarshal(question.Answers, &answers); err != nil {
		return ValidationResult{
			IsCorrect: false,
			Score:     0,
			Feedback: map[string]interface{}{
				keyError: "failed to parse question answers",
			},
		}
	}

	var answerSlug string
	switch answer.AnswerType {
	case "mcq":
		if slug, ok := answer.AnswerData.(string); ok {
			answerSlug = slug
		}
	default:
		return ValidationResult{
			IsCorrect: false,
			Score:     0,
			Feedback: map[string]interface{}{
				keyError: "invalid answer type for MCQ question",
			},
		}
	}

	correctSlug := ""
	for _, ans := range answers {
		if ans.IsCorrect {
			correctSlug = ans.Slug
			break
		}
	}

	for _, ans := range answers {
		if ans.Slug == answerSlug && ans.IsCorrect {
			return ValidationResult{
				IsCorrect: true,
				Score:     100,
				Feedback: map[string]interface{}{
					"answer_slug":         answerSlug,
					"correct_answer_slug": correctSlug,
					keyMatchType:          "exact",
				},
			}
		}
	}

	return ValidationResult{
		IsCorrect: false,
		Score:     0,
		Feedback: map[string]interface{}{
			"answer_slug":         answerSlug,
			"correct_answer_slug": correctSlug,
		},
	}
}
