// backend/internal/usecase/admin/game_template.go

package admin

import (
	"fmt"
	"strings"

	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/repository"
	"github.com/Culturae-org/culturae/internal/service"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AdminGameTemplatesUsecase struct {
	templateRepo   repository.GameTemplateRepositoryInterface
	loggingService service.LoggingServiceInterface
	logger         *zap.Logger
}

func NewAdminGameTemplatesUsecase(
	templateRepo repository.GameTemplateRepositoryInterface,
	loggingService service.LoggingServiceInterface,
	logger *zap.Logger,
) *AdminGameTemplatesUsecase {
	return &AdminGameTemplatesUsecase{
		templateRepo:   templateRepo,
		loggingService: loggingService,
		logger:         logger,
	}
}

// -----------------------------------------------
// Admin Game Templates Usecase Methods
//
// - ListGameTemplates
// - CountGameTemplates
// - GetGameTemplateByID
// - CreateGameTemplate
// - UpdateGameTemplate
// - SeedDefaultGameTemplates
// - DeleteGameTemplate
//
// -----------------------------------------------

func (u *AdminGameTemplatesUsecase) ListGameTemplates(params model.GameTemplateListParams) ([]model.GameTemplate, error) {
	return u.templateRepo.List(params)
}

func (u *AdminGameTemplatesUsecase) CountGameTemplates(params model.GameTemplateListParams) (int64, error) {
	return u.templateRepo.Count(params)
}

func (u *AdminGameTemplatesUsecase) GetGameTemplateByID(id uuid.UUID) (*model.GameTemplate, error) {
	t, err := u.templateRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("game template not found: %w", err)
	}
	return t, nil
}

func (u *AdminGameTemplatesUsecase) CreateGameTemplate(req *model.CreateGameTemplateRequest, adminID uuid.UUID) (*model.GameTemplate, error) {
	slug := strings.ToLower(strings.TrimSpace(req.Slug))
	if slug == "" {
		return nil, fmt.Errorf("slug is required")
	}

	scoreMode := req.ScoreMode
	if scoreMode == "" {
		scoreMode = "time_bonus"
	}

	pointsPerCorrect := req.PointsPerCorrect
	if pointsPerCorrect == 0 {
		pointsPerCorrect = 100
	}

	t := &model.GameTemplate{
		Name:               strings.TrimSpace(req.Name),
		Description:        strings.TrimSpace(req.Description),
		NameI18n:           req.NameI18n,
		DescI18n:           req.DescI18n,
		Slug:               slug,
		Mode:               req.Mode,
		MinPlayers:         req.MinPlayers,
		MaxPlayers:         req.MaxPlayers,
		QuestionCount:      req.QuestionCount,
		QuestionType:       req.QuestionType,
		DatasetID:          req.DatasetID,
		Category:           req.Category,
		FlagVariant:        req.FlagVariant,
		Continent:          req.Continent,
		IncludeTerritories: req.IncludeTerritories,
		Language:           req.Language,
		PointsPerCorrect:   pointsPerCorrect,
		TimeBonus:          req.TimeBonus,
		ScoreMode:          scoreMode,
		IsActive:           req.IsActive,
	}

	if err := u.templateRepo.Create(t); err != nil {
		u.logger.Error("Failed to create game template", zap.Error(err))
		return nil, fmt.Errorf("failed to create game template: %w", err)
	}

	_ = u.loggingService.LogAdminAction(adminID, "", "create", "game_template", &t.ID, "", "", map[string]interface{}{
		"template_name": t.Name,
	}, true, nil)

	return t, nil
}

func (u *AdminGameTemplatesUsecase) UpdateGameTemplate(id uuid.UUID, req *model.UpdateGameTemplateRequest, adminID uuid.UUID) (*model.GameTemplate, error) {
	t, err := u.templateRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("game template not found: %w", err)
	}

	if req.Name != nil {
		t.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		t.Description = strings.TrimSpace(*req.Description)
	}
	if req.NameI18n != nil {
		t.NameI18n = *req.NameI18n
	}
	if req.DescI18n != nil {
		t.DescI18n = *req.DescI18n
	}
	if req.Slug != nil {
		t.Slug = strings.ToLower(strings.TrimSpace(*req.Slug))
	}
	if req.MinPlayers != nil {
		t.MinPlayers = *req.MinPlayers
	}
	if req.MaxPlayers != nil {
		t.MaxPlayers = *req.MaxPlayers
	}
	if req.QuestionCount != nil {
		t.QuestionCount = *req.QuestionCount
	}
	if req.DatasetID != nil {
		t.DatasetID = req.DatasetID
	}
	if req.Category != nil {
		t.Category = *req.Category
	}
	if req.FlagVariant != nil {
		t.FlagVariant = *req.FlagVariant
	}
	if req.QuestionType != nil {
		t.QuestionType = *req.QuestionType
	}
	if req.Continent != nil {
		t.Continent = *req.Continent
	}
	if req.IncludeTerritories != nil {
		t.IncludeTerritories = *req.IncludeTerritories
	}
	if req.Language != nil {
		t.Language = *req.Language
	}
	if req.PointsPerCorrect != nil {
		t.PointsPerCorrect = *req.PointsPerCorrect
	}
	if req.TimeBonus != nil {
		t.TimeBonus = *req.TimeBonus
	}
	if req.ScoreMode != nil {
		t.ScoreMode = *req.ScoreMode
	}
	if req.IsActive != nil {
		t.IsActive = *req.IsActive
	}

	if err := u.templateRepo.Update(t); err != nil {
		u.logger.Error("Failed to update game template", zap.String("id", id.String()), zap.Error(err))
		return nil, fmt.Errorf("failed to update game template: %w", err)
	}

	_ = u.loggingService.LogAdminAction(adminID, "", "update", "game_template", &t.ID, "", "", nil, true, nil)

	return t, nil
}

func (u *AdminGameTemplatesUsecase) SeedDefaultGameTemplates() (int, error) {
	type def struct {
		name         string
		slug         string
		mode         string
		category     string
		flagVariant  string
		questionType string
		minPlayers   int
		maxPlayers   int
		scoreMode    string
		timeBonus    bool
		qCount       int
		points       int
		nameI18n     model.I18nStrings
		descI18n     model.I18nStrings
	}

	modePrefix := map[string]model.I18nStrings{
		modeSolo: {"en": labelSolo, "fr": labelSolo, "es": labelSolo},
		label1v1:  {"en": label1v1, "fr": label1v1, "es": label1v1},
	}

	qtLabel := map[string]model.I18nStrings{
		qtMCQ4:      {"en": labelMCQ4, "fr": "QCM ×4", "es": "Test ×4"},
		qtMCQ2:      {"en": labelMCQ2, "fr": "QCM ×2", "es": "Test ×2"},
		qtTextInput: {"en": labelTextInput, "fr": "Texte libre", "es": "Texto libre"},
	}

	fvLabel := map[string]model.I18nStrings{
		fvMix:               {"en": labelMix, "fr": labelMix, "es": labelMix},
		fvFlagToName4:    {"en": "Flag → Country ×4", "fr": "Drapeau → Pays ×4", "es": "Bandera → País ×4"},
		fvFlagToName2:    {"en": "Flag → Country ×2", "fr": "Drapeau → Pays ×2", "es": "Bandera → País ×2"},
		fvNameToFlag4:    {"en": "Country → Flag ×4", "fr": "Pays → Drapeau ×4", "es": "País → Bandera ×4"},
		fvNameToFlag2:    {"en": "Country → Flag ×2", "fr": "Pays → Drapeau ×2", "es": "País → Bandera ×2"},
		slugCapitalToFlag4: {"en": "Capital → Flag ×4", "fr": "Capitale → Drapeau ×4", "es": "Capital → Bandera ×4"},
		slugCapitalToFlag2: {"en": "Capital → Flag ×2", "fr": "Capitale → Drapeau ×2", "es": "Capital → Bandera ×2"},
		slugCapitalToName4: {"en": "Capital → Country ×4", "fr": "Capitale → Pays ×4", "es": "Capital → País ×4"},
		slugCapitalToName2: {"en": "Capital → Country ×2", "fr": "Capitale → Pays ×2", "es": "Capital → País ×2"},
		fvFlagToCapital:   {"en": "Flag → Capital", "fr": "Drapeau → Capitale", "es": "Bandera → Capital"},
		fvFlagToText:      {"en": "Flag → Country (text)", "fr": "Drapeau → Pays (texte)", "es": "Bandera → País (texto)"},
	}

	fvDesc := map[string]model.I18nStrings{
		fvMix: {"en": "A mix of flag exercises", "fr": "Un mélange d'exercices sur les drapeaux", "es": "Una combinación de ejercicios de banderas"},
		fvFlagToName4:    {"en": "See a flag, pick the country from 4 options", "fr": "Voyez un drapeau, choisissez le pays parmi 4 options", "es": "Ve una bandera, elige el país entre 4 opciones"},
		fvFlagToName2:    {"en": "See a flag, pick the country from 2 options", "fr": "Voyez un drapeau, choisissez le pays parmi 2 options", "es": "Ve una bandera, elige el país entre 2 opciones"},
		fvNameToFlag4:    {"en": "See a country name, pick the flag from 4 options", "fr": "Voyez un nom de pays, choisissez le drapeau parmi 4 options", "es": "Ve un nombre de país, elige la bandera entre 4 opciones"},
		fvNameToFlag2:    {"en": "See a country name, pick the flag from 2 options", "fr": "Voyez un nom de pays, choisissez le drapeau parmi 2 options", "es": "Ve un nombre de país, elige la bandera entre 2 opciones"},
		slugCapitalToFlag4: {"en": "See a capital, pick the flag from 4 options", "fr": "Voyez une capitale, choisissez le drapeau parmi 4 options", "es": "Ve una capital, elige la bandera entre 4 opciones"},
		slugCapitalToFlag2: {"en": "See a capital, pick the flag from 2 options", "fr": "Voyez une capitale, choisissez le drapeau parmi 2 options", "es": "Ve una capital, elige la bandera entre 2 opciones"},
		slugCapitalToName4: {"en": "See a capital, pick the country from 4 options", "fr": "Voyez une capitale, choisissez le pays parmi 4 options", "es": "Ve una capital, elige el país entre 4 opciones"},
		slugCapitalToName2: {"en": "See a capital, pick the country from 2 options", "fr": "Voyez une capitale, choisissez le pays parmi 2 options", "es": "Ve una capital, elige el país entre 2 opciones"},
		fvFlagToCapital:   {"en": "See a flag, name the capital city", "fr": "Voyez un drapeau, nommez la capitale", "es": "Ve una bandera, nombra la capital"},
		fvFlagToText:      {"en": "See a flag, type the country name", "fr": "Voyez un drapeau, tapez le nom du pays", "es": "Ve una bandera, escribe el nombre del país"},
	}

	combineI18n := func(prefix model.I18nStrings, sep string, label model.I18nStrings) model.I18nStrings {
		result := make(model.I18nStrings)
		for _, lang := range []string{"en", "fr", "es"} {
			result[lang] = prefix[lang] + sep + label[lang]
		}
		return result
	}

	generalName := model.I18nStrings{"en": "General Knowledge", "fr": "Culture Générale", "es": "Cultura General"}
	generalDesc := model.I18nStrings{
		"en": "Answer general knowledge questions",
		"fr": "Répondez à des questions de culture générale",
		"es": "Responde preguntas de cultura general",
	}

	flagVariants := []struct{ variant, label string }{
		{fvMix, labelMix},
		{fvFlagToName4, "Flag → Country ×4"},
		{fvFlagToName2, "Flag → Country ×2"},
		{fvNameToFlag4, "Country → Flag ×4"},
		{fvNameToFlag2, "Country → Flag ×2"},
		{slugCapitalToFlag4, "Capital → Flag ×4"},
		{slugCapitalToFlag2, "Capital → Flag ×2"},
		{slugCapitalToName4, "Capital → Country ×4"},
		{slugCapitalToName2, "Capital → Country ×2"},
		{fvFlagToCapital, "Flag → Capital"},
		{fvFlagToText, "Flag → Country (text)"},
	}

	var defs []def

	defs = append(defs, def{
		name: "Solo — General Knowledge", slug: "solo-general",
		mode: modeSolo, category: categoryGeneral,
		minPlayers: 1, maxPlayers: 1, scoreMode: scoreModeClassic, timeBonus: false,
		qCount: 10, points: 100,
		nameI18n: combineI18n(modePrefix[modeSolo], " — ", generalName),
		descI18n: generalDesc,
	})

	for _, qt := range []struct{ label, slug, qtype string }{
		{labelMCQ4, "solo-general-mcq4", qtMCQ4},
		{labelMCQ2, "solo-general-mcq2", qtMCQ2},
		{labelTextInput, "solo-general-text", qtTextInput},
	} {
		qtLabelI18n := qtLabel[qt.qtype]
		nameI18n := model.I18nStrings{
			"en": modePrefix[modeSolo]["en"] + " — " + generalName["en"] + " (" + qtLabelI18n["en"] + ")",
			"fr": modePrefix[modeSolo]["fr"] + " — " + generalName["fr"] + " (" + qtLabelI18n["fr"] + ")",
			"es": modePrefix[modeSolo]["es"] + " — " + generalName["es"] + " (" + qtLabelI18n["es"] + ")",
		}
		defs = append(defs, def{
			name: "Solo — General Knowledge (" + qt.label + ")", slug: qt.slug,
			mode: modeSolo, category: categoryGeneral, questionType: qt.qtype,
			minPlayers: 1, maxPlayers: 1, scoreMode: scoreModeClassic, timeBonus: false,
			qCount: 10, points: 100,
			nameI18n: nameI18n, descI18n: generalDesc,
		})
	}

	for _, fv := range flagVariants {
		defs = append(defs, def{
			name: "Solo — " + fv.label,
			slug: "solo-flags-" + strings.ReplaceAll(fv.variant, "_", "-"),
			mode: modeSolo, category: "flags", flagVariant: fv.variant,
			minPlayers: 1, maxPlayers: 1, scoreMode: scoreModeClassic, timeBonus: false,
			qCount: 10, points: 100,
			nameI18n: combineI18n(modePrefix[modeSolo], " — ", fvLabel[fv.variant]),
			descI18n: fvDesc[fv.variant],
		})
	}

	defs = append(defs, def{
		name: "1v1 — General Knowledge", slug: "1v1-general",
		mode: label1v1, category: categoryGeneral,
		minPlayers: 2, maxPlayers: 2, scoreMode: scoreModeClassic, timeBonus: false,
		qCount: 10, points: 100,
		nameI18n: combineI18n(modePrefix[label1v1], " — ", generalName),
		descI18n: generalDesc,
	})

	for _, qt := range []struct{ label, slug, qtype string }{
		{labelMCQ4, "1v1-general-mcq4", qtMCQ4},
		{labelMCQ2, "1v1-general-mcq2", qtMCQ2},
		{labelTextInput, "1v1-general-text", qtTextInput},
	} {
		qtLabelI18n := qtLabel[qt.qtype]
		nameI18n := model.I18nStrings{
			"en": modePrefix[label1v1]["en"] + " — " + generalName["en"] + " (" + qtLabelI18n["en"] + ")",
			"fr": modePrefix[label1v1]["fr"] + " — " + generalName["fr"] + " (" + qtLabelI18n["fr"] + ")",
			"es": modePrefix[label1v1]["es"] + " — " + generalName["es"] + " (" + qtLabelI18n["es"] + ")",
		}
		defs = append(defs, def{
			name: "1v1 — General Knowledge (" + qt.label + ")", slug: qt.slug,
			mode: label1v1, category: categoryGeneral, questionType: qt.qtype,
			minPlayers: 2, maxPlayers: 2, scoreMode: scoreModeClassic, timeBonus: false,
			qCount: 10, points: 100,
			nameI18n: nameI18n, descI18n: generalDesc,
		})
	}

	for _, fv := range flagVariants {
		defs = append(defs, def{
			name: "1v1 — " + fv.label,
			slug: "1v1-flags-" + strings.ReplaceAll(fv.variant, "_", "-"),
			mode: label1v1, category: "flags", flagVariant: fv.variant,
			minPlayers: 2, maxPlayers: 2, scoreMode: scoreModeClassic, timeBonus: false,
			qCount: 10, points: 100,
			nameI18n: combineI18n(modePrefix[label1v1], " — ", fvLabel[fv.variant]),
			descI18n: fvDesc[fv.variant],
		})
	}

	created := 0
	for _, d := range defs {
		if _, err := u.templateRepo.GetBySlug(d.slug); err == nil {
			continue
		}

		t := &model.GameTemplate{
			Name:             d.name,
			Slug:             d.slug,
			Mode:             d.mode,
			Category:         d.category,
			FlagVariant:      d.flagVariant,
			QuestionType:     d.questionType,
			MinPlayers:       d.minPlayers,
			MaxPlayers:       d.maxPlayers,
			QuestionCount:    d.qCount,
			PointsPerCorrect: d.points,
			TimeBonus:        d.timeBonus,
			ScoreMode:        d.scoreMode,
			Language:         "en",
			IsActive:         true,
			NameI18n:         d.nameI18n,
			DescI18n:         d.descI18n,
		}
		if err := u.templateRepo.Create(t); err != nil {
			u.logger.Warn("Failed to seed game template", zap.String("slug", d.slug), zap.Error(err))
			continue
		}
		created++
	}

	u.logger.Info("Seeded default game templates", zap.Int("created", created))
	return created, nil
}

func (u *AdminGameTemplatesUsecase) DeleteGameTemplate(id uuid.UUID, adminID uuid.UUID) error {
	if _, err := u.templateRepo.GetByID(id); err != nil {
		return fmt.Errorf("game template not found: %w", err)
	}

	if err := u.templateRepo.Delete(id); err != nil {
		u.logger.Error("Failed to delete game template", zap.String("id", id.String()), zap.Error(err))
		return fmt.Errorf("failed to delete game template: %w", err)
	}

	_ = u.loggingService.LogAdminAction(adminID, "", "delete", "game_template", &id, "", "", nil, true, nil)

	return nil
}
