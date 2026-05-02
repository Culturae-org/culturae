// backend/internal/usecase/game.go

package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/Culturae-org/culturae/internal/game"
	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/pkg/identifier"
	"github.com/Culturae-org/culturae/internal/repository"
	adminRepo "github.com/Culturae-org/culturae/internal/repository/admin"
	"github.com/Culturae-org/culturae/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

type GameUsecase struct {
	gameRepo          repository.GameRepositoryInterface
	questionRepo      repository.QuestionRepositoryInterface
	adminQuestionRepo adminRepo.AdminQuestionRepositoryInterface
	geographyRepo     repository.GeographyRepositoryInterface
	userRepo          repository.UserRepositoryInterface
	friendsRepo       repository.FriendsRepositoryInterface
	templateRepo      repository.GameTemplateRepositoryInterface
	gameManager       game.GameManagerInterface
	loggingService    service.LoggingServiceInterface
	wsService         service.WebSocketServiceInterface
	datasetReader     DatasetReader
	xpCalc            *game.XPCalculator
	eloCalc           *game.ELOCalculator
	notifRepo         repository.NotificationRepositoryInterface
	logger            *zap.Logger
}

func NewGameUsecase(
	gameRepo repository.GameRepositoryInterface,
	questionRepo repository.QuestionRepositoryInterface,
	adminQuestionRepo adminRepo.AdminQuestionRepositoryInterface,
	geographyRepo repository.GeographyRepositoryInterface,
	userRepo repository.UserRepositoryInterface,
	friendsRepo repository.FriendsRepositoryInterface,
	templateRepo repository.GameTemplateRepositoryInterface,
	gameManager game.GameManagerInterface,
	loggingService service.LoggingServiceInterface,
	wsService service.WebSocketServiceInterface,
	datasetReader DatasetReader,
	xpCalc *game.XPCalculator,
	eloCalc *game.ELOCalculator,
	notifRepo repository.NotificationRepositoryInterface,
	logger *zap.Logger,
) *GameUsecase {
	return &GameUsecase{
		gameRepo:          gameRepo,
		questionRepo:      questionRepo,
		adminQuestionRepo: adminQuestionRepo,
		geographyRepo:     geographyRepo,
		userRepo:          userRepo,
		friendsRepo:       friendsRepo,
		templateRepo:      templateRepo,
		gameManager:       gameManager,
		loggingService:    loggingService,
		wsService:         wsService,
		datasetReader:     datasetReader,
		xpCalc:            xpCalc,
		eloCalc:           eloCalc,
		notifRepo:         notifRepo,
		logger:            logger,
	}
}

// -----------------------------------------------------
// User Game Usecase
//
// - CreateMatchmakedGame
// - CreateGame
// - InviteToGame
// - AcceptGameInvite
// - RejectGameInvite
// - JoinGame
// - LeaveGame
// - StartGame
// - SubmitAnswer
// - finalizeGame
// - finalizeGameWithWinnerTx
// - GetGameStatus
// - GetActiveGames
// - GetUserGameInvites
// - GetUserStats
// - GetGameHistory
// - CancelGame
// - selectRandomQuestions
// - selectRandomQuestionsFromAll
//
// -----------------------------------------------------

func (u *GameUsecase) CreateMatchmakedGame(user1, user2 uuid.UUID, mode model.GameMode, gameParams map[string]interface{}) error {
	questionCount := 10
	if qc, ok := gameParams["question_count"].(float64); ok {
		questionCount = int(qc)
	}

	category := ""
	if cat, ok := gameParams["category"].(string); ok {
		category = cat
	}

	flagVariant := ""
	if fv, ok := gameParams["flag_variant"].(string); ok {
		flagVariant = fv
	}

	language := ""
	if lang, ok := gameParams["language"].(string); ok {
		language = lang
	}

	questionType := ""
	if qt, ok := gameParams["question_type"].(string); ok {
		questionType = qt
	}

	var questions []*model.Question
	var geoQuestions []*model.GameQuestion
	var datasetID *uuid.UUID
	isGeography := category == model.CategoryFlags || category == model.CategoryGeography
	if isGeography {
		defaultGeoDS, geoErr := u.geographyRepo.GetDefaultDataset()
		if geoErr != nil {
			isGeography = false
		} else {
			datasetID = &defaultGeoDS.ID
			gq, vq, geoErr2 := u.generateGeographyQuestions(defaultGeoDS.ID, questionCount, "", false, flagVariant)
			if geoErr2 != nil {
				isGeography = false
			} else {
				geoQuestions = gq
				questions = vq
			}
		}
	}
	if !isGeography {
		var qErr error
		questions, qErr = u.selectRandomQuestions(nil, questionCount, questionType)
		if qErr != nil || len(questions) == 0 {
			return fmt.Errorf("matchmaking: failed to select questions: %w", qErr)
		}
	}

	publicID := uuid.New().String()[:8]
	now := time.Now()

	user1Data, err := u.userRepo.GetByID(user1.String())
	if err != nil {
		return fmt.Errorf("matchmaking: failed to get user1: %w", err)
	}

	if category == "" {
		category = model.CategoryGeneral
	}
	gameModel := &model.Game{
		PublicID:         publicID,
		Mode:             mode,
		Category:         category,
		FlagVariant:      flagVariant,
		Language:         language,
		DatasetID:        datasetID,
		Status:           model.GameStatusWaiting,
		CreatorID:        user1,
		CreatorPublicID:  user1Data.PublicID,
		QuestionCount:    questionCount,
		PointsPerCorrect: 100,
		TimeBonus:        true,
		MaxPlayers:       2,
		MinPlayers:       2,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := u.gameRepo.CreateGame(gameModel); err != nil {
		return fmt.Errorf("matchmaking: failed to create game: %w", err)
	}

	for _, uid := range []uuid.UUID{user1, user2} {
		if err := u.gameRepo.AddPlayerToGame(&model.GamePlayer{
			GameID:   gameModel.ID,
			UserID:   uid,
			Score:    0,
			IsReady:  true,
			JoinedAt: now,
		}); err != nil {
			return fmt.Errorf("matchmaking: failed to add player %s: %w", uid, err)
		}
	}

	if isGeography {
		for _, gq := range geoQuestions {
			gq.GameID = gameModel.ID
			if err := u.gameRepo.AddQuestionToGame(gq); err != nil {
				u.logger.Warn("matchmaking: failed to add question to game", zap.Error(err))
			}
		}
	} else {
		for i, q := range questions {
			if err := u.gameRepo.AddQuestionToGame(&model.GameQuestion{
				GameID:      gameModel.ID,
				QuestionID:  &q.ID,
				OrderNumber: i + 1,
				EntityKey:   "q:" + q.ID.String(),
			}); err != nil {
				u.logger.Warn("matchmaking: failed to add question to game", zap.Error(err))
			}
		}
	}

	settings := game.DefaultGameSettings()
	settings.QuestionCount = questionCount
	settings.Category = category
	settings.FlagVariant = flagVariant
	settings.DatasetID = datasetID
	if mode == model.GameMode1v1 {
		settings.MaxPlayers = 2
		settings.MinPlayers = 2
		settings.ScoreMode = "fastest_wins"
	}

	_, err = u.gameManager.CreateGameWithPlayers(
		gameModel.ID, publicID, mode, settings, questions,
		[]uuid.UUID{user1, user2},
	)
	if err != nil {
		return fmt.Errorf("matchmaking: failed to create engine: %w", err)
	}

	if err := u.gameManager.SetPlayerReady(gameModel.ID, user1, true); err != nil {
		return fmt.Errorf("matchmaking: failed to set player ready: %w", err)
	}
	if err := u.gameManager.SetPlayerReady(gameModel.ID, user2, true); err != nil {
		return fmt.Errorf("matchmaking: failed to set player ready: %w", err)
	}

	for _, uid := range []uuid.UUID{user1, user2} {
		_ = u.userRepo.UpdateUserGameStatus(uid, &gameModel.ID)
		_ = u.wsService.SendToUser(uid, map[string]interface{}{
			keyType:           "match_found",
			keyGamePublicID: publicID,
			keyMode:           string(mode),
		})
	}

	u.logger.Info("Match found, notifying clients",
		zap.String(keyGameID, gameModel.ID.String()),
		zap.String(keyPublicID, publicID),
		zap.String("u1", user1.String()),
		zap.String("u2", user2.String()),
	)

	time.Sleep(2 * time.Second)

	if err := u.gameManager.StartGame(gameModel.ID); err != nil {
		return fmt.Errorf("matchmaking: failed to start game: %w", err)
	}

	gameModel.Status = model.GameStatusInProgress
	gameModel.UpdatedAt = time.Now()
	_ = u.gameRepo.UpdateGame(gameModel)

	u.logger.Info("Matchmaked game started",
		zap.String(keyGameID, gameModel.ID.String()),
		zap.String(keyPublicID, publicID),
		zap.String("u1", user1.String()),
		zap.String("u2", user2.String()),
	)

	return nil
}

func (u *GameUsecase) CreateGame(c *gin.Context, creatorID uuid.UUID, req model.CreateGameRequest) (*model.Game, error) {
	if req.Mode != model.GameModeSolo && req.Mode != model.GameMode1v1 && req.Mode != model.GameModeMulti {
		return nil, errors.New("invalid game mode")
	}

	questionCount := 10
	pointsPerCorrect := 100
	timeBonus := true
	scoreMode := "time_bonus"
	maxPlayers := 1
	minPlayers := 1
	var templateID *uuid.UUID

	if req.Mode == model.GameMode1v1 {
		maxPlayers = 2
		minPlayers = 2
		scoreMode = "fastest_wins"
	}
	if req.Mode == model.GameModeMulti {
		maxPlayers = 4
		minPlayers = 2
		scoreMode = "time_bonus"
	}

	if req.TemplateID != nil {
		tmpl, tmplErr := u.templateRepo.GetByID(*req.TemplateID)
		if tmplErr != nil {
			return nil, fmt.Errorf("game template not found: %w", tmplErr)
		}
		if !tmpl.IsActive {
			return nil, errors.New("game template is not active")
		}
		questionCount = tmpl.QuestionCount
		pointsPerCorrect = tmpl.PointsPerCorrect
		timeBonus = tmpl.TimeBonus
		scoreMode = tmpl.ScoreMode
		maxPlayers = tmpl.MaxPlayers
		minPlayers = tmpl.MinPlayers
		if tmpl.DatasetID != nil && req.DatasetID == nil {
			req.DatasetID = tmpl.DatasetID
		}
		if tmpl.Category != "" && req.Category == "" {
			req.Category = tmpl.Category
		}
		if tmpl.FlagVariant != "" && req.FlagVariant == "" {
			req.FlagVariant = tmpl.FlagVariant
		}
		if tmpl.QuestionType != "" && req.QuestionType == "" {
			req.QuestionType = tmpl.QuestionType
		}
		if tmpl.Continent != "" && req.Continent == "" {
			req.Continent = tmpl.Continent
		}
		if tmpl.IncludeTerritories && !req.IncludeTerritories {
			req.IncludeTerritories = tmpl.IncludeTerritories
		}
		if tmpl.Language != "" && req.Language == "" {
			req.Language = tmpl.Language
		}
		templateID = &tmpl.ID

		if snapshot, err := json.Marshal(map[string]interface{}{
			keyName:               tmpl.Name,
			keySlug:               tmpl.Slug,
			"category":           tmpl.Category,
			"flag_variant":       tmpl.FlagVariant,
			"question_type":      tmpl.QuestionType,
			"language":           tmpl.Language,
			"continent":          tmpl.Continent,
			"question_count":     tmpl.QuestionCount,
			"points_per_correct": tmpl.PointsPerCorrect,
			"time_bonus":         tmpl.TimeBonus,
			"score_mode":         tmpl.ScoreMode,
			"xp_multiplier":      tmpl.XPMultiplier,
			keyMode:               tmpl.Mode,
		}); err == nil {
			req.TemplateSnapshot = string(snapshot)
		}
	}

	switch req.Mode {
	case model.GameModeSolo:
		minPlayers, maxPlayers = 1, 1
	case model.GameMode1v1:
		minPlayers, maxPlayers = 2, 2
	case model.GameModeMulti:
		if maxPlayers < 2 {
			maxPlayers = 2
		}
		if maxPlayers > 6 {
			maxPlayers = 6
		}
		if minPlayers < 2 {
			minPlayers = 2
		}
		if minPlayers > maxPlayers {
			minPlayers = maxPlayers
		}
	}

	if req.QuestionCount != nil {
		questionCount = *req.QuestionCount
	}
	if req.PointsPerCorrect != nil {
		pointsPerCorrect = *req.PointsPerCorrect
	}
	if req.TimeBonus != nil {
		timeBonus = *req.TimeBonus
	}

	publicID := identifier.GeneratePublicID()

	category := req.Category
	if category == "" {
		category = model.CategoryGeneral
	}
	flagVariant := req.FlagVariant
	continent := req.Continent
	includeTerritories := req.IncludeTerritories

	gameModel := &model.Game{
		PublicID:         publicID,
		Mode:             req.Mode,
		Status:           model.GameStatusWaiting,
		CreatorID:        creatorID,
		TemplateID:       templateID,
		TemplateSnapshot: req.TemplateSnapshot,
		MaxPlayers:       maxPlayers,
		MinPlayers:       minPlayers,
		QuestionCount:    questionCount,
		PointsPerCorrect: pointsPerCorrect,
		TimeBonus:        timeBonus,
		DatasetID:        req.DatasetID,
		Category:         category,
		FlagVariant:      flagVariant,
		QuestionType:     req.QuestionType,
		Language:         req.Language,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	creatorUser, err := u.userRepo.GetByID(creatorID.String())
	if err != nil {
		return nil, err
	}
	gameModel.CreatorPublicID = creatorUser.PublicID

	if req.Mode == model.GameModeSolo {
		if activeGames, activeErr := u.gameRepo.GetUserActiveGames(creatorID); activeErr == nil {
			for i := range activeGames {
				g := &activeGames[i]
				if g.Mode == model.GameModeSolo && g.Status == model.GameStatusInProgress {
					return g, nil
				}
			}
		}
	}

	if cancelledIDs, cancelErr := u.gameRepo.CancelUserWaitingGames(creatorID); cancelErr == nil {
		for _, gid := range cancelledIDs {
			_ = u.gameManager.CancelGame(gid)
		}
	}

	if err := u.gameRepo.CreateGame(gameModel); err != nil {
		return nil, err
	}

	gameCreationSucceeded := false
	defer func() {
		if !gameCreationSucceeded {
			if delErr := u.gameRepo.DeleteGameRecord(gameModel); delErr != nil {
				u.logger.Warn("Failed to clean up game record after creation error",
					zap.String(keyGameID, gameModel.ID.String()),
					zap.Error(delErr),
				)
			}
		}
	}()

	gamePlayer := &model.GamePlayer{
		GameID:   gameModel.ID,
		UserID:   creatorID,
		Score:    0,
		IsReady:  false,
		JoinedAt: time.Now(),
	}

	if err := u.gameRepo.AddPlayerToGame(gamePlayer); err != nil {
		return nil, err
	}

	var questions []*model.Question
	var gameQuestions []*model.GameQuestion

	isGeography := false
	isFlagQuiz := category == model.CategoryFlags

	if isFlagQuiz || category == model.CategoryGeography {
		if req.DatasetID != nil {
			_, datasetErr := u.geographyRepo.GetDatasetByID(*req.DatasetID)
			if datasetErr == nil {
				isGeography = true
			}
		} else {
			defaultGeoDS, geoErr := u.geographyRepo.GetDefaultDataset()
			if geoErr == nil {
				req.DatasetID = &defaultGeoDS.ID
				isGeography = true
			}
		}
	} else if req.DatasetID != nil {
		_, datasetErr := u.geographyRepo.GetDatasetByID(*req.DatasetID)
		if datasetErr == nil {
			isGeography = true
		}
	}

	if isGeography {
		var dynamicQuestions []*model.GameQuestion
		var virtualQuestions []*model.Question
		dynamicQuestions, virtualQuestions, err = u.generateGeographyQuestions(*req.DatasetID, questionCount, continent, includeTerritories, flagVariant)
		if err != nil {
			return nil, err
		}
		questions = virtualQuestions
		gameQuestions = dynamicQuestions
	} else {
		questions, err = u.selectRandomQuestions(req.DatasetID, questionCount, req.QuestionType)
		if err != nil {
			return nil, err
		}

		for _, q := range questions {
			if q.Kind == categoryGeography || q.Kind == "flags" {
				var answers []model.Answer
				if err := json.Unmarshal(q.Answers, &answers); err != nil {
					continue
				}

				changed := false
				for i, a := range answers {
					if uid, err := uuid.Parse(a.Slug); err == nil {
						country, err := u.geographyRepo.GetCountryByID(uid)
						if err == nil {
							if country.ISOAlpha2 != "" {
								answers[i].Slug = country.ISOAlpha2
								changed = true
							} else if country.Slug != "" {
								answers[i].Slug = country.Slug
								changed = true
							}
						}
					}
				}

				if changed {
					newBytes, _ := json.Marshal(answers)
					q.Answers = datatypes.JSON(newBytes)
				}
			}
		}
	}

	for i := range questionCount {
		var gameQuestion *model.GameQuestion

		if isGeography {
			if i < len(gameQuestions) {
				gameQuestion = gameQuestions[i]
				gameQuestion.GameID = gameModel.ID
			}
		} else {
			if i < len(questions) {
				q := questions[i]
				if (req.QuestionType == "single_choice_2" || req.QuestionType == "mcq_2_mix") && q.QType != model.QTypeTrueFalse {
					trimAnswersToTwo(q)
				}
				gameQuestion = &model.GameQuestion{
					GameID:      gameModel.ID,
					QuestionID:  &q.ID,
					OrderNumber: i + 1,
					EntityKey:   "q:" + q.ID.String(),
				}
			}
		}

		if gameQuestion != nil {
			if err := u.gameRepo.AddQuestionToGame(gameQuestion); err != nil {
				u.logger.Warn("Failed to add question to game", zap.Error(err))
			}
		}
	}

	settings := game.GameSettings{
		QuestionCount:      questionCount,
		PointsPerCorrect:   pointsPerCorrect,
		TimeBonus:          timeBonus,
		DatasetID:          req.DatasetID,
		Category:           category,
		FlagVariant:        flagVariant,
		Continent:          continent,
		IncludeTerritories: includeTerritories,
		InterRoundDelayMs:  2000,
		MaxPlayers:         maxPlayers,
		MinPlayers:         minPlayers,
		ScoreMode:          scoreMode,
		TemplateID:         templateID,
	}

	_, err = u.gameManager.CreateGame(gameModel.ID, gameModel.PublicID, req.Mode, settings, questions)
	if err != nil {
		return nil, err
	}

	if err := u.gameManager.AddPlayerToGame(gameModel.ID, creatorID); err != nil {
		return nil, err
	}

	if req.Mode == model.GameModeSolo {
		if err := u.gameManager.SetPlayerReady(gameModel.ID, creatorID, true); err != nil {
			u.logger.Error("Solo auto-start: failed to set player ready", zap.Error(err))
			return nil, fmt.Errorf("failed to initialise solo game: %w", err)
		} else if err := u.gameManager.StartGame(gameModel.ID); err != nil {
			u.logger.Error("Solo auto-start: failed to start game", zap.Error(err))
			return nil, fmt.Errorf("failed to start solo game: %w", err)
		} else {
			now := time.Now()
			gameModel.Status = model.GameStatusInProgress
			gameModel.StartedAt = &now
			gameModel.UpdatedAt = now
			_ = u.gameRepo.UpdateGame(gameModel)
		}
	}

	gameCreationSucceeded = true

	_ = u.loggingService.LogUserAction(creatorID, "create_game", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{
		keyGameID: gameModel.ID,
		keyMode:    req.Mode,
	}, true, nil)

	u.wsService.BroadcastAdminNotification(service.AdminNotification{
		Event: "game_created",
		Data: map[string]interface{}{
			keyMode:       string(req.Mode),
			"creator_id": creatorID.String(),
		},
		EntityType: "game",
		EntityID:   gameModel.PublicID,
		ActionURL:  "/games/" + gameModel.PublicID,
	})

	return gameModel, nil
}

func (u *GameUsecase) InviteToGame(c *gin.Context, gameID, fromUserID uuid.UUID, toUserPublicID string) (*model.GameInvite, error) {
	gameModel, err := u.gameRepo.GetGameByID(gameID)
	if err != nil {
		return nil, errors.New("game not found")
	}

	if err := u.validateGameCanBeModified(gameModel); err != nil {
		return nil, err
	}

	players, err := u.gameRepo.GetGamePlayers(gameID)
	if err != nil {
		return nil, err
	}

	isInGame := false
	for _, p := range players {
		if p.UserID == fromUserID {
			isInGame = true
			break
		}
	}
	if !isInGame {
		return nil, errors.New("you must be in the game to invite others")
	}

	toUser, err := u.userRepo.GetByPublicID(toUserPublicID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	isFriend, _ := u.friendsRepo.IsFriend(fromUserID, toUser.ID)

	if !toUser.IsProfilePublic && !isFriend {
		return nil, errors.New("this user's profile is private")
	}

	if !toUser.AllowPartyInvites {
		return nil, errors.New("user does not accept game invites")
	}

	isBlocked, err := u.friendsRepo.IsBlocked(fromUserID, toUser.ID)
	if err != nil {
		return nil, err
	}
	if isBlocked {
		return nil, errors.New("you cannot invite this user")
	}

	invite := &model.GameInvite{
		GameID:     gameID,
		FromUserID: fromUserID,
		ToUserID:   toUser.ID,
		Status:     model.GameInviteStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := u.gameRepo.CreateGameInvite(invite); err != nil {
		return nil, err
	}

	invite.ToUserPublicID = toUser.PublicID

	_ = u.loggingService.LogUserAction(fromUserID, "invite_to_game", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{
		keyGameID:           gameID,
		"to_user_public_id": toUserPublicID,
	}, true, nil)

	fromUser, _ := u.userRepo.GetByID(fromUserID.String())
	username := "Someone"
	if fromUser != nil {
		username = fromUser.Username
		invite.FromUserPublicID = fromUser.PublicID
	}

	if u.wsService != nil {
		go func() {
			_ = u.wsService.SendToUser(toUser.ID, map[string]interface{}{
				keyType: "game_invite",
				keyData: map[string]interface{}{
					keyInviteID:      invite.ID.String(),
					keyGameID:        gameID.String(),
					keyGamePublicID: gameModel.PublicID,
					"from_username":  username,
					keyFromPublicID: invite.FromUserPublicID,
					"game_mode":      string(gameModel.Mode),
				},
			})
		}()
	}

	if u.notifRepo != nil {
		data, _ := json.Marshal(map[string]string{keyInviteID: invite.ID.String()})
		_ = u.notifRepo.Create(&model.Notification{
			UserID: toUser.ID,
			Type:   "game_invite",
			Title:  "Game invitation",
			Body:   username + " invited you to a game",
			Data:   datatypes.JSON(data),
		})
	}

	return invite, nil
}

func (u *GameUsecase) AcceptGameInvite(c *gin.Context, inviteID, userID uuid.UUID) (*model.GameInvite, error) {
	invite, err := u.gameRepo.GetGameInviteByID(inviteID)
	if err != nil {
		return nil, errors.New("invite not found")
	}

	if invite.ToUserID != userID {
		return nil, errors.New("not authorized")
	}

	if invite.Status != model.GameInviteStatusPending {
		return nil, errors.New("invite already processed")
	}

	if err := u.gameRepo.UpdateGameInviteStatus(inviteID, model.GameInviteStatusAccepted); err != nil {
		return nil, err
	}

	if err := u.JoinGame(c, invite.GameID, userID); err != nil {
		return nil, err
	}

	gameModel, err := u.gameRepo.GetGameByID(invite.GameID)
	if err != nil {
		return nil, errors.New("game not found")
	}

	if gameModel.Mode == model.GameModeMulti {
		if err := u.gameManager.SetPlayerReady(invite.GameID, userID, true); err != nil {
			u.logger.Warn("AcceptGameInvite(multi): failed to set invitee ready", zap.Error(err))
		}
	} else {
		if err := u.gameManager.SetPlayerReady(invite.GameID, invite.FromUserID, true); err != nil {
			u.logger.Warn("AcceptGameInvite: failed to set creator ready", zap.Error(err))
		}
		if err := u.gameManager.SetPlayerReady(invite.GameID, userID, true); err != nil {
			u.logger.Warn("AcceptGameInvite: failed to set invitee ready", zap.Error(err))
		}

		gameIDCopy := invite.GameID
		go func() {
			time.Sleep(1500 * time.Millisecond)
			if err := u.gameManager.StartGame(gameIDCopy); err != nil {
				u.logger.Warn("Could not auto-start game after invite accept", zap.Error(err))
				return
			}
			if fresh, err := u.gameRepo.GetGameByID(gameIDCopy); err == nil {
				now := time.Now()
				fresh.Status = model.GameStatusInProgress
				fresh.StartedAt = &now
				fresh.UpdatedAt = now
				_ = u.gameRepo.UpdateGame(fresh)
			}
		}()
	}

	_ = u.loggingService.LogUserAction(userID, "accept_game_invite", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{
		keyInviteID: inviteID,
		keyGameID:   invite.GameID,
	}, true, nil)

	if u.wsService != nil {
		go func() {
			acceptorData := map[string]interface{}{
				keyInviteID:      inviteID.String(),
				keyGamePublicID: gameModel.PublicID,
				"status":         string(model.GameInviteStatusAccepted),
			}
			if acceptor, err := u.userRepo.GetByID(userID.String()); err == nil {
				acceptorData["from_username"] = acceptor.Username
				acceptorData[keyFromPublicID] = acceptor.PublicID
			}
			_ = u.wsService.SendToUser(invite.FromUserID, map[string]interface{}{
				keyType: "game_invite_accepted",
				keyData: acceptorData,
			})
			_ = u.wsService.SendToUser(userID, map[string]interface{}{
				keyType: "game_invite_updated",
				keyData: map[string]interface{}{
					keyInviteID:      inviteID.String(),
					keyGamePublicID: gameModel.PublicID,
					"status":         string(model.GameInviteStatusAccepted),
				},
			})
		}()
	}

	return invite, nil
}

func (u *GameUsecase) RejectGameInvite(c *gin.Context, inviteID, userID uuid.UUID) error {
	invite, err := u.gameRepo.GetGameInviteByID(inviteID)
	if err != nil {
		return errors.New("invite not found")
	}

	if invite.ToUserID != userID {
		return errors.New("not authorized")
	}

	if invite.Status != model.GameInviteStatusPending {
		return errors.New("invite already processed")
	}

	gameModel, err := u.gameRepo.GetGameByID(invite.GameID)
	if err != nil {
		return errors.New("game not found")
	}

	if err := u.gameRepo.UpdateGameInviteStatus(inviteID, model.GameInviteStatusRejected); err != nil {
		return err
	}

	_ = u.loggingService.LogUserAction(userID, "reject_game_invite", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{
		keyInviteID: inviteID,
	}, true, nil)

	if u.wsService != nil {
		go func() {
			_ = u.wsService.SendToUser(invite.FromUserID, map[string]interface{}{
				keyType: "game_invite_rejected",
				keyData: map[string]interface{}{
					keyInviteID:      inviteID.String(),
					keyGamePublicID: gameModel.PublicID,
				},
			})
			_ = u.wsService.SendToUser(userID, map[string]interface{}{
				keyType: "game_invite_updated",
				keyData: map[string]interface{}{
					keyInviteID:      inviteID.String(),
					keyGamePublicID: gameModel.PublicID,
					"status":         string(model.GameInviteStatusRejected),
				},
			})
		}()
	}

	return nil
}

func (u *GameUsecase) JoinGame(c *gin.Context, gameID, userID uuid.UUID) error {
	gameModel, err := u.gameRepo.GetGameByID(gameID)
	if err != nil {
		return errors.New("game not found")
	}

	if err := u.validateGameCanBeModified(gameModel); err != nil {
		return err
	}

	if gameModel.CreatorID != userID {
		if gameModel.Mode == model.GameModeSolo {
			return errors.New("you are not authorized to join this game")
		}
		if isBlocked, err := u.friendsRepo.IsBlocked(userID, gameModel.CreatorID); err == nil && isBlocked {
			return errors.New("you cannot join this game")
		}
	}

	gamePlayer := &model.GamePlayer{
		GameID:   gameID,
		UserID:   userID,
		Score:    0,
		IsReady:  false,
		JoinedAt: time.Now(),
	}

	if err := u.gameRepo.AddPlayerToGame(gamePlayer); err != nil {
		return err
	}

	if err := u.gameManager.AddPlayerToGame(gameID, userID); err != nil {
		return err
	}

	if err := u.userRepo.UpdateUserGameStatus(userID, &gameID); err != nil {
		u.logger.Warn("Failed to update user game status", zap.Error(err))
	}

	if u.wsService != nil {
		go func() {
			playerData := map[string]interface{}{keyAction: "joined"}
			if joiner, err := u.userRepo.GetByID(userID.String()); err == nil {
				playerData[keyUserPublicID] = joiner.PublicID
				playerData["username"] = joiner.Username
			}
			_ = u.wsService.SendToGame(gameModel.PublicID, map[string]interface{}{
				keyType:           msgLobbyUpdated,
				keyGamePublicID: gameModel.PublicID,
				keyData:           playerData,
			})
		}()
	}

	_ = u.loggingService.LogUserAction(userID, "join_game", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{
		keyGameID: gameID,
	}, true, nil)

	return nil
}

func (u *GameUsecase) LeaveGame(c *gin.Context, gameID, userID uuid.UUID) error {
	gameModel, err := u.gameRepo.GetGameByID(gameID)
	if err != nil {
		return errors.New("game not found")
	}

	var player *model.GamePlayer
	for _, p := range gameModel.Players {
		if p.UserID == userID {
			player = &p
			break
		}
	}
	if player == nil {
		return errors.New("user not in game")
	}

	if gameModel.Status == model.GameStatusCompleted || gameModel.Status == model.GameStatusCancelled {
		return errors.New("game is already finished")
	}

	wasInProgress := gameModel.Status == model.GameStatusInProgress
	var winnerPublicID string
	var wasCancelled bool

	err = u.gameRepo.WithTransaction(func(txRepo repository.GameRepositoryInterface) error {
		switch gameModel.Status {
		case model.GameStatusWaiting, model.GameStatusReady:
			if err := txRepo.DeletePlayerFromGame(gameID, userID); err != nil {
				return err
			}
			playerCount, err := txRepo.CountPlayersInGame(gameID)
			if err != nil {
				return err
			}
			if playerCount == 0 {
				if err := txRepo.DeleteGameRecord(gameModel); err != nil {
					return err
				}
			}
		case model.GameStatusInProgress:
			if err := txRepo.UpdatePlayerStatus(gameID, userID, string(model.PlayerStatusLeft)); err != nil {
				return err
			}
			if err := txRepo.UpdateUserGamesLost(userID, string(gameModel.Mode)); err != nil {
				return err
			}
			activePlayers, err := txRepo.FindActivePlayersInGame(gameID, string(model.PlayerStatusLeft))
			if err != nil {
				return err
			}
			if len(activePlayers) == 0 {
				wasCancelled = true
				gameModel.Status = model.GameStatusCancelled
				gameModel.UpdatedAt = time.Now()
				if err := txRepo.SaveGame(gameModel); err != nil {
					return err
				}
			} else if len(activePlayers) == 1 || gameModel.Mode == model.GameMode1v1 {
				winnerID := activePlayers[0].UserID
				for _, p := range gameModel.Players {
					if p.UserID == winnerID {
						if p.User != nil {
							winnerPublicID = p.User.PublicID
						} else {
							winnerPublicID = p.UserPublicID
						}
						break
					}
				}
				if err := u.finalizeGameWithWinnerTx(txRepo, gameID, &winnerID); err != nil {
					return err
				}
			}
		default:
			return errors.New("game is already finished")
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := u.gameManager.RemovePlayerFromGame(gameID, userID); err != nil {
		u.logger.Warn("Failed to remove player from game manager", zap.Error(err))
	}
	if err := u.userRepo.UpdateUserGameStatus(userID, nil); err != nil {
		u.logger.Warn("Failed to clear user game status", zap.Error(err))
	}

	u.emitEvent(gameID, "player_left", map[string]interface{}{keyUserID: userID})

	if !wasInProgress && u.wsService != nil {
		go func() {
			_ = u.wsService.SendToGame(gameModel.PublicID, map[string]interface{}{
				keyType:           msgLobbyUpdated,
				keyGamePublicID: gameModel.PublicID,
				keyData: map[string]interface{}{
					keyUserID: userID.String(),
					keyAction:  "left",
				},
			})
		}()
	}

	if wasInProgress {
		playersFinal := make([]map[string]interface{}, 0, len(gameModel.Players))
		for _, p := range gameModel.Players {
			username := p.UserPublicID
			if p.User != nil {
				username = p.User.Username
			}
			playersFinal = append(playersFinal, map[string]interface{}{
				keyUserPublicID: p.UserPublicID,
				"username":       username,
				keyScore:          p.Score,
			})
		}
		if wasCancelled {
			if err := u.gameManager.CancelGame(gameID); err != nil {
				u.logger.Warn("Failed to cancel game in manager", zap.Error(err))
			}
			_ = u.wsService.SendToGame(gameModel.PublicID, map[string]interface{}{
				keyType:      "game_cancelled",
				keyPublicID: gameModel.PublicID,
				keyData: map[string]interface{}{
					"reason": "all_players_left",
				},
			})
		} else if winnerPublicID != "" {
			_ = u.wsService.SendToGame(gameModel.PublicID, map[string]interface{}{
				keyType:      "game_completed",
				keyGameID:   gameID,
				keyPublicID: gameModel.PublicID,
				keyData: map[string]interface{}{
					keyGameID:          gameID,
					"players_final":    playersFinal,
					keyWinnerPublicID: winnerPublicID,
				},
			})
		}
	}

	if c != nil {
		_ = u.loggingService.LogUserAction(userID, "leave_game", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{
			keyGameID: gameID,
		}, true, nil)
	}

	return nil
}

func (u *GameUsecase) StartGame(c *gin.Context, gameID, userID uuid.UUID) error {
	var gameModel *model.Game
	var players []model.GamePlayer

	err := u.gameRepo.WithTransaction(func(txRepo repository.GameRepositoryInterface) error {
		var err error
		gameModel, err = txRepo.GetGameByIDSimple(gameID)
		if err != nil {
			return errors.New("game not found")
		}

		if gameModel.Status != model.GameStatusWaiting && gameModel.Status != model.GameStatusReady {
			return errors.New("game already started or finished")
		}

		players, err = txRepo.FindPlayersByGame(gameID)
		if err != nil {
			return err
		}

		isInGame := false
		for _, p := range players {
			if p.UserID == userID {
				isInGame = true
				break
			}
		}
		if !isInGame {
			return errors.New("you are not in this game")
		}

		if gameModel.Mode == model.GameModeMulti {
			if gameModel.CreatorID != userID {
				return errors.New("only the game creator can start a multi game")
			}
			if len(players) < gameModel.MinPlayers {
				return fmt.Errorf("not enough players (%d/%d)", len(players), gameModel.MinPlayers)
			}
			for _, p := range players {
				_ = txRepo.SetPlayerReady(gameID, p.UserID)
			}
			gameModel.Status = model.GameStatusReady
			gameModel.UpdatedAt = time.Now()
			return txRepo.SaveGame(gameModel)
		}

		if err := txRepo.SetPlayerReady(gameID, userID); err != nil {
			return err
		}

		allReady := true
		for _, p := range players {
			if p.UserID == userID {
				continue
			}
			if !p.IsReady {
				allReady = false
				break
			}
		}

		if !allReady {
			gameModel.Status = model.GameStatusWaiting
		} else {
			gameModel.Status = model.GameStatusReady
		}
		gameModel.UpdatedAt = time.Now()

		return txRepo.SaveGame(gameModel)
	})
	if err != nil {
		return err
	}

	if gameModel.Mode == model.GameModeMulti {
		for _, p := range players {
			_ = u.gameManager.SetPlayerReady(gameID, p.UserID, true)
		}
	} else {
		if err := u.gameManager.SetPlayerReady(gameID, userID, true); err != nil {
			return err
		}
	}

	gameEngine, err := u.gameManager.GetGame(gameID)
	if err != nil {
		return err
	}

	if gameEngine.CanStart() {
		if err := u.gameManager.StartGame(gameID); err != nil {
			return err
		}

		now := time.Now()
		gameModel.Status = model.GameStatusInProgress
		gameModel.StartedAt = &now
		gameModel.UpdatedAt = now

		if err := u.gameRepo.UpdateGame(gameModel); err != nil {
			u.logger.Warn("Failed to update game in database", zap.Error(err))
		}

		_ = u.loggingService.LogUserAction(userID, "start_game", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{
			keyGameID: gameID,
		}, true, nil)

		u.wsService.BroadcastAdminNotification(service.AdminNotification{
			Event: "game_started",
			Data: map[string]interface{}{
				keyMode: string(gameModel.Mode),
			},
			EntityType: "game",
			EntityID:   gameModel.PublicID,
			ActionURL:  "/games/" + gameModel.PublicID,
		})
	}

	return nil
}

func (u *GameUsecase) SubmitAnswer(c *gin.Context, gameID, userID uuid.UUID, req model.SubmitAnswerRequest) error {
	gameEngine, err := u.gameManager.GetGame(gameID)
	if err != nil {
		return errors.New("game not found or not active")
	}

	if req.QuestionID == (uuid.UUID{}) {
		return errors.New("question_id is required")
	}

	currentQuestion, err := gameEngine.GetCurrentQuestion()
	if err != nil {
		return err
	}

	if currentQuestion.ID != req.QuestionID {
		return errors.New("not the current question")
	}

	var answer game.Answer

	timeSpent := time.Duration(req.TimeSpent) * time.Millisecond

	switch currentQuestion.QType {
	case model.QTypeTextInput:
		text := req.AnswerSlug
		if len(req.Data) > 0 {
			var dataMap map[string]interface{}
			if err := json.Unmarshal(req.Data, &dataMap); err == nil {
				if val, ok := dataMap["text"].(string); ok {
					text = val
				}
			} else {
				var strVal string
				if err := json.Unmarshal(req.Data, &strVal); err == nil {
					text = strVal
				}
			}
		}
		answer = game.NewTextAnswer(req.QuestionID, text, timeSpent)

	case model.QTypeTrueFalse, "true_false":
		boolVal := false
		if req.AnswerSlug == model.AnswerSlugTrue {
			boolVal = true
		} else if len(req.Data) > 0 {
			var dataMap map[string]interface{}
			if err := json.Unmarshal(req.Data, &dataMap); err == nil {
				if v, ok := dataMap["value"].(bool); ok {
					boolVal = v
				} else if v, ok := dataMap["value"].(string); ok {
					boolVal = v == model.AnswerSlugTrue
				}
			}
		}
		answer = game.NewBooleanAnswer(req.QuestionID, boolVal, timeSpent)

	default:
		slug := req.AnswerSlug
		if slug == "" && len(req.Data) > 0 {
			var dataMap map[string]interface{}
			if err := json.Unmarshal(req.Data, &dataMap); err == nil {
				if val, ok := dataMap[keySlug].(string); ok {
					slug = val
				}
			}
		}
		answer = game.NewMCQAnswer(req.QuestionID, slug, timeSpent)
	}

	if err := u.gameManager.SubmitAnswer(gameID, userID, answer); err != nil {
		return err
	}

	freshEngine, err := u.gameManager.GetGame(gameID)
	if err != nil {
		freshEngine = gameEngine
	}

	player, err := freshEngine.GetPlayer(userID)
	if err != nil {
		return err
	}

	var lastAnswer game.Answer
	if len(player.Answers) > 0 {
		lastAnswer = player.Answers[len(player.Answers)-1]
	}

	players, _ := u.gameRepo.GetGamePlayers(gameID)
	var gamePlayerID uuid.UUID
	for _, p := range players {
		if p.UserID == userID {
			gamePlayerID = p.ID
			break
		}
	}

	var gameQuestionID *uuid.UUID
	if gq, err := u.gameRepo.GetGameQuestionByOrder(gameID, len(player.Answers)); err == nil && gq != nil {
		gameQuestionID = &gq.ID
	}

	richData := map[string]interface{}{
		"question_slug":        currentQuestion.Slug,
		"question_type":        currentQuestion.QType,
		"user_answer":          answer.AnswerData,
		"server_time_spent_ms": lastAnswer.ServerTimeSpent.Milliseconds(),
	}
	for k, v := range lastAnswer.Metadata {
		richData[k] = v
	}
	richDataBytes, _ := json.Marshal(richData)

	answerSlug := lastAnswer.GetAnswerSlug()

	gameAnswer := &model.GameAnswer{
		GameID:         gameID,
		PlayerID:       gamePlayerID,
		QuestionID:     &req.QuestionID,
		GameQuestionID: gameQuestionID,
		AnswerSlug:     answerSlug,
		Data:           datatypes.JSON(richDataBytes),
		IsCorrect:      lastAnswer.IsCorrect,
		TimeSpent:      req.TimeSpent,
		Points:         lastAnswer.Points,
		AnsweredAt:     time.Now(),
	}

	if err := u.gameRepo.SaveGameAnswer(gameAnswer); err != nil {
		u.logger.Warn("Failed to save answer to database", zap.Error(err))
	}

	u.logger.Info("Answer recorded",
		zap.String(keyGameID, gameID.String()),
		zap.String(keyUserID, userID.String()),
		zap.String("question_id", currentQuestion.ID.String()),
		zap.String("question_slug", currentQuestion.Slug),
		zap.String("question_type", currentQuestion.QType),
		zap.String("answer_slug", answerSlug),
		zap.Bool("is_correct", lastAnswer.IsCorrect),
		zap.Int("points", lastAnswer.Points),
		zap.Int("time_spent_ms", req.TimeSpent),
		zap.Int64("server_time_spent_ms", lastAnswer.ServerTimeSpent.Milliseconds()),
		zap.Any("metadata", lastAnswer.Metadata),
	)

	if err := u.gameRepo.UpdatePlayerScore(gamePlayerID, player.Score); err != nil {
		u.logger.Warn("Failed to update player score", zap.Error(err))
	}

	if freshEngine.AllPlayersFinished() {
		if err := u.finalizeGame(c, gameID); err != nil {
			u.logger.Error("Failed to finalize game", zap.Error(err))
		}
	}

	return nil
}

func (u *GameUsecase) finalizeGame(c *gin.Context, gameID uuid.UUID) error {
	gameEngine, err := u.gameManager.GetGame(gameID)
	if err != nil {
		return err
	}

	gameModel, err := u.gameRepo.GetGameByID(gameID)
	if err != nil {
		return err
	}

	now := time.Now()
	gameModel.Status = model.GameStatusCompleted
	gameModel.CompletedAt = &now

	allPlayers := gameEngine.GetPlayers()

	players := make([]game.Player, 0, len(allPlayers))
	for _, p := range allPlayers {
		if p.Status != model.PlayerStatusLeft {
			players = append(players, p)
		}
	}

	var winnerUserID uuid.UUID
	var winnerID *uuid.UUID
	maxScore := -1
	for _, p := range players {
		if p.Score > maxScore {
			maxScore = p.Score
			winnerUserID = p.UserID
			winnerID = &winnerUserID
		} else if p.Score == maxScore {
			winnerID = nil
		}
	}
	if gameModel.Mode == model.GameModeSolo && len(players) > 0 {
		winnerID = &players[0].UserID
	}
	gameModel.WinnerID = winnerID
	gameModel.UpdatedAt = now

	if err := u.gameRepo.UpdateGame(gameModel); err != nil {
		return err
	}

	var durationSec int64
	if gameModel.StartedAt != nil {
		durationSec = int64(now.Sub(*gameModel.StartedAt).Seconds())
	}

	xpCtx, xpCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer xpCancel()
	xpCfg := u.xpCalc.LoadConfig(xpCtx)

	var templateXPMultiplier *float64
	if gameModel.TemplateID != nil {
		if tmpl, tmplErr := u.templateRepo.GetByID(*gameModel.TemplateID); tmplErr == nil {
			templateXPMultiplier = tmpl.XPMultiplier
		}
	}

	isDrawn := winnerID == nil && gameModel.Mode != model.GameModeSolo

	for _, player := range players {
		isWinner := winnerID != nil && *winnerID == player.UserID

		if err := u.userRepo.UpdateGameStats(player.UserID, isWinner, isDrawn, player.Score, durationSec, string(gameModel.Mode)); err != nil {
			u.logger.Warn("Failed to update user game stats", zap.String(keyUserID, player.UserID.String()), zap.Error(err))
		}

		xp := u.xpCalc.CalculateXPWithTemplateMultiplier(gameModel.Mode, player.Score, isWinner, xpCfg, templateXPMultiplier)
		if xp > 0 {
			if err := u.userRepo.AddExperience(player.UserID, xp, xpCfg); err != nil {
				u.logger.Warn("Failed to add user experience", zap.String(keyUserID, player.UserID.String()), zap.Error(err))
			}
		}

		if err := u.userRepo.UpdateUserGameStatus(player.UserID, nil); err != nil {
			u.logger.Warn("Failed to clear user game status", zap.Error(err))
		}

		_ = u.wsService.SendToUser(player.UserID, map[string]interface{}{
			keyType:      "game_results",
			keyPublicID: gameModel.PublicID,
			keyData: map[string]interface{}{
				keyScore:     player.Score,
				"xp_gained": xp,
				"is_winner": isWinner,
			},
		})
	}

	u.emitGameCompletedEvent(gameModel, players, winnerID)

	if err := u.gameManager.RemoveGame(gameID); err != nil {
		u.logger.Warn("Failed to remove finalized game from Redis",
			zap.String(keyGameID, gameID.String()),
			zap.Error(err),
		)
	}

	if err := u.updateQuestionStats(gameID); err != nil {
		u.logger.Warn("Failed to update question stats",
			zap.String(keyGameID, gameID.String()),
			zap.Error(err),
		)
	}

	if gameModel.Mode != model.GameModeSolo && len(players) == 2 && winnerID != nil {
		var winPlayer, losePlayer game.Player
		for _, p := range players {
			if p.UserID == *winnerID {
				winPlayer = p
			} else {
				losePlayer = p
			}
		}

		winUser, err1 := u.userRepo.GetByID(winPlayer.UserID.String())
		loseUser, err2 := u.userRepo.GetByID(losePlayer.UserID.String())
		if err1 != nil || err2 != nil {
			u.logger.Warn("Failed to fetch user ELO data, skipping ELO update",
				zap.String("winner_id", winPlayer.UserID.String()),
				zap.String("loser_id", losePlayer.UserID.String()),
			)
		} else {
			var eloCtx context.Context
			var eloCancel context.CancelFunc
			if c != nil {
				eloCtx, eloCancel = context.WithTimeout(c.Request.Context(), 5*time.Second)
			} else {
				eloCtx, eloCancel = context.WithTimeout(context.Background(), 5*time.Second)
			}
			defer eloCancel()
			eloCfg := u.eloCalc.LoadConfig(eloCtx)
			newWinRating, newLoseRating := u.eloCalc.CalculateElo(
				winUser.EloRating, loseUser.EloRating,
				winUser.EloGamesPlayed, loseUser.EloGamesPlayed,
				eloCfg,
			)
			if err := u.userRepo.UpdateEloRating(winPlayer.UserID, newWinRating); err != nil {
				u.logger.Warn("Failed to update winner ELO", zap.String(keyUserID, winPlayer.UserID.String()), zap.Error(err))
			}
			if err := u.userRepo.UpdateEloRating(losePlayer.UserID, newLoseRating); err != nil {
				u.logger.Warn("Failed to update loser ELO", zap.String(keyUserID, losePlayer.UserID.String()), zap.Error(err))
			}
		}
	}

	return nil
}

func (u *GameUsecase) finalizeGameWithWinnerTx(txRepo repository.GameRepositoryInterface, gameID uuid.UUID, winnerID *uuid.UUID) error {
	gameModel, err := txRepo.GetGameByIDSimple(gameID)
	if err != nil {
		return err
	}

	now := time.Now()
	gameModel.Status = model.GameStatusCompleted
	gameModel.CompletedAt = &now
	gameModel.WinnerID = winnerID
	gameModel.UpdatedAt = now

	if err := txRepo.SaveGame(gameModel); err != nil {
		return err
	}

	if winnerID != nil {
		if err := txRepo.UpdateUserGamesWon(*winnerID, string(gameModel.Mode)); err != nil {
			return err
		}
	}

	players, err := txRepo.FindPlayersByGame(gameID)
	if err != nil {
		return err
	}
	for _, p := range players {
		if err := u.userRepo.UpdateUserGameStatus(p.UserID, nil); err != nil {
			u.logger.Warn("Failed to clear user game status", zap.Error(err))
		}
	}

	return nil
}

func (u *GameUsecase) GetGameStatus(gameID uuid.UUID) (*model.GameStatusResponse, error) {
	gameEngine, err := u.gameManager.GetGame(gameID)
	if err != nil {
		gameModel, dbErr := u.gameRepo.GetGameByID(gameID)
		if dbErr != nil {
			return nil, errors.New("game not found")
		}

		players, playersErr := u.gameRepo.GetGamePlayers(gameID)
		if playersErr != nil {
			u.logger.Warn("Failed to get game players", zap.Error(playersErr))
			players = gameModel.Players
		}

		response := &model.GameStatusResponse{
			Game:           gameModel,
			Players:        players,
			QuestionNumber: 0,
			TotalQuestions: gameModel.QuestionCount,
		}
		return response, nil
	}

	gameModel, err := u.gameRepo.GetGameByID(gameID)
	if err != nil {
		return nil, err
	}

	currentQuestion, _ := gameEngine.GetCurrentQuestion()

	response := &model.GameStatusResponse{
		Game:            gameModel,
		Players:         gameModel.Players,
		CurrentQuestion: currentQuestion,
		QuestionNumber:  gameEngine.GetQuestionNumber(),
		TotalQuestions:  gameEngine.GetTotalQuestions(),
	}

	return response, nil
}

func (u *GameUsecase) GetActiveGames(userID uuid.UUID) ([]model.Game, error) {
	return u.gameRepo.GetUserActiveGames(userID)
}

func (u *GameUsecase) GetUserGameInvites(userID uuid.UUID, status model.GameInviteStatus) ([]model.GameInvite, error) {
	return u.gameRepo.GetUserGameInvites(userID, status)
}

func (u *GameUsecase) GetUserStats(userID uuid.UUID) (*model.UserStats, error) {
	user, err := u.userRepo.GetByID(userID.String())
	if err != nil {
		return nil, err
	}

	gameStats, err := u.userRepo.GetUserGameStats(userID)
	if err != nil || gameStats == nil {
		gameStats = &model.UserGameStats{}
	}

	modeStatsRaw, _ := u.userRepo.GetUserGameStatsByMode(userID)
	recentGames, _ := u.gameRepo.GetUserRecentGames(userID, 10)

	winRate := 0.0
	if gameStats.TotalGames > 0 {
		winRate = float64(gameStats.GamesWon) / float64(gameStats.TotalGames) * 100
	}

	gamesByMode := make([]model.GameModeStats, len(modeStatsRaw))
	for i, ms := range modeStatsRaw {
		gamesByMode[i] = model.GameModeStats{
			Mode:         ms.Mode,
			TotalGames:   ms.TotalGames,
			GamesWon:     ms.GamesWon,
			GamesLost:    ms.GamesLost,
			GamesDrawn:   ms.GamesDrawn,
			TotalScore:   ms.TotalScore,
			AverageScore: ms.AverageScore,
			PlayTime:     ms.PlayTime,
		}
	}

	if recentGames == nil {
		recentGames = []model.RecentGameInfo{}
	}

	return &model.UserStats{
		TotalGames:     gameStats.TotalGames,
		GamesWon:       gameStats.GamesWon,
		GamesLost:      gameStats.GamesLost,
		GamesDrawn:     gameStats.GamesDrawn,
		WinRate:        winRate,
		DayStreak:      gameStats.DayStreak,
		BestDayStreak:  gameStats.BestDayStreak,
		TotalScore:     gameStats.TotalScore,
		AverageScore:   gameStats.AverageScore,
		PlayTime:       gameStats.PlayTime,
		EloRating:      user.EloRating,
		EloGamesPlayed: user.EloGamesPlayed,
		Experience:     user.Experience,
		Level:          user.Level,
		Rank:           user.Rank,
		GamesByMode:    gamesByMode,
		RecentGames:    recentGames,
	}, nil
}

func (u *GameUsecase) GetUserStatsByPeriod(userID uuid.UUID, period string) (*model.UserStats, error) {
	user, err := u.userRepo.GetByID(userID.String())
	if err != nil {
		return nil, err
	}

	var since time.Time
	switch period {
	case "week":
		since = time.Now().AddDate(0, 0, -7)
	case "month":
		since = time.Now().AddDate(0, -1, 0)
	default:
		return nil, errors.New("unsupported period")
	}

	periodStats, err := u.gameRepo.GetUserStatsByPeriod(userID, since, 10)
	if err != nil || periodStats == nil {
		periodStats = &model.UserStatsByPeriod{RecentGames: []model.RecentGameInfo{}}
	}

	winRate := 0.0
	if periodStats.TotalGames > 0 {
		winRate = float64(periodStats.GamesWon) / float64(periodStats.TotalGames) * 100
	}

	gameStats, _ := u.userRepo.GetUserGameStats(userID)
	dayStreak := 0
	bestDayStreak := 0
	if gameStats != nil {
		dayStreak = gameStats.DayStreak
		bestDayStreak = gameStats.BestDayStreak
	}

	return &model.UserStats{
		TotalGames:     periodStats.TotalGames,
		GamesWon:       periodStats.GamesWon,
		GamesLost:      periodStats.GamesLost,
		GamesDrawn:     periodStats.GamesDrawn,
		WinRate:        winRate,
		DayStreak:      dayStreak,
		BestDayStreak:  bestDayStreak,
		TotalScore:     periodStats.TotalScore,
		AverageScore:   periodStats.AverageScore,
		PlayTime:       periodStats.PlayTime,
		EloRating:      user.EloRating,
		EloGamesPlayed: user.EloGamesPlayed,
		Experience:     user.Experience,
		Level:          user.Level,
		Rank:           user.Rank,
		GamesByMode:    []model.GameModeStats{},
		RecentGames:    periodStats.RecentGames,
	}, nil
}

func (u *GameUsecase) CountGameHistory(userID uuid.UUID, status, mode string) (int64, error) {
	return u.gameRepo.CountUserGameHistory(userID, status, mode)
}

func (u *GameUsecase) GetGameHistory(userID uuid.UUID, limit, offset int, status, mode string) ([]model.GameHistoryResponse, error) {
	games, err := u.gameRepo.GetUserGameHistory(userID, limit, offset, status, mode)
	if err != nil {
		return nil, err
	}

	history := make([]model.GameHistoryResponse, 0, len(games))
	for _, g := range games {
		var userScore, opponentScore int
		isWinner := false

		for _, p := range g.Players {
			if p.UserID == userID {
				userScore = p.Score
				if g.WinnerID != nil && *g.WinnerID == userID {
					isWinner = true
				}
			} else {
				opponentScore = p.Score
			}
		}

		history = append(history, model.GameHistoryResponse{
			Game:          &g,
			Players:       g.Players,
			UserScore:     userScore,
			OpponentScore: opponentScore,
			IsWinner:      isWinner,
		})
	}

	return history, nil
}

func (u *GameUsecase) CancelGame(c *gin.Context, gameID, userID uuid.UUID) error {
	var gameModel *model.Game

	err := u.gameRepo.WithTransaction(func(txRepo repository.GameRepositoryInterface) error {
		var err error
		gameModel, err = txRepo.GetGameByIDSimple(gameID)
		if err != nil {
			return errors.New("game not found")
		}

		if gameModel.Status == model.GameStatusCompleted || gameModel.Status == model.GameStatusCancelled {
			return errors.New("game is already finished")
		}

		gameModel.Status = model.GameStatusCancelled
		gameModel.UpdatedAt = time.Now()
		return txRepo.SaveGame(gameModel)
	})
	if err != nil {
		return err
	}

	var winnerPublicId string
	var playersFinalData []map[string]interface{}
	if gameEngine, engineErr := u.gameManager.GetGame(gameID); engineErr == nil {
		enginePlayers := gameEngine.GetPlayers()
		for _, p := range enginePlayers {
			playersFinalData = append(playersFinalData, map[string]interface{}{
				keyUserPublicID: p.PublicID,
				"username":       p.Username,
				keyScore:          p.Score,
			})
			if p.UserID != userID && p.PublicID != "" {
				winnerPublicId = p.PublicID
			}
		}
		if gameModel.Mode == model.GameModeSolo {
			winnerPublicId = ""
		}
	}

	if err := u.gameManager.CancelGame(gameID); err != nil {
		u.logger.Warn("Failed to cancel game in manager", zap.Error(err))
	}

	players, _ := u.gameRepo.GetGamePlayers(gameID)
	for _, p := range players {
		if err := u.userRepo.UpdateUserGameStatus(p.UserID, nil); err != nil {
			u.logger.Warn("Failed to clear user game status", zap.Error(err))
		}
	}

	u.emitEvent(gameID, "game_cancelled", map[string]interface{}{
		"reason":           "player_quit",
		keyWinnerPublicID: winnerPublicId,
		"players_final":    playersFinalData,
	})

	if u.wsService != nil {
		_ = u.wsService.SendToGame(gameModel.PublicID, map[string]interface{}{
			keyType:      "game_cancelled",
			keyGameID:   gameID.String(),
			keyPublicID: gameModel.PublicID,
			keyData: map[string]interface{}{
				"reason":           "player_quit",
				keyWinnerPublicID: winnerPublicId,
				"players_final":    playersFinalData,
			},
		})
	}

	if pendingInvites, err := u.gameRepo.GetPendingInvitesByGameID(gameID); err == nil {
		for _, inv := range pendingInvites {
			_ = u.gameRepo.UpdateGameInviteStatus(inv.ID, model.GameInviteStatusCancelled)
			if u.wsService != nil {
				invCopy := inv
				go func() {
					_ = u.wsService.SendToUser(invCopy.ToUserID, map[string]interface{}{
						keyType: "game_invite_cancelled",
						keyData: map[string]interface{}{
							keyInviteID:      invCopy.ID.String(),
							keyGamePublicID: gameModel.PublicID,
						},
					})
				}()
			}
		}
	}

	_ = u.loggingService.LogUserAction(userID, "cancel_game", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{
		keyGameID: gameID,
	}, true, nil)

	return nil
}

func (u *GameUsecase) selectRandomQuestions(datasetID *uuid.UUID, count int, qtype string) ([]*model.Question, error) {
	var targetDatasetID uuid.UUID

	if datasetID != nil {
		targetDatasetID = *datasetID
		_, err := u.datasetReader.GetDatasetByID("questions", targetDatasetID)
		if err != nil {
			return nil, fmt.Errorf("invalid dataset: %w", err)
		}
	} else {
		defaultDataset, err := u.datasetReader.GetDefaultDataset("questions")
		if err != nil {
			u.logger.Warn("No default dataset found, falling back to all questions", zap.Error(err))
			return u.selectRandomQuestionsFromAll(count)
		}
		targetDatasetID = defaultDataset.ID
	}

	total, err := u.questionRepo.CountByDataset(targetDatasetID, qtype)
	if err != nil {
		u.logger.Error("Failed to count questions in dataset", zap.String("dataset_id", targetDatasetID.String()), zap.Error(err))
		return nil, err
	}

	if total < int64(count) {
		return nil, fmt.Errorf("not enough questions in dataset (%d available, %d requested)", total, count)
	}

	questions, err := u.questionRepo.ListRandomQuestionsByDataset(&targetDatasetID, count, qtype)
	if err != nil {
		return nil, err
	}

	return questions, nil
}

func (u *GameUsecase) selectRandomQuestionsFromAll(count int) ([]*model.Question, error) {
	total, err := u.questionRepo.GetTotalCount()
	if err != nil {
		return nil, err
	}

	if total < int64(count) {
		return nil, errors.New("not enough questions in database")
	}

	questions, err := u.questionRepo.ListRandomQuestionsByDataset(nil, count, "")
	if err != nil {
		return nil, err
	}

	return questions, nil
}

func trimAnswersToTwo(q *model.Question) {
	var answers []model.Answer
	if err := json.Unmarshal(q.Answers, &answers); err != nil || len(answers) <= 2 {
		return
	}
	var correct, wrong []model.Answer
	for _, a := range answers {
		if a.IsCorrect {
			correct = append(correct, a)
		} else {
			wrong = append(wrong, a)
		}
	}
	if len(correct) == 0 || len(wrong) == 0 {
		return
	}
	picked := []model.Answer{
		correct[rand.Intn(len(correct))],
		wrong[rand.Intn(len(wrong))],
	}
	rand.Shuffle(len(picked), func(i, j int) { picked[i], picked[j] = picked[j], picked[i] })
	if b, err := json.Marshal(picked); err == nil {
		q.Answers = datatypes.JSON(b)
	}
}

func (u *GameUsecase) GetGameByPublicID(publicID string) (*model.Game, error) {
	return u.gameRepo.GetGameByPublicID(publicID)
}

func (u *GameUsecase) validateGameCanBeModified(gameModel *model.Game) error {
	if gameModel.Status != model.GameStatusWaiting && gameModel.Status != model.GameStatusReady {
		return errors.New("game already started or finished")
	}
	return nil
}

func (u *GameUsecase) emitEvent(gameID uuid.UUID, eventType string, data map[string]interface{}) {
	u.logger.Info("Game event emitted", zap.String(keyGameID, gameID.String()), zap.String("event", eventType), zap.Any(keyData, data))
}

func (u *GameUsecase) emitGameCompletedEvent(gameModel *model.Game, players []game.Player, winnerID *uuid.UUID) {
	if u.wsService == nil {
		return
	}

	playersFinal := make([]map[string]interface{}, 0, len(players))
	for _, p := range players {
		playersFinal = append(playersFinal, map[string]interface{}{
			keyUserPublicID: p.PublicID,
			"username":       p.Username,
			keyScore:          p.Score,
		})
	}

	eventData := map[string]interface{}{
		keyGameID:          gameModel.ID,
		"players_final":    playersFinal,
	}
	if winnerID != nil {
		for _, p := range players {
			if p.UserID == *winnerID {
				eventData[keyWinnerPublicID] = p.PublicID
				break
			}
		}
	}

	_ = u.wsService.SendToGame(gameModel.PublicID, map[string]interface{}{
		keyType:      "game_completed",
		keyPublicID: gameModel.PublicID,
		keyData:      eventData,
	})
}

func (u *GameUsecase) HandleSubmitAnswer(userID uuid.UUID, gamePublicID string, answerData map[string]interface{}) error {
	gameModel, err := u.gameRepo.GetGameByPublicID(gamePublicID)
	if err != nil {
		return errors.New("game not found")
	}

	gameEngine, err := u.gameManager.GetGame(gameModel.ID)
	if err != nil {
		return errors.New("game not active")
	}

	currentQuestion, err := gameEngine.GetCurrentQuestion()
	if err != nil {
		return err
	}

	var answer game.Answer

	timeSpent := time.Duration(0)
	if tsPtr, ok := answerData["time_spent"].(*int64); ok && tsPtr != nil {
		timeSpent = time.Duration(*tsPtr)
	} else if tsFloat, ok := answerData["time_spent"].(float64); ok {
		timeSpent = time.Duration(tsFloat) * time.Millisecond
	}

	payloadData := answerData
	if inner, ok := answerData["answer"].(map[string]interface{}); ok {
		payloadData = inner
	}

	switch currentQuestion.QType {
	case model.QTypeTextInput:
		text := ""
		if t, ok := payloadData["text"].(string); ok {
			text = t
		} else if t, ok := payloadData["value"].(string); ok {
			text = t
		}
		answer = game.NewTextAnswer(currentQuestion.ID, text, timeSpent)

	case model.QTypeTrueFalse, "true_false":
		boolVal := false
		if v, ok := payloadData["value"].(bool); ok {
			boolVal = v
		} else if v, ok := payloadData["value"].(string); ok {
			boolVal = v == model.AnswerSlugTrue
		} else if v, ok := payloadData["answer_slug"].(string); ok {
			boolVal = v == model.AnswerSlugTrue
		}
		answer = game.NewBooleanAnswer(currentQuestion.ID, boolVal, timeSpent)

	default:
		answerSlug, _ := payloadData["answer_slug"].(string)
		if answerSlug == "" {
			if slug, ok := payloadData["value"].(string); ok {
				answerSlug = slug
			} else if idx, ok := payloadData["value"].(float64); ok {
				answerSlug = fmt.Sprintf("%d", int(idx))
			} else if slug, ok := payloadData[keySlug].(string); ok {
				answerSlug = slug
			}
		}
		answer = game.NewMCQAnswer(currentQuestion.ID, answerSlug, timeSpent)
	}

	if err := u.gameManager.SubmitAnswer(gameModel.ID, userID, answer); err != nil {
		return err
	}

	if freshEngine, loadErr := u.gameManager.GetGame(gameModel.ID); loadErr == nil {
		gameEngine = freshEngine
	}

	player, err := gameEngine.GetPlayer(userID)
	if err == nil && len(player.Answers) > 0 {
		lastAnswer := player.Answers[len(player.Answers)-1]

		players, _ := u.gameRepo.GetGamePlayers(gameModel.ID)
		var gamePlayerID uuid.UUID
		for _, p := range players {
			if p.UserID == userID {
				gamePlayerID = p.ID
				break
			}
		}

		var wsGameQuestionID *uuid.UUID
		if gq, gqErr := u.gameRepo.GetGameQuestionByOrder(gameModel.ID, len(player.Answers)); gqErr == nil && gq != nil {
			wsGameQuestionID = &gq.ID
		}

		wsRichData := map[string]interface{}{
			"question_slug":        currentQuestion.Slug,
			"question_type":        currentQuestion.QType,
			"user_answer":          answer.AnswerData,
			"server_time_spent_ms": lastAnswer.ServerTimeSpent.Milliseconds(),
		}
		for k, v := range lastAnswer.Metadata {
			wsRichData[k] = v
		}
		wsRichDataBytes, _ := json.Marshal(wsRichData)

		wsAnswerSlug := lastAnswer.GetAnswerSlug()

		gameAnswer := &model.GameAnswer{
			GameID:         gameModel.ID,
			PlayerID:       gamePlayerID,
			QuestionID:     &currentQuestion.ID,
			GameQuestionID: wsGameQuestionID,
			AnswerSlug:     wsAnswerSlug,
			Data:           datatypes.JSON(wsRichDataBytes),
			IsCorrect:      lastAnswer.IsCorrect,
			TimeSpent:      int(timeSpent.Milliseconds()),
			Points:         lastAnswer.Points,
			AnsweredAt:     time.Now(),
		}

		if err := u.gameRepo.SaveGameAnswer(gameAnswer); err != nil {
			u.logger.Warn("Failed to save answer to database", zap.Error(err))
		}

		u.logger.Info("Answer recorded (WS)",
			zap.String(keyGameID, gameModel.ID.String()),
			zap.String(keyUserID, userID.String()),
			zap.String("question_id", currentQuestion.ID.String()),
			zap.String("question_slug", currentQuestion.Slug),
			zap.String("question_type", currentQuestion.QType),
			zap.String("answer_slug", wsAnswerSlug),
			zap.Bool("is_correct", lastAnswer.IsCorrect),
			zap.Int("points", lastAnswer.Points),
			zap.Int64("time_spent_ms", timeSpent.Milliseconds()),
			zap.Int64("server_time_spent_ms", lastAnswer.ServerTimeSpent.Milliseconds()),
			zap.Any("metadata", lastAnswer.Metadata),
		)

		if err := u.gameRepo.UpdatePlayerScore(gamePlayerID, player.Score); err != nil {
			u.logger.Warn("Failed to update player score", zap.Error(err))
		}
	}

	if gameEngine.AllPlayersFinished() {
		if err := u.finalizeGame(nil, gameModel.ID); err != nil {
			u.logger.Error("Failed to finalize game via WebSocket", zap.Error(err))
		}
	}

	u.logger.Debug("Answer submitted via WebSocket",
		zap.String(keyGamePublicID, gamePublicID),
		zap.String(keyUserID, userID.String()),
	)

	return nil
}

func (u *GameUsecase) HandlePlayerReady(userID uuid.UUID, gamePublicID string, ready bool) error {
	gameModel, err := u.gameRepo.GetGameByPublicID(gamePublicID)
	if err != nil {
		return errors.New("game not found")
	}

	if err := u.gameManager.SetPlayerReady(gameModel.ID, userID, ready); err != nil {
		return err
	}

	u.logger.Debug("Player ready status updated via WebSocket",
		zap.String(keyGamePublicID, gamePublicID),
		zap.String(keyUserID, userID.String()),
		zap.Bool("ready", ready),
	)

	return nil
}

func (u *GameUsecase) HandleStartGame(userID uuid.UUID, gamePublicID string) error {
	gameModel, err := u.gameRepo.GetGameByPublicID(gamePublicID)
	if err != nil {
		return errors.New("game not found")
	}

	if gameModel.CreatorID != userID {
		return errors.New("only the game creator can start the game")
	}

	if err := u.gameManager.StartGame(gameModel.ID); err != nil {
		return err
	}

	now := time.Now()
	gameModel.Status = model.GameStatusInProgress
	gameModel.StartedAt = &now
	if err := u.gameRepo.UpdateGame(gameModel); err != nil {
		u.logger.Warn("Failed to update game in database", zap.Error(err))
	}

	u.logger.Info("Game started via WebSocket",
		zap.String(keyGamePublicID, gamePublicID),
		zap.String(keyUserID, userID.String()),
	)

	return nil
}

func (u *GameUsecase) HandleLeaveGame(userID uuid.UUID, gamePublicID string) error {
	gameModel, err := u.gameRepo.GetGameByPublicID(gamePublicID)
	if err != nil {
		return errors.New("game not found")
	}

	return u.LeaveGame(nil, gameModel.ID, userID)
}

func (u *GameUsecase) MarkPlayerDisconnected(userID uuid.UUID, gamePublicID string) error {
	gameModel, err := u.gameRepo.GetGameByPublicID(gamePublicID)
	if err != nil {
		_ = err
		return nil
	}
	return u.gameManager.MarkPlayerDisconnected(gameModel.ID, userID)
}

func (u *GameUsecase) MarkPlayerReconnected(userID uuid.UUID, gamePublicID string) error {
	gameModel, err := u.gameRepo.GetGameByPublicID(gamePublicID)
	if err != nil {
		_ = err
		return nil
	}
	return u.gameManager.MarkPlayerReconnected(gameModel.ID, userID)
}

func (u *GameUsecase) OnUserConnected(userID uuid.UUID) {
	if u.wsService == nil {
		return
	}
	invites, err := u.gameRepo.GetUserGameInvites(userID, model.GameInviteStatusPending)
	if err != nil || len(invites) == 0 {
		return
	}

	payload := make([]map[string]interface{}, 0, len(invites))
	for _, inv := range invites {
		entry := map[string]interface{}{
			keyInviteID:      inv.ID.String(),
			keyGameID:        inv.GameID.String(),
			keyFromPublicID: inv.FromUserPublicID,
		}
		if gameModel, err := u.gameRepo.GetGameByID(inv.GameID); err == nil {
			entry[keyGamePublicID] = gameModel.PublicID
			entry["game_mode"] = string(gameModel.Mode)
		}
		if fromUser, err := u.userRepo.GetByID(inv.FromUserID.String()); err == nil {
			entry["from_username"] = fromUser.Username
			entry[keyFromPublicID] = fromUser.PublicID
		}
		payload = append(payload, entry)
	}

	_ = u.wsService.SendToUser(userID, map[string]interface{}{
		keyType: "pending_invites",
		keyData: payload,
	})
}

func (u *GameUsecase) GetCurrentQuestionPayload(gamePublicID string) (map[string]interface{}, error) {
	gameModel, err := u.gameRepo.GetGameByPublicID(gamePublicID)
	if err != nil {
		_ = err
		return nil, nil
	}

	if gameModel.Status != model.GameStatusInProgress {
		return nil, nil
	}

	gameEngine, err := u.gameManager.GetGame(gameModel.ID)
	if err != nil {
		_ = err
		return nil, nil
	}

	if gameEngine.GetStatus() != model.GameStatusInProgress {
		return nil, nil
	}

	question, err := gameEngine.GetCurrentQuestion()
	if err != nil || question == nil {
		_ = err
		return nil, nil
	}

	elapsedSeconds := 0.0
	for _, p := range gameEngine.GetPlayers() {
		if !p.CurrentQuestionSentAt.IsZero() {
			elapsedSeconds = time.Since(p.CurrentQuestionSentAt).Seconds()
			break
		}
	}
	remainingSeconds := question.EstimatedSeconds - int(elapsedSeconds)
	if remainingSeconds < 0 {
		remainingSeconds = 0
	}

	return map[string]interface{}{
		"question":          game.QuestionToPayload(question),
		"question_number":   gameEngine.GetQuestionNumber(),
		"total_questions":   gameEngine.GetTotalQuestions(),
		"time_limit":        question.EstimatedSeconds,
		"remaining_seconds": remainingSeconds,
	}, nil
}

func (u *GameUsecase) GetGameStateForReconnect(gamePublicID string) (map[string]interface{}, error) {
	gameModel, err := u.gameRepo.GetGameByPublicID(gamePublicID)
	if err != nil {
		_ = err
		return nil, nil
	}

	if gameModel.Status != model.GameStatusInProgress {
		return nil, nil
	}

	gameEngine, err := u.gameManager.GetGame(gameModel.ID)
	if err != nil {
		_ = err
		return nil, nil
	}

	if gameEngine.GetStatus() != model.GameStatusInProgress {
		return nil, nil
	}

	players := gameEngine.GetPlayers()
	playerSnapshots := make([]map[string]interface{}, 0, len(players))
	for _, p := range players {
		playerSnapshots = append(playerSnapshots, map[string]interface{}{
			keyUserPublicID: p.PublicID,
			"username":       p.Username,
			keyScore:          p.Score,
		})
	}

	state := map[string]interface{}{
		"players":         playerSnapshots,
		"question_number": gameEngine.GetQuestionNumber(),
		"total_questions": gameEngine.GetTotalQuestions(),
	}

	if gameEngine.GetPaused() {
		state["paused"] = true
		if deadline := gameEngine.GetReconnectDeadline(); deadline != nil {
			remaining := time.Until(*deadline)
			if remaining < 0 {
				remaining = 0
			}
			state["reconnect_countdown_remaining_secs"] = int(remaining.Seconds())
		}
	}

	return state, nil
}

func (u *GameUsecase) ListActiveTemplates() ([]model.GameTemplate, error) {
	t := true
	return u.templateRepo.List(model.GameTemplateListParams{IsActive: &t})
}

func (u *GameUsecase) updateQuestionStats(gameID uuid.UUID) error {
	answers, err := u.gameRepo.GetGameAnswers(gameID)
	if err != nil {
		u.logger.Warn("Failed to get game answers for stats update", zap.Error(err))
		return err
	}

	for _, answer := range answers {
		if answer.QuestionID != nil {
			if err := u.questionRepo.IncrementQuestionStats(*answer.QuestionID, answer.IsCorrect, answer.TimeSpent); err != nil {
				u.logger.Warn("Failed to update question stats",
					zap.String("question_id", answer.QuestionID.String()),
					zap.Error(err),
				)
			}
		}
	}
	return nil
}

func (u *GameUsecase) CancelUserGameInvite(c *gin.Context, inviteID, userID uuid.UUID) error {
	invite, err := u.gameRepo.GetGameInviteByID(inviteID)
	if err != nil {
		return errors.New("invite not found")
	}
	if invite.FromUserID != userID {
		return errors.New("forbidden")
	}
	if invite.Status != model.GameInviteStatusPending {
		return errors.New("invite is not pending")
	}
	if err := u.gameRepo.UpdateGameInviteStatus(inviteID, model.GameInviteStatusCancelled); err != nil {
		return err
	}
	if u.wsService != nil {
		go func() {
			_ = u.wsService.SendToUser(invite.ToUserID, map[string]interface{}{
				keyType:      "game_invite_cancelled",
				keyInviteID: inviteID.String(),
			})
		}()
	}
	if u.notifRepo != nil {
		_ = u.notifRepo.DeleteGameInviteNotification(invite.ToUserID, inviteID)
	}
	return nil
}

func (u *GameUsecase) GetGameResults(publicID string, userID uuid.UUID) (*model.GameResultsResponse, error) {
	g, err := u.gameRepo.GetGameByPublicID(publicID)
	if err != nil {
		return nil, errors.New("game not found")
	}
	if g.Status != model.GameStatusCompleted {
		return nil, errors.New("game is not completed")
	}

	isPlayer := false
	for _, p := range g.Players {
		if p.UserID == userID {
			isPlayer = true
			break
		}
	}
	if !isPlayer {
		return nil, errors.New("forbidden")
	}

	var winnerPublicID *string
	if g.WinnerID != nil {
		for _, p := range g.Players {
			if p.User != nil && p.UserID == *g.WinnerID {
				id := p.User.PublicID
				winnerPublicID = &id
				break
			}
		}
	}

	gameSummary := model.GameSummary{
		PublicID:       g.PublicID,
		Mode:           string(g.Mode),
		Status:         string(g.Status),
		StartedAt:      g.StartedAt,
		CompletedAt:    g.CompletedAt,
		WinnerPublicID: winnerPublicID,
	}

	players := make([]model.PlayerResultSummary, 0, len(g.Players))
	for _, p := range g.Players {
		pubID := ""
		username := ""
		if p.User != nil {
			pubID = p.User.PublicID
			username = p.User.Username
		}
		players = append(players, model.PlayerResultSummary{
			PublicID: pubID,
			Username: username,
			Score:    p.Score,
			IsWinner: g.WinnerID != nil && p.UserID == *g.WinnerID,
		})
	}

	// Index answers by GameQuestion.ID
	answersByGQID := make(map[uuid.UUID][]model.GameAnswer)
	for _, a := range g.Answers {
		if a.GameQuestionID != nil {
			answersByGQID[*a.GameQuestionID] = append(answersByGQID[*a.GameQuestionID], a)
		}
	}

	questions := make([]model.QuestionResult, 0, len(g.Questions))
	for _, gq := range g.Questions {
		title := ""
		qtype := gq.Type
		correctAnswer := ""

		if gq.Question != nil {
			var i18n map[string]model.QuestionI18n
			if err := json.Unmarshal(gq.Question.I18n, &i18n); err == nil {
				if en, ok := i18n["en"]; ok {
					title = en.Title
				} else {
					for _, v := range i18n {
						title = v.Title
						break
					}
				}
			}
			var answers []model.Answer
			if err := json.Unmarshal(gq.Question.Answers, &answers); err == nil {
				for _, a := range answers {
					if a.IsCorrect {
						correctAnswer = a.Slug
						break
					}
				}
			}
			qtype = gq.Question.QType
		}

		gaList := answersByGQID[gq.ID]
		playerAnswers := make([]model.PlayerAnswerResult, 0, len(gaList))
		for _, ga := range gaList {
			playerPubID := ""
			for _, p := range g.Players {
				if p.ID == ga.PlayerID {
					if p.User != nil {
						playerPubID = p.User.PublicID
					}
					break
				}
			}
			playerAnswers = append(playerAnswers, model.PlayerAnswerResult{
				PlayerPublicID: playerPubID,
				Answer:         ga.AnswerSlug,
				IsCorrect:      ga.IsCorrect,
				Points:         ga.Points,
				TimeSpentMs:    ga.TimeSpent,
			})
		}

		questions = append(questions, model.QuestionResult{
			OrderNumber:   gq.OrderNumber,
			QuestionTitle: title,
			QuestionType:  qtype,
			CorrectAnswer: correctAnswer,
			Answers:       playerAnswers,
		})
	}

	return &model.GameResultsResponse{
		Game:      gameSummary,
		Players:   players,
		Questions: questions,
	}, nil
}

func (u *GameUsecase) JoinGameByCode(c *gin.Context, code string, userID uuid.UUID) error {
	g, err := u.gameRepo.GetGameByPublicID(code)
	if err != nil {
		return errors.New("game not found")
	}

	if g.Mode != model.GameMode1v1 && g.Mode != model.GameModeMulti {
		return errors.New("join by code is only available for 1v1 and multi games")
	}
	if g.Status != model.GameStatusWaiting {
		return errors.New("game is not accepting players")
	}

	for _, p := range g.Players {
		if p.UserID == userID {
			return errors.New("you are already in this game")
		}
	}

	if len(g.Players) >= g.MaxPlayers {
		return errors.New("game is full")
	}

	if isBlocked, err := u.friendsRepo.IsBlocked(userID, g.CreatorID); err == nil && isBlocked {
		return errors.New("you cannot join this game")
	}

	gamePlayer := &model.GamePlayer{
		GameID:   g.ID,
		UserID:   userID,
		Score:    0,
		IsReady:  false,
		JoinedAt: time.Now(),
	}
	if err := u.gameRepo.AddPlayerToGame(gamePlayer); err != nil {
		return err
	}
	if err := u.gameManager.AddPlayerToGame(g.ID, userID); err != nil {
		return err
	}
	if err := u.userRepo.UpdateUserGameStatus(userID, &g.ID); err != nil {
		u.logger.Warn("Failed to update user game status", zap.Error(err))
	}

	if u.wsService != nil {
		go func() {
			playerData := map[string]interface{}{keyAction: "joined"}
			if joiner, err := u.userRepo.GetByID(userID.String()); err == nil {
				playerData[keyUserPublicID] = joiner.PublicID
				playerData["username"] = joiner.Username
			}
			_ = u.wsService.SendToGame(g.PublicID, map[string]interface{}{
				keyType:           msgLobbyUpdated,
				keyGamePublicID: g.PublicID,
				keyData:           playerData,
			})
		}()
	}

	_ = u.loggingService.LogUserAction(userID, "join_game_by_code", httputil.GetRealIP(c), c.Request.UserAgent(), map[string]interface{}{
		keyGamePublicID: code,
	}, true, nil)

	return nil
}
