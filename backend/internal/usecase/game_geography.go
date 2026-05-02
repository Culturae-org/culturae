// backend/internal/usecase/game_geography.go

package usecase

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/Culturae-org/culturae/internal/model"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const (
	modeFlagToName    = "flag_to_name"
	modeNameToFlag    = "name_to_flag"
	modeCapitalToFlag = "capital_to_flag"
	modeCapitalToName = "capital_to_name"
)

// -----------------------------------------------
// Game Geography Usecase Methods
//
// - generateGeographyQuestions
//
// -----------------------------------------------

func (u *GameUsecase) generateGeographyQuestions(datasetID uuid.UUID, count int, continent string, includeTerritories bool, flagVariant ...string) ([]*model.GameQuestion, []*model.Question, error) {
	independentOnly := !includeTerritories

	countries, err := u.geographyRepo.ListRandomCountriesFiltered(datasetID, count, independentOnly, continent)
	if err != nil {
		return nil, nil, err
	}

	if len(countries) == 0 {
		return nil, nil, fmt.Errorf("no countries found in dataset - Please import dataset")
	}

	questions := make([]*model.GameQuestion, 0, count)
	virtualQuestions := make([]*model.Question, 0, count)

	variant := ""
	if len(flagVariant) > 0 {
		variant = flagVariant[0]
	}

	allFlagVariants := []string{
		model.FlagVariantFlagToName4,
		model.FlagVariantFlagToName2,
		model.FlagVariantNameToFlag4,
		model.FlagVariantNameToFlag2,
		model.FlagVariantFlagToCapital,
		model.FlagVariantCapitalToFlag4,
		model.FlagVariantCapitalToFlag2,
		model.FlagVariantCapitalToName4,
		model.FlagVariantCapitalToName2,
	}

	for i := range count {
		country := countries[i%len(countries)]

		var mode string
		activeVariant := variant

		switch variant {
		case "name_to_flag":
			variant = model.FlagVariantNameToFlag4
			activeVariant = variant
		case "flag_to_name":
			variant = model.FlagVariantFlagToName4
			activeVariant = variant
		case "capital_to_flag":
			variant = model.FlagVariantCapitalToFlag4
			activeVariant = variant
		case "capital_to_name":
			variant = model.FlagVariantCapitalToName4
			activeVariant = variant
		}

		if variant != "" && model.ValidFlagVariants[variant] {
			if variant == model.FlagVariantMix {
				activeVariant = allFlagVariants[rand.Intn(len(allFlagVariants))]
			}
			mode = flagVariantToMode(activeVariant)
		} else {
			mode = model.QTypeTextInput
			activeVariant = model.FlagVariantFlagToText
		}

		slug := fmt.Sprintf("geo-%s-%s-%s", country.Slug, activeVariant, datasetID.String()[:8])

		var questionID *uuid.UUID

		existingQ, err := u.questionRepo.FindGeographyQuestion(slug)
		if err == nil && existingQ != nil {
			questionID = &existingQ.ID
		} else {
			theme, _ := u.adminQuestionRepo.FindOrCreateTheme(categoryGeography)
			newQ := &model.Question{
				Kind:             categoryGeography,
				Version:          "1.0",
				Slug:             slug,
				QType:            mode,
				Difficulty:       "medium",
				EstimatedSeconds: 20,
				ShuffleAnswers:   true,
				ThemeID:          theme.ID,
				I18n:             datatypes.JSON("{}"),
				Answers:          datatypes.JSON("[]"),
				Data:             datatypes.JSON("{}"),
				CountryID:        &country.ID,
				CountrySlug:      country.Slug,
				Variant:          activeVariant,
				TimesPlayed:      0,
				TimesCorrect:     0,
				SuccessRate:      0,
			}
			if err := u.adminQuestionRepo.Create(newQ); err == nil {
				questionID = &newQ.ID
			}
		}

		gq := &model.GameQuestion{
			OrderNumber: i + 1,
			Type:        mode,
			QuestionID:  questionID,
			EntityKey:   fmt.Sprintf("geo:%s:%s", country.Slug, strings.TrimSuffix(strings.TrimSuffix(activeVariant, "_4"), "_2")),
		}

		vq := &model.Question{
			ID:               uuid.New(),
			Kind:             categoryGeography,
			EstimatedSeconds: 20,
			I18n:             datatypes.JSON{},
		}

		data := map[string]interface{}{}
		i18n := map[string]model.QuestionI18n{}

		var validatorAnswers []model.Answer

		switch mode {
		case model.QTypeTextInput:
			vq.QType = model.QTypeTextInput
			data[keyFlag] = getFlagCode(country)
			data["variant"] = activeVariant

			nameByLang := parseCountryNamesByLang(country.Name)

			for lang, name := range nameByLang {
				validatorAnswers = append(validatorAnswers, model.Answer{
					Slug:      country.Slug,
					IsCorrect: true,
					I18n: map[string]model.AnswerI18n{
						lang: {Label: name},
					},
				})
			}

			i18n["fr"] = model.QuestionI18n{Title: "Drapeau", Stem: "Quel est ce pays ?"}
			i18n["en"] = model.QuestionI18n{Title: "Flag", Stem: "What country is this?"}
			i18n["es"] = model.QuestionI18n{Title: "Bandera", Stem: "¿Qué país es este?"}

			data["correct_answer"] = map[string]interface{}{
				"slug": country.Slug,
				keyName: parseCountryNamesByLang(country.Name),
				keyFlag: getFlagCode(country),
			}

			if activeVariant == model.FlagVariantFlagToCapital {
				validatorAnswers = nil
				capitals := parseCapitalNames(country.Capital)

				for _, capName := range capitals {
					validatorAnswers = append(validatorAnswers, model.Answer{
						Slug:      country.Slug,
						IsCorrect: true,
						I18n: map[string]model.AnswerI18n{
							"fr": {Label: capName},
						},
					})
				}

				i18n["fr"] = model.QuestionI18n{Title: "Capitale", Stem: "Quelle est la capitale de ce pays (drapeau) ?"}
				i18n["en"] = model.QuestionI18n{Title: "Capital", Stem: "What is the capital of the country with this flag?"}
				i18n["es"] = model.QuestionI18n{Title: "Capital", Stem: "¿Cuál es la capital de este país (bandera)?"}

				var capitalNames map[string]string
				_ = json.Unmarshal(country.Capital, &capitalNames)
				data["correct_answer"] = map[string]interface{}{
					"slug": country.Slug,
					keyName: capitalNames,
					keyFlag: getFlagCode(country),
				}
			}

		case model.QTypeMCQ:
			vq.QType = model.QTypeMCQ

			numChoices := 4
			var subMode string

			switch activeVariant {
			case model.FlagVariantFlagToName2:
				numChoices = 2
				subMode = modeFlagToName
			case model.FlagVariantFlagToName4:
				numChoices = 4
				subMode = modeFlagToName
			case model.FlagVariantNameToFlag2:
				numChoices = 2
				subMode = modeNameToFlag
			case model.FlagVariantNameToFlag4:
				numChoices = 4
				subMode = modeNameToFlag
			case model.FlagVariantCapitalToFlag2:
				numChoices = 2
				subMode = modeCapitalToFlag
			case model.FlagVariantCapitalToFlag4:
				numChoices = 4
				subMode = modeCapitalToFlag
			case model.FlagVariantCapitalToName2:
				numChoices = 2
				subMode = modeCapitalToName
			case model.FlagVariantCapitalToName4:
				numChoices = 4
				subMode = modeCapitalToName
			default:
				if rand.Intn(2) == 0 {
					subMode = modeFlagToName
				} else {
					subMode = modeNameToFlag
				}
			}

			distractorCount := numChoices - 1
			distractors, _ := u.geographyRepo.ListRandomCountriesFiltered(
				datasetID,
				distractorCount+2,
				independentOnly,
				continent,
			)
			options := []map[string]interface{}{}

			options = append(options, map[string]interface{}{
				"slug": country.Slug,
				keyName: parseCountryNamesByLang(country.Name),
				keyFlag: getFlagCode(country),
			})
			validatorAnswers = append(validatorAnswers, model.Answer{
				Slug:      country.Slug,
				IsCorrect: true,
			})

			added := 0
			for _, d := range distractors {
				if d.ID == country.ID {
					continue
				}
				if added >= distractorCount {
					break
				}
				options = append(options, map[string]interface{}{
					"slug": d.Slug,
					keyName: parseCountryNamesByLang(d.Name),
					keyFlag: getFlagCode(d),
				})
				validatorAnswers = append(validatorAnswers, model.Answer{
					Slug:      d.Slug,
					IsCorrect: false,
				})
				added++
			}

			rand.Shuffle(len(options), func(ii, jj int) { options[ii], options[jj] = options[jj], options[ii] })

			data["options"] = options
			data["variant"] = activeVariant

			data["correct_answer"] = map[string]interface{}{
				"slug": country.Slug,
				keyName: parseCountryNamesByLang(country.Name),
				keyFlag: getFlagCode(country),
			}

			switch subMode {
			case modeFlagToName:
				data[keyFlag] = country.Flag
				data["mode"] = modeFlagToName
				i18n["fr"] = model.QuestionI18n{Title: labelQuiz, Stem: "A qui appartient ce drapeau ?"}
				i18n["en"] = model.QuestionI18n{Title: labelQuiz, Stem: "Whose flag is this?"}
			case modeNameToFlag:
				data["target_name"] = country.Name
				data["mode"] = modeNameToFlag
				i18n["fr"] = model.QuestionI18n{Title: labelQuiz, Stem: "Quel est le drapeau de : " + getCountryName(country.Name, "fr")}
				i18n["en"] = model.QuestionI18n{Title: labelQuiz, Stem: "Which flag belongs to: " + getCountryName(country.Name, "en")}
			case modeCapitalToFlag:
				data["target_capital"] = country.Capital
				data["mode"] = modeCapitalToFlag
				i18n["fr"] = model.QuestionI18n{Title: labelQuiz, Stem: "Quel est le drapeau du pays dont la capitale est : " + getCapitalName(country.Capital, "fr")}
				i18n["en"] = model.QuestionI18n{Title: labelQuiz, Stem: "Which flag belongs to the country with capital: " + getCapitalName(country.Capital, "en")}
			case modeCapitalToName:
				data["target_capital"] = country.Capital
				data["mode"] = modeCapitalToName
				i18n["fr"] = model.QuestionI18n{Title: labelQuiz, Stem: "Quel pays a pour capitale : " + getCapitalName(country.Capital, "fr")}
				i18n["en"] = model.QuestionI18n{Title: labelQuiz, Stem: "Which country has the capital: " + getCapitalName(country.Capital, "en")}
			}
		}

		jsonData, _ := json.Marshal(data)
		gq.Data = datatypes.JSON(jsonData)

		jsonI18n, _ := json.Marshal(i18n)
		vq.I18n = datatypes.JSON(jsonI18n)

		jsonAnswers, _ := json.Marshal(validatorAnswers)
		vq.Answers = datatypes.JSON(jsonAnswers)

		vqData := map[string]interface{}{}
		if data[keyFlag] != nil {
			vqData[keyFlag] = data[keyFlag]
		}
		if data["variant"] != nil {
			vqData["variant"] = data["variant"]
		}
		if data["mode"] != nil {
			vqData["mode"] = data["mode"]
		}
		if data["options"] != nil {
			vqData["options"] = data["options"]
		}
		if data["target_name"] != nil {
			vqData["target_name"] = data["target_name"]
		}
		if data["target_capital"] != nil {
			vqData["target_capital"] = data["target_capital"]
		}
		if data["correct_answer"] != nil {
			vqData["correct_answer"] = data["correct_answer"]
		}

		if len(vqData) > 0 {
			jsonVqData, _ := json.Marshal(vqData)
			vq.Data = datatypes.JSON(jsonVqData)
		}

		questions = append(questions, gq)
		virtualQuestions = append(virtualQuestions, vq)
	}

	return questions, virtualQuestions, nil
}

func getFlagCode(c model.Country) string {
	if c.ISOAlpha2 != "" {
		return strings.ToLower(c.ISOAlpha2)
	}
	return c.Flag
}

func flagVariantToMode(variant string) string {
	switch variant {
	case model.FlagVariantFlagToText, model.FlagVariantFlagToCapital:
		return model.QTypeTextInput
	case model.FlagVariantFlagToName2, model.FlagVariantFlagToName4,
		model.FlagVariantNameToFlag2, model.FlagVariantNameToFlag4,
		model.FlagVariantCapitalToFlag2, model.FlagVariantCapitalToFlag4,
		model.FlagVariantCapitalToName2, model.FlagVariantCapitalToName4:
		return model.QTypeMCQ
	default:
		return model.QTypeMCQ
	}
}

func getCountryName(nameJSON datatypes.JSON, lang string) string {
	var names map[string]interface{}
	_ = json.Unmarshal(nameJSON, &names)

	if val, ok := names[lang].(string); ok {
		return val
	}

	if nested, ok := names[lang].(map[string]interface{}); ok {
		if common, ok := nested["common"].(string); ok {
			return common
		}
	}

	if common, ok := names["common"].(string); ok {
		return common
	}
	return "Unknown"
}

func parseCountryNamesByLang(nameJSON datatypes.JSON) map[string]string {
	var namesMap map[string]interface{}
	_ = json.Unmarshal(nameJSON, &namesMap)
	result := make(map[string]string)

	for lang, v := range namesMap {
		if s, ok := v.(string); ok {
			result[lang] = s
		} else if nested, ok := v.(map[string]interface{}); ok {
			if common, ok := nested["common"].(string); ok {
				result[lang] = common
			}
		}
	}

	if _, ok := result["en"]; !ok {
		if common, ok := namesMap["common"].(string); ok {
			result["en"] = common
		}
	}

	return result
}

func getCapitalName(capitalJSON datatypes.JSON, lang string) string {
	var capitals map[string]string
	_ = json.Unmarshal(capitalJSON, &capitals)

	if val, ok := capitals[lang]; ok {
		return val
	}
	if val, ok := capitals["common"]; ok {
		return val
	}
	for _, v := range capitals {
		return v
	}
	return "Unknown Capital"
}

func parseCapitalNames(capitalJSON datatypes.JSON) []string {
	var capitals map[string]string
	_ = json.Unmarshal(capitalJSON, &capitals)

	var names []string
	for _, v := range capitals {
		names = append(names, v)
	}
	return names
}
