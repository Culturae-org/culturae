// backend/internal/game/versus.go

package game

import (
	"time"

	"github.com/Culturae-org/culturae/internal/model"
	adminRepo "github.com/Culturae-org/culturae/internal/repository/admin"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type dynamicHooks struct {
	settings GameSettings
}

func (h *dynamicHooks) maxPlayers() int {
	if h.settings.MaxPlayers > 0 {
		return h.settings.MaxPlayers
	}
	return 1
}

func (h *dynamicHooks) minPlayers() int {
	if h.settings.MinPlayers > 0 {
		return h.settings.MinPlayers
	}
	return 1
}

func (h *dynamicHooks) MaxPlayers() int {
	return h.maxPlayers()
}

func (h *dynamicHooks) MinPlayersToStart() int {
	return h.minPlayers()
}

func (h *dynamicHooks) IsReadyToStart(playerCount int) bool {
	return playerCount >= h.minPlayers()
}

func (h *dynamicHooks) ShouldAdvanceQuestion(base *BaseGame) bool {
	if base.paused {
		return false
	}
	currentQ := base.GetCurrentQuestionIndex()
	for _, p := range base.players {
		if p.Status == model.PlayerStatusLeft || p.Status == model.PlayerStatusDisconnected {
			continue
		}
		if len(p.Answers) <= currentQ {
			return false
		}
	}
	return true
}

func (h *dynamicHooks) EvaluateQuestionResults(base *BaseGame, questionIndex int) {
	scoreMode := h.settings.ScoreMode
	if scoreMode == "" {
		scoreMode = ScoreModeTimeBonus
	}

	switch scoreMode {
	case ScoreModeFastestWins:
		h.evaluateFastestWins(base, questionIndex)
	case "classic":
		h.evaluateClassic(base, questionIndex)
	default:
		h.evaluateTimeBonus(base, questionIndex)
	}
}

func (h *dynamicHooks) evaluateClassic(base *BaseGame, questionIndex int) {
	basePoints := h.settings.PointsPerCorrect
	if basePoints <= 0 {
		basePoints = 100
	}
	for _, player := range base.players {
		if len(player.Answers) <= questionIndex {
			continue
		}
		answer := player.Answers[questionIndex]
		if !answer.IsCorrect {
			continue
		}
		player.Answers[questionIndex].Points = basePoints
		player.Score += basePoints
	}
}

func (h *dynamicHooks) evaluateTimeBonus(base *BaseGame, questionIndex int) {
	settings := base.GetSettings()
	question := base.GetQuestionByIndex(questionIndex)

	basePoints := h.settings.PointsPerCorrect
	if basePoints <= 0 {
		basePoints = 100
	}

	for _, player := range base.players {
		if len(player.Answers) <= questionIndex {
			continue
		}
		answer := player.Answers[questionIndex]
		if !answer.IsCorrect {
			continue
		}

		finalPoints := basePoints

		if settings.TimeBonus && question != nil && question.EstimatedSeconds > 0 &&
			answer.ServerTimeSpent > 0 {
			timeLimit := time.Duration(question.EstimatedSeconds) * time.Second
			if answer.ServerTimeSpent < timeLimit {
				bonusFraction := 1.0 - float64(answer.ServerTimeSpent)/float64(timeLimit)
				finalPoints += int(float64(basePoints) * 0.5 * bonusFraction)
			}
		}

		player.Answers[questionIndex].Points = finalPoints
		player.Score += finalPoints
	}
}

func (h *dynamicHooks) evaluateFastestWins(base *BaseGame, questionIndex int) {
	question := base.GetQuestionByIndex(questionIndex)
	if question == nil {
		return
	}
	questionPoints := h.settings.PointsPerCorrect
	if questionPoints <= 0 {
		questionPoints = 100
	}

	players := base.players

	type correctEntry struct {
		playerID  uuid.UUID
		timeSpent time.Duration
	}
	var correct []correctEntry

	for userID, player := range players {
		if len(player.Answers) > questionIndex {
			ans := player.Answers[questionIndex]
			if ans.IsCorrect {
				correct = append(correct, correctEntry{userID, ans.ServerTimeSpent})
			}
		}
	}

	if len(correct) == 1 {
		p := players[correct[0].playerID]
		p.Score += questionPoints
		p.Answers[questionIndex].Points = questionPoints
	} else if len(correct) > 1 {
		fastest := correct[0]
		for _, c := range correct[1:] {
			if c.timeSpent < fastest.timeSpent {
				fastest = c
			}
		}
		for _, c := range correct {
			p := players[c.playerID]
			if c.playerID == fastest.playerID {
				p.Score += questionPoints
				p.Answers[questionIndex].Points = questionPoints
			} else {
				p.Answers[questionIndex].Points = 0
			}
		}
	}
}

type VersusGame struct {
	*BaseGame
}

func NewVersusGame(
	gameID uuid.UUID,
	publicID string,
	mode model.GameMode,
	settings GameSettings,
	questions []*model.Question,
	logger *zap.Logger,
	repo adminRepo.AdminLogsRepositoryInterface,
	countdownConfig *model.CountdownConfig,
) *VersusGame {
	hooks := &dynamicHooks{settings: settings}
	base := NewBaseGame(
		gameID,
		publicID,
		mode,
		settings,
		questions,
		logger,
		hooks,
		repo,
		countdownConfig,
	)
	return &VersusGame{BaseGame: base}
}

func (g *VersusGame) GetQuestions() []*model.Question {
	return g.BaseGame.GetQuestions()
}

var _ GameEngine = (*VersusGame)(nil)
