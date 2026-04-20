// backend/internal/game/question_dto.go

package game

import (
	"encoding/json"
	"math/rand"

	"github.com/Culturae-org/culturae/internal/game/types"
	"github.com/Culturae-org/culturae/internal/model"
)

type QuestionPayload struct {
	ID        string              `json:"id"`
	QType     string              `json:"qtype"`
	Variant   string              `json:"variant"`
	Stem      map[string]string   `json:"stem"`
	Points    int                 `json:"points"`
	TimeLimit int                 `json:"time_limit"`
	Data      QuestionDataPayload `json:"data"`
}

type QuestionDataPayload struct {

	Flag string `json:"flag,omitempty"`

	Mode string `json:"mode,omitempty"`

	Options []AnswerOptionPayload `json:"options,omitempty"`

	TargetName map[string]string `json:"target_name,omitempty"`

	TargetCapital map[string]string `json:"target_capital,omitempty"`
}

type AnswerOptionPayload struct {
	Slug string            `json:"slug"`
	Name map[string]string `json:"name,omitempty"`
	Flag string            `json:"flag,omitempty"`
}

type AnswerResultPayload struct {
	IsCorrect     bool              `json:"is_correct"`
	Points        int               `json:"points"`
	TimeSpentMs   int64             `json:"time_spent_ms"`
	CorrectAnswer CorrectAnswerInfo `json:"correct_answer"`
}

type CorrectAnswerInfo struct {
	Slug string            `json:"slug"`
	Name map[string]string `json:"name,omitempty"`
	Flag string            `json:"flag,omitempty"`
}

func QuestionToPayload(q *model.Question) QuestionPayload {
	var i18nMap map[string]model.QuestionI18n
	_ = json.Unmarshal(q.I18n, &i18nMap)

	stem := make(map[string]string, len(i18nMap))
	for lang, v := range i18nMap {
		stem[lang] = v.Stem
	}

	var rawData map[string]interface{}
	_ = json.Unmarshal(q.Data, &rawData)

	variant := ""
	if v, ok := rawData["variant"].(string); ok {
		variant = v
	}

	data := questionDataFromRaw(rawData)

	qtype := q.QType
	if isMCQQType(qtype) {
		qtype = model.QTypeMCQ
	}

	if isMCQQType(q.QType) && len(data.Options) == 0 {
		var answers []model.Answer
		if err := json.Unmarshal(q.Answers, &answers); err == nil && len(answers) > 0 {
			opts := make([]AnswerOptionPayload, 0, len(answers))
			for _, ans := range answers {
				opt := AnswerOptionPayload{Slug: ans.Slug}
				if len(ans.I18n) > 0 {
					opt.Name = make(map[string]string, len(ans.I18n))
					for lang, i18n := range ans.I18n {
						opt.Name[lang] = i18n.Label
					}
				}
				opts = append(opts, opt)
			}
			rand.Shuffle(len(opts), func(i, j int) { opts[i], opts[j] = opts[j], opts[i] })
			data.Options = opts
		}
	}

	return QuestionPayload{
		ID:        q.ID.String(),
		QType:     qtype,
		Variant:   variant,
		Stem:      stem,
		Points:    100,
		TimeLimit: q.EstimatedSeconds,
		Data:      data,
	}
}

func CorrectAnswerFromQuestion(q *model.Question) CorrectAnswerInfo {
	if q == nil {
		return CorrectAnswerInfo{}
	}

	if len(q.Data) > 0 {
		var rawData map[string]interface{}
		_ = json.Unmarshal(q.Data, &rawData)
		if ca, ok := rawData["correct_answer"].(map[string]interface{}); ok {
			info := CorrectAnswerInfo{}
			if slug, ok := ca["slug"].(string); ok {
				info.Slug = slug
			}
			if flag, ok := ca["flag"].(string); ok {
				info.Flag = flag
			}
			if name, ok := ca["name"].(map[string]interface{}); ok {
				info.Name = make(map[string]string)
				for lang, v := range name {
					if s, ok := v.(string); ok {
						info.Name[lang] = s
					}
				}
			}
			if info.Slug != "" {
				return info
			}
		}
	}

	var answers []model.Answer
	if err := json.Unmarshal(q.Answers, &answers); err == nil {
		for _, ans := range answers {
			if ans.IsCorrect {
				info := CorrectAnswerInfo{Slug: ans.Slug}
				if len(ans.I18n) > 0 {
					info.Name = make(map[string]string, len(ans.I18n))
					for lang, i18n := range ans.I18n {
						info.Name[lang] = i18n.Label
					}
				}
				return info
			}
		}
	}

	return CorrectAnswerInfo{}
}

func BuildAnswerResultPayload(answer types.Answer, question *model.Question) AnswerResultPayload {
	return AnswerResultPayload{
		IsCorrect:     answer.IsCorrect,
		Points:        answer.Points,
		TimeSpentMs:   answer.ServerTimeSpent.Milliseconds(),
		CorrectAnswer: CorrectAnswerFromQuestion(question),
	}
}

func isMCQQType(qtype string) bool {
	switch qtype {
	case model.QTypeMCQ, "single_choice":
		return true
	}
	return false
}

func questionDataFromRaw(rawData map[string]interface{}) QuestionDataPayload {
	data := QuestionDataPayload{}

	if flag, ok := rawData["flag"].(string); ok {
		data.Flag = flag
	}
	if mode, ok := rawData["mode"].(string); ok {
		data.Mode = mode
	}

	if opts, ok := rawData["options"].([]interface{}); ok {
		for _, opt := range opts {
			o, ok := opt.(map[string]interface{})
			if !ok {
				continue
			}
			option := AnswerOptionPayload{}
			if slug, ok := o["slug"].(string); ok {
				option.Slug = slug
			}
			if flag, ok := o["flag"].(string); ok {
				option.Flag = flag
			}
			if name, ok := o["name"].(map[string]interface{}); ok {
				option.Name = make(map[string]string)
				for lang, v := range name {
					if s, ok := v.(string); ok {
						option.Name[lang] = s
					}
				}
			}
			data.Options = append(data.Options, option)
		}
	}

	if tn, ok := rawData["target_name"].(map[string]interface{}); ok {
		data.TargetName = make(map[string]string)
		for lang, v := range tn {
			if s, ok := v.(string); ok {
				data.TargetName[lang] = s
			}
		}
	}
	if tc, ok := rawData["target_capital"].(map[string]interface{}); ok {
		data.TargetCapital = make(map[string]string)
		for lang, v := range tc {
			if s, ok := v.(string); ok {
				data.TargetCapital[lang] = s
			}
		}
	}
	return data
}
