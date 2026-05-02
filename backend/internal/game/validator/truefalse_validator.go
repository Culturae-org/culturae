// backend/internal/game/validator/truefalse_validator.go

package validator

import (
	"encoding/json"

	"github.com/Culturae-org/culturae/internal/game/types"
	"github.com/Culturae-org/culturae/internal/model"
)

type TrueFalseValidator struct{}

func (v *TrueFalseValidator) Validate(answer types.Answer, question *model.Question) ValidationResult {
	userAnswer, ok := answer.AnswerData.(bool)
	if !ok {
		return ValidationResult{
			IsCorrect: false,
			Score:     0,
			Feedback: map[string]interface{}{
				keyError: "invalid answer type for true/false question",
			},
		}
	}

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

	var correctAnswer bool
	for _, ans := range answers {
		if ans.IsCorrect {
			correctAnswer = ans.Slug == "true"
			break
		}
	}

	isCorrect := userAnswer == correctAnswer

	return ValidationResult{
		IsCorrect: isCorrect,
		Score: func() int {
			if isCorrect {
				return 100
			}
			return 0
		}(),
		Feedback: map[string]interface{}{
			keyUserAnswer:    userAnswer,
			"correct_answer": correctAnswer,
		},
	}
}
