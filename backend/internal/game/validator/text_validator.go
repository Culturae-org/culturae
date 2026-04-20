// backend/internal/game/validator/text_validator.go

package validator

import (
	"encoding/json"
	"strings"

	"github.com/Culturae-org/culturae/internal/game/types"
	"github.com/Culturae-org/culturae/internal/model"
)

type TextInputValidator struct {
	TypoTolerance int
	CaseSensitive bool
}

func (v *TextInputValidator) Validate(answer types.Answer, question *model.Question) ValidationResult {
	userAnswer, ok := answer.AnswerData.(string)
	if !ok {
		return ValidationResult{
			IsCorrect: false,
			Score:     0,
			Feedback: map[string]interface{}{
				"error": "invalid answer type for text question",
			},
		}
	}

	acceptedAnswers := v.getAcceptedAnswers(question)
	if len(acceptedAnswers) == 0 {
		return ValidationResult{
			IsCorrect: false,
			Score:     0,
			Feedback: map[string]interface{}{
				"error": "no accepted answers configured",
			},
		}
	}

	normalizedUser := normalizeString(userAnswer)

	for _, accepted := range acceptedAnswers {
		normalizedAccepted := normalizeString(accepted)

		if normalizedUser == normalizedAccepted {
			return ValidationResult{
				IsCorrect: true,
				Score:     100,
				Feedback: map[string]interface{}{
					"match_type":  "exact",
					"user_answer": userAnswer,
				},
			}
		}
	}

	if v.TypoTolerance > 0 && len(normalizedUser) >= 2 {
		for _, accepted := range acceptedAnswers {
			normalizedAccepted := normalizeString(accepted)

			if len(normalizedAccepted) <= 3 {
				continue
			}

			if len(normalizedUser) < len(normalizedAccepted)/2 {
				continue
			}

			distance := levenshteinDistance(normalizedUser, normalizedAccepted)
			if distance <= v.TypoTolerance {
				similarity := 1.0 - float64(distance)/float64(len(normalizedAccepted))
				score := int(similarity * 100)

				return ValidationResult{
					IsCorrect: true,
					Score:     score,
					Feedback: map[string]interface{}{
						"match_type":  "fuzzy",
						"similarity":  similarity,
						"typos":       distance,
						"user_answer": userAnswer,
					},
				}
			}
		}
	}

	correctHint := ""
	if len(acceptedAnswers) > 0 {
		correctHint = acceptedAnswers[0]
	}

	return ValidationResult{
		IsCorrect: false,
		Score:     0,
		Feedback: map[string]interface{}{
			"user_answer":    userAnswer,
			"correct_answer": correctHint,
		},
	}
}

func (v *TextInputValidator) getAcceptedAnswers(question *model.Question) []string {
	var answers []model.Answer
	if err := json.Unmarshal(question.Answers, &answers); err != nil {
		return []string{}
	}

	var accepted []string
	for _, ans := range answers {
		if ans.IsCorrect {
			for _, i18n := range ans.I18n {
				if i18n.Label != "" {
					accepted = append(accepted, i18n.Label)
				}
			}
		}
	}

	return accepted
}

func normalizeString(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	
	accentMap := map[string]string{
		"à": "a", "â": "a", "ä": "a", "á": "a", "ã": "a",
		"è": "e", "ê": "e", "ë": "e", "é": "e",
		"ì": "i", "î": "i", "ï": "i",
		"ò": "o", "ô": "o", "ö": "o", "ó": "o", "õ": "o",
		"ù": "u", "û": "u", "ü": "u", "ú": "u",
		"ñ": "n", "ç": "c", "ß": "ss", "œ": "oe", "æ": "ae",
		"À": "a", "Â": "a", "Ä": "a", "Á": "a", "Ã": "a",
		"È": "e", "Ê": "e", "Ë": "e", "É": "e",
		"Ì": "i", "Î": "i", "Ï": "i",
		"Ò": "o", "Ô": "o", "Ö": "o", "Ó": "o", "Õ": "o",
		"Ù": "u", "Û": "u", "Ü": "u", "Ú": "u",
		"Ñ": "n", "Ç": "c", "Œ": "oe", "Æ": "ae",
	}

	var result []rune
	for _, r := range s {
		if replacement, ok := accentMap[string(r)]; ok {
			result = append(result, []rune(replacement)...)
		} else if r == '-' || r == ' ' {
			continue
		} else {
			result = append(result, r)
		}
	}
	
	return string(result)
}

func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,
				matrix[i][j-1]+1,
				matrix[i-1][j-1]+cost,
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
