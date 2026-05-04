// backend/culturae/internal/game/base.go

package game

import (
	"errors"
	"sync"
	"time"

	"github.com/Culturae-org/culturae/internal/game/validator"
	"github.com/Culturae-org/culturae/internal/model"
	adminRepo "github.com/Culturae-org/culturae/internal/repository/admin"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type GameModeHooks interface {
	MaxPlayers() int

	MinPlayersToStart() int

	IsReadyToStart(playerCount int) bool

	ShouldAdvanceQuestion(base *BaseGame) bool

	EvaluateQuestionResults(base *BaseGame, questionIndex int)
}

type BaseGame struct {
	id          uuid.UUID
	publicID    string
	mode        model.GameMode
	status      model.GameStatus
	players     map[uuid.UUID]*Player
	settings    GameSettings
	questions   []*model.Question
	currentQ    int
	winnerID    *uuid.UUID
	startedAt   *time.Time
	completedAt *time.Time
	mutex       sync.RWMutex

	eventChan     chan GameEvent
	commandChan   chan GameCommand
	stopChan      chan struct{}
	ticker        *time.Ticker
	running       bool
	pendingTimers []*time.Timer

	paused             bool
	pausedAt           time.Time
	tickerRemainingMs  int64
	pausedQuestionIdx  int
	reconnectTimer     *time.Timer
	reconnectDeadline  *time.Time

	logger *zap.Logger

	hooks GameModeHooks

	countdownConfig *model.CountdownConfig

	saveCallback func() error
}

func NewBaseGame(
	gameID uuid.UUID,
	publicID string,
	mode model.GameMode,
	settings GameSettings,
	questions []*model.Question,
	logger *zap.Logger,
	hooks GameModeHooks,
	repo adminRepo.AdminLogsRepositoryInterface,
	countdownConfig *model.CountdownConfig,
) *BaseGame {
	if countdownConfig == nil {
		countdownConfig = &model.CountdownConfig{PreGameCountdownSeconds: 3}
	}
	return &BaseGame{
		id:              gameID,
		publicID:        publicID,
		mode:            mode,
		status:          model.GameStatusWaiting,
		players:         make(map[uuid.UUID]*Player),
		settings:        settings,
		questions:       questions,
		currentQ:        0,
		logger:          logger,
		hooks:           hooks,
		countdownConfig: countdownConfig,
	}
}

func (g *BaseGame) GetID() uuid.UUID {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.id
}

func (g *BaseGame) GetPublicID() string {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.publicID
}

func (g *BaseGame) GetMode() model.GameMode {
	return g.mode
}

func (g *BaseGame) GetStatus() model.GameStatus {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.status
}

func (g *BaseGame) GetPlayers() []Player {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	players := make([]Player, 0, len(g.players))
	for _, p := range g.players {
		players = append(players, *p)
	}
	return players
}

func (g *BaseGame) GetSettings() GameSettings {
	return g.settings
}

func (g *BaseGame) GetPlayer(userID uuid.UUID) (*Player, error) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	player, exists := g.players[userID]
	if !exists {
		return nil, errors.New("player not found")
	}

	playerCopy := *player
	return &playerCopy, nil
}

func (g *BaseGame) GetCurrentQuestion() (*model.Question, error) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	if g.currentQ >= len(g.questions) {
		return nil, errors.New("no more questions")
	}

	return g.questions[g.currentQ], nil
}

func (g *BaseGame) GetQuestionNumber() int {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.currentQ + 1
}

func (g *BaseGame) GetTotalQuestions() int {
	return len(g.questions)
}

func (g *BaseGame) GetQuestions() []*model.Question {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.questions
}

func (g *BaseGame) IsFinished() bool {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.currentQ >= len(g.questions) ||
		g.status == model.GameStatusCancelled ||
		g.status == model.GameStatusAbandoned ||
		g.status == model.GameStatusCompleted
}

func (g *BaseGame) AllPlayersFinished() bool {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	if g.status == model.GameStatusCancelled || g.status == model.GameStatusAbandoned || g.status == model.GameStatusCompleted {
		return true
	}

	if g.currentQ >= len(g.questions) {
		if len(g.players) == 0 {
			return true
		}
		for _, p := range g.players {
			if p.Status == model.PlayerStatusLeft || p.Status == model.PlayerStatusDisconnected {
				continue
			}
			if len(p.Answers) < len(g.questions) {
				return false
			}
		}
		return true
	}

	return false
}

func (g *BaseGame) determineWinner() *uuid.UUID {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.determineWinnerLocked()
}

func (g *BaseGame) determineWinnerLocked() *uuid.UUID {
	if g.mode == model.GameModeSolo {
		for _, p := range g.players {
			if p.Status != model.PlayerStatusLeft {
				id := p.UserID
				return &id
			}
		}
		return nil
	}
	
	bestScore := -1
	var winnerID *uuid.UUID
	for _, p := range g.players {
		if p.Status == model.PlayerStatusLeft {
			continue
		}
		if p.Score > bestScore {
			bestScore = p.Score
			id := p.UserID
			winnerID = &id
		} else if p.Score == bestScore {
			winnerID = nil
		}
	}
	return winnerID
}

func (g *BaseGame) GetWinnerID() *uuid.UUID {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	if g.winnerID != nil {
		return g.winnerID
	}
	if g.currentQ >= len(g.questions) ||
		g.status == model.GameStatusCompleted ||
		g.status == model.GameStatusCancelled {
		return g.determineWinnerLocked()
	}
	return nil
}

func (g *BaseGame) GetStartedAt() *time.Time {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.startedAt
}

func (g *BaseGame) GetCompletedAt() *time.Time {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.completedAt
}

func (g *BaseGame) AddPlayer(userID uuid.UUID) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.status != model.GameStatusWaiting && g.status != model.GameStatusReady {
		return errors.New("cannot add player: game already started or finished")
	}

	if len(g.players) >= g.hooks.MaxPlayers() {
		return errors.New("game is full")
	}

	if _, exists := g.players[userID]; exists {
		return errors.New("player already in game")
	}

	g.players[userID] = &Player{
		ID:       uuid.New(),
		UserID:   userID,
		Score:    0,
		IsReady:  false,
		Status:   model.PlayerStatusActive,
		Answers:  make([]Answer, 0),
		JoinedAt: time.Now(),
	}

	if g.hooks.IsReadyToStart(len(g.players)) {
		g.status = model.GameStatusReady
	}

	return nil
}

func (g *BaseGame) RemovePlayer(userID uuid.UUID) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	player, exists := g.players[userID]
	if !exists {
		return errors.New("player not in game")
	}

	player.Status = model.PlayerStatusLeft

	activeCount := 0
	for _, p := range g.players {
		if p.Status != model.PlayerStatusLeft {
			activeCount++
		}
	}

	if g.status == model.GameStatusInProgress {
		if activeCount < g.hooks.MinPlayersToStart() {
			g.finalizeWithWinner()
		} else if g.hooks.ShouldAdvanceQuestion(g) {
			g.hooks.EvaluateQuestionResults(g, g.currentQ)
			g.currentQ++
			if g.currentQ >= len(g.questions) {
				g.endLocked(g.determineWinnerLocked())
			} else {
				questionSentAt := time.Now()
				for _, p := range g.players {
					p.CurrentQuestionSentAt = questionSentAt
				}
			}
		}
	} else {
		if activeCount < g.hooks.MinPlayersToStart() {
			g.status = model.GameStatusWaiting
		}
	}

	return nil
}

func (g *BaseGame) SetPlayerReady(userID uuid.UUID, ready bool) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	player, exists := g.players[userID]
	if !exists {
		return errors.New("player not in game")
	}

	player.IsReady = ready
	return nil
}

func (g *BaseGame) SetPlayerPublicID(userID uuid.UUID, publicID string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	player, exists := g.players[userID]
	if !exists {
		return errors.New("player not in game")
	}

	player.PublicID = publicID
	return nil
}

func (g *BaseGame) SetPlayerUsername(userID uuid.UUID, username string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	player, exists := g.players[userID]
	if !exists {
		return errors.New("player not in game")
	}

	player.Username = username
	return nil
}

func (g *BaseGame) DisconnectPlayer(userID uuid.UUID) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	player, exists := g.players[userID]
	if !exists {
		return errors.New("player not in game")
	}
	if player.Status != model.PlayerStatusActive {
		return nil
	}
	player.Status = model.PlayerStatusDisconnected
	return nil
}

func (g *BaseGame) ReconnectPlayer(userID uuid.UUID) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	player, exists := g.players[userID]
	if !exists {
		return errors.New("player not in game")
	}
	if player.Status != model.PlayerStatusDisconnected {
		return nil
	}
	player.Status = model.PlayerStatusActive
	return nil
}

func (g *BaseGame) CanStart() bool {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	if len(g.players) < g.hooks.MinPlayersToStart() {
		return false
	}

	for _, player := range g.players {
		if !player.IsReady {
			return false
		}
	}

	return g.status == model.GameStatusReady
}

func (g *BaseGame) Start() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.status != model.GameStatusReady {
		return errors.New("game is not ready to start")
	}

	if len(g.questions) == 0 {
		return errors.New("game has no questions")
	}

	activeCount := 0
	for _, p := range g.players {
		if p.Status != model.PlayerStatusLeft {
			activeCount++
		}
	}

	if activeCount < g.hooks.MinPlayersToStart() {
		return errors.New("not enough players to start")
	}

	now := time.Now()
	g.startedAt = &now
	g.status = model.GameStatusInProgress
	g.currentQ = 0

	return nil
}

func (g *BaseGame) End(winnerID *uuid.UUID) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.status != model.GameStatusInProgress {
		return errors.New("game is not in progress")
	}

	g.endLocked(winnerID)
	return nil
}

func (g *BaseGame) endLocked(winnerID *uuid.UUID) {
	now := time.Now()
	g.completedAt = &now
	g.status = model.GameStatusCompleted
	g.winnerID = winnerID
}

func (g *BaseGame) Cancel() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.status == model.GameStatusCompleted {
		return errors.New("cannot cancel completed game")
	}

	g.status = model.GameStatusCancelled
	return nil
}

func (g *BaseGame) finalizeWithWinner() {
	var winner *Player
	for _, p := range g.players {
		if p.Status != model.PlayerStatusLeft {
			winner = p
			break
		}
	}
	if winner != nil {
		g.endLocked(&winner.UserID)
		winnerPublicID := winner.PublicID
		playersFinal := g.buildPlayersFinalLocked()
		g.emitEvent(GameEvent{
			Type: EventGameCompleted,
			Data: map[string]interface{}{
				keyGameID:          g.publicID,
				keyWinnerPublicID: winnerPublicID,
				keyPlayersFinal:    playersFinal,
			},
		})
	} else {
		now := time.Now()
		g.status = model.GameStatusCancelled
		g.completedAt = &now
		g.emitEvent(GameEvent{
			Type: EventGameCancelled,
			Data: map[string]interface{}{
				keyReason: "all_players_left",
			},
		})
	}
}

func (g *BaseGame) SubmitAnswer(userID uuid.UUID, answer Answer) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.status != model.GameStatusInProgress {
		return errors.New("game is not in progress")
	}

	if g.currentQ >= len(g.questions) {
		return errors.New("no more questions")
	}

	player, exists := g.players[userID]
	if !exists {
		return errors.New("player not in game")
	}

	currentQuestion := g.questions[g.currentQ]
	if answer.QuestionID != currentQuestion.ID {
		return errors.New("answer is not for current question")
	}

	hasAlreadyAnswered := len(player.Answers) > g.currentQ
	playerHasLeft := player.Status == model.PlayerStatusLeft

	serverReceivedAt := time.Now()

	if player.CurrentQuestionSentAt.IsZero() {
		return errors.New("server error: question start time not recorded, please retry")
	}

	serverTimeSpent := serverReceivedAt.Sub(player.CurrentQuestionSentAt)
	if serverTimeSpent < MinAnswerTimeMs*time.Millisecond {
		serverTimeSpent = MinAnswerTimeMs * time.Millisecond
	}
	answer.ServerTimeSpent = serverTimeSpent
	answer.ReceivedAt = serverReceivedAt

	if hasAlreadyAnswered {
		return errors.New("already answered this question")
	}
	if playerHasLeft {
		return errors.New("player has left the game")
	}

	v := validator.GetValidator(currentQuestion.QType)
	result := v.Validate(answer, currentQuestion)

	answer.IsCorrect = result.IsCorrect
	answer.Points = result.Score

	if answer.Metadata == nil {
		answer.Metadata = make(map[string]interface{})
	}
	for k, v := range result.Feedback {
		answer.Metadata[k] = v
	}

	player.Answers = append(player.Answers, answer)

	if g.hooks.ShouldAdvanceQuestion(g) {
		g.hooks.EvaluateQuestionResults(g, g.currentQ)
		g.currentQ++
		questionSentAt := time.Now()
		for _, p := range g.players {
			p.CurrentQuestionSentAt = questionSentAt
		}
	}

	return nil
}

func (g *BaseGame) StartGoroutine() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.running {
		return
	}

	g.eventChan = make(chan GameEvent, 64)
	g.commandChan = make(chan GameCommand, 32)
	g.stopChan = make(chan struct{})
	g.running = true

	go g.run()
}

func (g *BaseGame) StopGoroutine() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if !g.running {
		return
	}

	for _, t := range g.pendingTimers {
		t.Stop()
	}
	g.pendingTimers = nil

	if g.reconnectTimer != nil {
		g.reconnectTimer.Stop()
		g.reconnectTimer = nil
	}

	if g.stopChan != nil {
		close(g.stopChan)
		g.stopChan = nil
	}
	g.running = false

	if g.ticker != nil {
		g.ticker.Stop()
		g.ticker = nil
	}
}

func (g *BaseGame) SendCommand(cmd GameCommand) {
	g.mutex.RLock()
	running := g.running
	commandChan := g.commandChan
	g.mutex.RUnlock()

	if !running {
		return
	}
	select {
	case commandChan <- cmd:
	default:
		g.logger.Warn("Command channel full, dropping command",
			zap.String("command_type", cmd.Type),
			zap.String(keyGameID, g.id.String()),
		)
	}
}

func (g *BaseGame) Events() <-chan GameEvent {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.eventChan
}

func (g *BaseGame) SetSaveCallback(callback func() error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.saveCallback = callback
}

func (g *BaseGame) run() {
	g.mutex.RLock()
	commandChan := g.commandChan
	stopChan := g.stopChan
	g.mutex.RUnlock()

	defer func() {
		g.mutex.Lock()
		g.running = false
		if g.eventChan != nil {
			close(g.eventChan)
		}
		g.mutex.Unlock()
	}()

	for {
		g.mutex.RLock()
		var tickerC <-chan time.Time
		if g.ticker != nil {
			tickerC = g.ticker.C
		}
		saveCallback := g.saveCallback
		g.mutex.RUnlock()

		select {
		case cmd := <-commandChan:
			g.handleCommand(cmd)
			if saveCallback != nil {
				if err := saveCallback(); err != nil {
					g.logger.Error("Failed to save game after command",
						zap.String("command_type", cmd.Type),
						zap.Error(err),
					)
				}
			}

		case <-stopChan:
			return

		case <-tickerC:
			if tickerC != nil {
				g.handleTimer()
				if saveCallback != nil {
					if err := saveCallback(); err != nil {
						g.logger.Error("Failed to save game after timer", zap.Error(err))
					}
				}
			}
		}
	}
}

func (g *BaseGame) handleCommand(cmd GameCommand) {
	switch cmd.Type {
	case CmdStartGame:
		g.handleStartGame()
	case CmdEndGame:
		g.handleEndGame(cmd.Payload)
	case CmdSubmitAnswer:
		g.handleSubmitAnswer(cmd.Payload)
	case CmdAddPlayer:
		g.handleAddPlayer(cmd.Payload)
	case CmdRemovePlayer:
		g.handleRemovePlayer(cmd.Payload)
	case CmdSetPlayerReady:
		g.handleSetPlayerReady(cmd.Payload)
	case CmdCancelGame:
		g.handleCancelGame()
	case CmdNextQuestion:
		g.handleNextQuestion(cmd.Payload)
	case CmdPlayerDisconnected:
		g.handlePlayerDisconnected(cmd.Payload)
	case CmdPlayerReconnected:
		g.handlePlayerReconnected(cmd.Payload)
	case CmdReconnectTimeout:
		g.handleReconnectTimeout()
	}
}

func (g *BaseGame) handleStartGame() {
	isMultiplayer := g.mode != model.GameModeSolo
	if isMultiplayer {
		g.mutex.RLock()
		var playerPayloads []map[string]interface{}
		for _, p := range g.players {
			playerPayloads = append(playerPayloads, map[string]interface{}{
				keyUserPublicID: p.PublicID,
				keyUsername:       p.Username,
			})
		}
		g.mutex.RUnlock()

		countdownSecs := 3
		if g.countdownConfig != nil && g.countdownConfig.PreGameCountdownSeconds > 0 {
			countdownSecs = g.countdownConfig.PreGameCountdownSeconds
		}
		g.emitEvent(GameEvent{
			Type: EventGameStarting,
			Data: map[string]interface{}{
				keyCountdownSecs: countdownSecs,
				keyPlayers:        playerPayloads,
			},
		})
	}

	err := g.Start()
	if err != nil {
		g.emitEvent(GameEvent{
			Type: keyGameError,
			Data: map[string]interface{}{keyError: err.Error()},
		})
		return
	}

	g.mutex.Lock()
	if g.ticker != nil {
		g.ticker.Stop()
	}
	firstQuestion := g.questions[g.currentQ]
	estimatedSecs := firstQuestion.EstimatedSeconds
	if estimatedSecs <= 0 {
		estimatedSecs = 30
	}
	g.ticker = time.NewTicker(time.Duration(estimatedSecs) * time.Second)

	questionSentAt := time.Now()
	for _, p := range g.players {
		p.CurrentQuestionSentAt = questionSentAt
		if g.logger != nil {
			g.logger.Debug("Set CurrentQuestionSentAt",
				zap.String("userID", p.UserID.String()),
				zap.Time("sentAt", questionSentAt),
			)
		}
	}
	g.mutex.Unlock()

	g.emitEvent(GameEvent{
		Type: EventGameStarted,
		Data: map[string]interface{}{keyPlayers: g.buildPlayersFinal()},
	})

	g.mutex.RLock()
	firstQ := g.questions[g.currentQ]
	qNum := g.currentQ + 1
	totalQ := len(g.questions)
	g.mutex.RUnlock()

	g.emitEvent(GameEvent{
		Type: EventQuestionSent,
		Data: map[string]interface{}{
			keyQuestion:        QuestionToPayload(firstQ),
			keyQuestionNumber: qNum,
			keyTotalQuestions: totalQ,
			keyTimeLimit:      firstQ.EstimatedSeconds,
		},
	})
}

func (g *BaseGame) handleEndGame(payload interface{}) {
	data, ok := payload.(map[string]interface{})
	if !ok {
		return
	}

	var winnerID *uuid.UUID
	if wid, ok := data["winnerID"].(uuid.UUID); ok {
		winnerID = &wid
	}

	err := g.End(winnerID)
	if err != nil {
		g.emitEvent(GameEvent{
			Type: keyGameError,
			Data: map[string]interface{}{keyError: err.Error()},
		})
		return
	}

	g.mutex.Lock()
	if g.ticker != nil {
		g.ticker.Stop()
		g.ticker = nil
	}
	g.mutex.Unlock()

	var winnerPublicID *string
	if winnerID != nil {
		pid := g.playerPublicID(*winnerID)
		winnerPublicID = &pid
	}
	g.emitEvent(GameEvent{
		Type: EventGameCompleted,
		Data: map[string]interface{}{
			keyGameID:         g.publicID,
			keyWinnerPublicID: winnerPublicID,
			keyPlayersFinal:   g.buildPlayersFinal(),
		},
	})
}

func (g *BaseGame) handleSubmitAnswer(payload interface{}) {
	data, ok := payload.(SubmitAnswerPayload)
	if !ok {
		return
	}

	g.mutex.RLock()
	prevQ := g.currentQ
	g.mutex.RUnlock()

	err := g.SubmitAnswer(data.UserID, data.Answer)
	if err != nil {
		g.emitEvent(GameEvent{
			Type: keyGameError,
			Data: map[string]interface{}{keyError: err.Error()},
		})
		return
	}

	g.mutex.RLock()
	currentQ := g.currentQ
	questionsLen := len(g.questions)
	player, playerExists := g.players[data.UserID]
	var validatedAnswer Answer
	if playerExists && len(player.Answers) > prevQ {
		validatedAnswer = player.Answers[prevQ]
	} else {
		validatedAnswer = data.Answer
	}
	var prevQuestion *model.Question
	if prevQ < len(g.questions) {
		prevQuestion = g.questions[prevQ]
	}
	g.mutex.RUnlock()

	resultPayload := BuildAnswerResultPayload(validatedAnswer, prevQuestion)

	userPublicID := ""
	if playerExists {
		userPublicID = player.PublicID
	}

	g.emitEvent(GameEvent{
		Type: EventAnswerReceived,
		Data: map[string]interface{}{
			keyUserPublicID:  userPublicID,
			"is_correct":      resultPayload.IsCorrect,
			"points":          resultPayload.Points,
			"time_spent_ms":   resultPayload.TimeSpentMs,
			"correct_answer":  resultPayload.CorrectAnswer,
			"submitted_slug":  validatedAnswer.GetAnswerSlug(),
			keyQuestionNumber: prevQ + 1,
		},
	})

	if currentQ != prevQ {
		g.mutex.RLock()
		scoreUpdates := make([]map[string]interface{}, 0, len(g.players))
		for _, p := range g.players {
			scoreUpdates = append(scoreUpdates, map[string]interface{}{
				keyUserPublicID: p.PublicID,
				keyScore:          p.Score,
			})
		}
		g.mutex.RUnlock()
		for _, su := range scoreUpdates {
			g.emitEvent(GameEvent{Type: EventScoreUpdated, Data: su})
		}
	}

	if currentQ != prevQ && currentQ < questionsLen {
		g.mutex.Lock()
		if g.ticker != nil {
			g.ticker.Stop()
			g.ticker = nil
		}
		g.mutex.Unlock()
		g.scheduleNextQuestion(currentQ)
	}

	if currentQ >= questionsLen {
		winnerID := g.determineWinner()

		g.mutex.Lock()
		if g.ticker != nil {
			g.ticker.Stop()
			g.ticker = nil
		}
		g.mutex.Unlock()

		if err := g.End(winnerID); err != nil {
			return
		}

		var winnerPublicID *string
		if winnerID != nil {
			pid := g.playerPublicID(*winnerID)
			winnerPublicID = &pid
		}
		g.emitEvent(GameEvent{
			Type: EventGameCompleted,
			Data: map[string]interface{}{
				keyGameID:          g.publicID,
				keyWinnerPublicID: winnerPublicID,
				keyPlayersFinal:    g.buildPlayersFinal(),
			},
		})
	}
}

func (g *BaseGame) handleTimer() {
	g.mutex.Lock()

	if g.paused {
		g.mutex.Unlock()
		return
	}

	if g.status == model.GameStatusCompleted || g.status == model.GameStatusCancelled || g.status == model.GameStatusAbandoned {
		g.mutex.Unlock()
		return
	}

	if g.currentQ >= len(g.questions) {
		g.mutex.Unlock()
		return
	}

	currentQuestion := g.questions[g.currentQ]
	now := time.Now()

	for _, p := range g.players {
		if p.Status == model.PlayerStatusLeft {
			continue
		}
		if len(p.Answers) <= g.currentQ {
			answer := NewMCQAnswer(
				currentQuestion.ID,
				"",
				time.Duration(currentQuestion.EstimatedSeconds)*time.Second,
			)
			answer.IsCorrect = false
			answer.Points = 0
			answer.ServerTimeSpent = now.Sub(p.CurrentQuestionSentAt)
			answer.ReceivedAt = now
			p.Answers = append(p.Answers, answer)
		}
	}

	timedOutQ := g.currentQ
	g.hooks.EvaluateQuestionResults(g, timedOutQ)
	g.currentQ++

	if g.ticker != nil {
		g.ticker.Stop()
		g.ticker = nil
	}

	if g.currentQ >= len(g.questions) {
		winnerID := g.determineWinnerLocked()
		g.endLocked(winnerID)
		g.mutex.Unlock()

		g.emitEvent(GameEvent{
			Type: EventQuestionTimeout,
			Data: map[string]interface{}{keyQuestionNumber: timedOutQ + 1},
		})
		var winnerPublicID *string
		if winnerID != nil {
			pid := g.playerPublicID(*winnerID)
			winnerPublicID = &pid
		}
		g.emitEvent(GameEvent{
			Type: EventGameCompleted,
			Data: map[string]interface{}{
				keyGameID:          g.publicID,
				keyWinnerPublicID: winnerPublicID,
				keyPlayersFinal:    g.buildPlayersFinal(),
			},
		})
		return
	}

	nextQIdx := g.currentQ
	g.mutex.Unlock()

	g.emitEvent(GameEvent{
		Type: EventQuestionTimeout,
		Data: map[string]interface{}{keyQuestionNumber: timedOutQ + 1},
	})
	g.scheduleNextQuestion(nextQIdx)
}

func (g *BaseGame) handleAddPlayer(payload interface{}) {
	data, ok := payload.(AddPlayerPayload)
	if !ok {
		return
	}

	err := g.AddPlayer(data.UserID)
	if err != nil {
		g.emitEvent(GameEvent{
			Type: keyGameError,
			Data: map[string]interface{}{keyError: err.Error()},
		})
		return
	}

	g.mutex.RLock()
	joinedPlayer := g.players[data.UserID]
	g.mutex.RUnlock()
	joinedPublicID := ""
	joinedUsername := ""
	if joinedPlayer != nil {
		joinedPublicID = joinedPlayer.PublicID
		joinedUsername = joinedPlayer.Username
	}
	g.emitEvent(GameEvent{
		Type: EventPlayerJoined,
		Data: map[string]interface{}{
			keyUserPublicID: joinedPublicID,
			keyUsername:       joinedUsername,
		},
	})
}

func (g *BaseGame) handleRemovePlayer(payload interface{}) {
	data, ok := payload.(RemovePlayerPayload)
	if !ok {
		return
	}

	err := g.RemovePlayer(data.UserID)
	if err != nil {
		g.emitEvent(GameEvent{
			Type: keyGameError,
			Data: map[string]interface{}{keyError: err.Error()},
		})
		return
	}

	g.emitEvent(GameEvent{
		Type: EventPlayerLeft,
		Data: map[string]interface{}{keyUserPublicID: g.playerPublicID(data.UserID)},
	})
}

func (g *BaseGame) handleSetPlayerReady(payload interface{}) {
	data, ok := payload.(SetPlayerReadyPayload)
	if !ok {
		return
	}

	err := g.SetPlayerReady(data.UserID, data.Ready)
	if err != nil {
		g.emitEvent(GameEvent{
			Type: keyGameError,
			Data: map[string]interface{}{keyError: err.Error()},
		})
		return
	}

	g.emitEvent(GameEvent{
		Type: EventPlayerReady,
		Data: map[string]interface{}{
			keyUserPublicID: g.playerPublicID(data.UserID),
			"ready":          data.Ready,
		},
	})
}

func (g *BaseGame) handleCancelGame() {
	err := g.Cancel()
	if err != nil {
		g.emitEvent(GameEvent{
			Type: keyGameError,
			Data: map[string]interface{}{keyError: err.Error()},
		})
		return
	}

	g.mutex.Lock()
	if g.ticker != nil {
		g.ticker.Stop()
		g.ticker = nil
	}
	g.mutex.Unlock()

	g.emitEvent(GameEvent{
		Type: EventGameCancelled,
		Data: map[string]interface{}{keyReason: "cancelled"},
	})
}

func (g *BaseGame) scheduleNextQuestion(questionIndex int) {
	delay := time.Duration(g.settings.InterRoundDelayMs) * time.Millisecond
	if delay <= 0 {
		g.handleNextQuestion(NextQuestionPayload{QuestionIndex: questionIndex})
		return
	}
	t := time.AfterFunc(delay, func() {
		g.SendCommand(GameCommand{
			Type:    CmdNextQuestion,
			Payload: NextQuestionPayload{QuestionIndex: questionIndex},
		})
	})
	g.mutex.Lock()
	g.pendingTimers = append(g.pendingTimers, t)
	g.mutex.Unlock()
}

func (g *BaseGame) handleNextQuestion(payload interface{}) {
	data, ok := payload.(NextQuestionPayload)
	if !ok {
		return
	}

	g.mutex.Lock()
	if g.status != model.GameStatusInProgress {
		g.mutex.Unlock()
		return
	}
	if g.paused {
		g.mutex.Unlock()
		return
	}
	if g.currentQ != data.QuestionIndex || g.currentQ >= len(g.questions) {
		g.mutex.Unlock()
		return
	}

	g.pendingTimers = nil

	if g.ticker != nil {
		g.ticker.Stop()
	}
	nextQuestion := g.questions[g.currentQ]
	nextEstimatedSecs := nextQuestion.EstimatedSeconds
	if nextEstimatedSecs <= 0 {
		nextEstimatedSecs = 30
	}
	g.ticker = time.NewTicker(time.Duration(nextEstimatedSecs) * time.Second)

	questionSentAt := time.Now()
	for _, p := range g.players {
		p.CurrentQuestionSentAt = questionSentAt
	}
	questionsLen := len(g.questions)
	currentQ := g.currentQ
	g.mutex.Unlock()

	g.emitEvent(GameEvent{
		Type: EventQuestionSent,
		Data: map[string]interface{}{
			keyQuestion:        QuestionToPayload(nextQuestion),
			keyQuestionNumber: currentQ + 1,
			keyTotalQuestions: questionsLen,
			keyTimeLimit:      nextQuestion.EstimatedSeconds,
		},
	})
}

func (g *BaseGame) handlePlayerDisconnected(payload interface{}) {
	data, ok := payload.(RemovePlayerPayload)
	if !ok {
		return
	}
	if err := g.DisconnectPlayer(data.UserID); err != nil {
		return
	}
	g.emitEvent(GameEvent{
		Type: EventPlayerDisconnected,
		Data: map[string]interface{}{keyUserPublicID: g.playerPublicID(data.UserID)},
	})

	g.mutex.RLock()
	status := g.status
	mode := g.mode
	g.mutex.RUnlock()

	if status == model.GameStatusInProgress && mode != model.GameModeSolo {
		graceSecs := g.reconnectGracePeriodSecs()
		g.pauseForReconnect(graceSecs)

		g.mutex.Lock()
		if g.reconnectTimer != nil {
			g.reconnectTimer.Stop()
		}
		g.reconnectTimer = time.AfterFunc(time.Duration(graceSecs)*time.Second, func() {
			g.SendCommand(GameCommand{Type: CmdReconnectTimeout})
		})
		g.mutex.Unlock()
	}
}

func (g *BaseGame) reconnectGracePeriodSecs() int {
	if g.countdownConfig != nil && g.countdownConfig.ReconnectGracePeriodSeconds > 0 {
		return g.countdownConfig.ReconnectGracePeriodSeconds
	}
	return model.DefaultCountdownConfig().ReconnectGracePeriodSeconds
}

func (g *BaseGame) handleReconnectTimeout() {
	g.mutex.Lock()
	g.reconnectTimer = nil

	for _, p := range g.players {
		if p.Status == model.PlayerStatusDisconnected {
			p.Status = model.PlayerStatusLeft
		}
	}

	activeCount := 0
	for _, p := range g.players {
		if p.Status == model.PlayerStatusActive {
			activeCount++
		}
	}

	winnerID := g.determineWinnerLocked()
	g.mutex.Unlock()

	if activeCount == 0 {
		_ = g.Cancel()
		g.emitEvent(GameEvent{
			Type: EventGameCancelled,
			Data: map[string]interface{}{keyReason: CmdReconnectTimeout},
		})
		return
	}

	_ = g.End(winnerID)
	var winnerPublicID *string
	if winnerID != nil {
		pid := g.playerPublicID(*winnerID)
		winnerPublicID = &pid
	}
	g.emitEvent(GameEvent{
		Type: EventGameCompleted,
		Data: map[string]interface{}{
			keyGameID:          g.publicID,
			keyWinnerPublicID: winnerPublicID,
			keyPlayersFinal:    g.buildPlayersFinal(),
			keyReason:           CmdReconnectTimeout,
		},
	})
}

func (g *BaseGame) handlePlayerReconnected(payload interface{}) {
	data, ok := payload.(ReconnectPlayerPayload)
	if !ok {
		return
	}
	if err := g.ReconnectPlayer(data.UserID); err != nil {
		return
	}

	g.mutex.Lock()
	if g.reconnectTimer != nil {
		g.reconnectTimer.Stop()
		g.reconnectTimer = nil
	}
	isPaused := g.paused
	g.mutex.Unlock()

	g.emitEvent(GameEvent{
		Type: EventPlayerReconnected,
		Data: map[string]interface{}{keyUserPublicID: g.playerPublicID(data.UserID)},
	})

	if isPaused {
		g.resumeFromPause()
	}
}

func (g *BaseGame) playerPublicID(userID uuid.UUID) string {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	if p, ok := g.players[userID]; ok {
		return p.PublicID
	}
	return ""
}

func (g *BaseGame) buildPlayersFinalLocked() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(g.players))
	for _, p := range g.players {
		result = append(result, map[string]interface{}{
			keyUserPublicID: p.PublicID,
			keyUsername:       p.Username,
			keyScore:          p.Score,
			"status":         string(p.Status),
		})
	}
	return result
}

func (g *BaseGame) buildPlayersFinal() []map[string]interface{} {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.buildPlayersFinalLocked()
}

func (g *BaseGame) emitEvent(event GameEvent) {
	event.GameID = g.id
	event.PublicID = g.publicID
	event.Timestamp = time.Now()

	select {
	case g.eventChan <- event:
	default:
		if g.logger != nil {
			g.logger.Warn("Game event channel full, dropping event",
				zap.String("event_type", event.Type),
				zap.String(keyGameID, g.id.String()),
			)
		}
	}
}

func (g *BaseGame) GetCurrentQuestionIndex() int {
	return g.currentQ
}

func (g *BaseGame) SetAllPlayersQuestionStartTime(t time.Time) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	for _, p := range g.players {
		p.CurrentQuestionSentAt = t
	}
}

func (g *BaseGame) CheckAndAdvanceTimeout(now time.Time) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.status != model.GameStatusInProgress {
		return false
	}
	if g.paused {
		return false
	}
	if g.currentQ >= len(g.questions) {
		return false
	}

	currentQuestion := g.questions[g.currentQ]
	timeLimit := time.Duration(currentQuestion.EstimatedSeconds) * time.Second

	anyTimedOut := false
	for _, p := range g.players {
		if p.Status == model.PlayerStatusLeft {
			continue
		}
		if len(p.Answers) <= g.currentQ {
			if !p.CurrentQuestionSentAt.IsZero() && now.Sub(p.CurrentQuestionSentAt) >= timeLimit {
				anyTimedOut = true
				break
			}
		}
	}
	if !anyTimedOut {
		return false
	}

	for _, p := range g.players {
		if p.Status == model.PlayerStatusLeft {
			continue
		}
		if len(p.Answers) <= g.currentQ {
			serverTimeSpent := timeLimit
			if !p.CurrentQuestionSentAt.IsZero() {
				serverTimeSpent = now.Sub(p.CurrentQuestionSentAt)
			}
			answer := NewMCQAnswer(currentQuestion.ID, "", timeLimit)
			answer.IsCorrect = false
			answer.Points = 0
			answer.ServerTimeSpent = serverTimeSpent
			answer.ReceivedAt = now
			p.Answers = append(p.Answers, answer)
		}
	}

	timedOutQ := g.currentQ
	g.hooks.EvaluateQuestionResults(g, timedOutQ)
	g.currentQ++

	if g.currentQ < len(g.questions) {
		questionSentAt := time.Now()
		for _, p := range g.players {
			p.CurrentQuestionSentAt = questionSentAt
		}
	}

	return true
}

func (g *BaseGame) GetPaused() bool {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.paused
}

func (g *BaseGame) GetPausedAt() time.Time {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.pausedAt
}

func (g *BaseGame) SetPausedState(paused bool, pausedAt time.Time) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.paused = paused
	g.pausedAt = pausedAt
}

func (g *BaseGame) GetReconnectDeadline() *time.Time {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.reconnectDeadline
}

func (g *BaseGame) SetReconnectDeadline(deadline *time.Time) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.reconnectDeadline = deadline
}

func (g *BaseGame) AdjustQuestionTimeForPause(pauseDuration time.Duration) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	for _, p := range g.players {
		if !p.CurrentQuestionSentAt.IsZero() {
			p.CurrentQuestionSentAt = p.CurrentQuestionSentAt.Add(pauseDuration)
		}
	}
}

func (g *BaseGame) GetQuestionByIndex(index int) *model.Question {
	if index < 0 || index >= len(g.questions) {
		return nil
	}
	return g.questions[index]
}

func (g *BaseGame) pauseForReconnect(graceSecs int) {
	g.mutex.Lock()

	if g.paused || g.status != model.GameStatusInProgress {
		g.mutex.Unlock()
		return
	}

	now := time.Now()
	g.paused = true
	g.pausedAt = now
	g.pausedQuestionIdx = g.currentQ

	if g.ticker != nil {
		g.ticker.Stop()
		g.ticker = nil

		if g.currentQ < len(g.questions) {
			q := g.questions[g.currentQ]
			timeLimit := time.Duration(q.EstimatedSeconds) * time.Second
			var sentAt time.Time
			for _, p := range g.players {
				if !p.CurrentQuestionSentAt.IsZero() {
					sentAt = p.CurrentQuestionSentAt
					break
				}
			}
			if !sentAt.IsZero() {
				remaining := timeLimit - now.Sub(sentAt)
				if remaining < 0 {
					remaining = 0
				}
				g.tickerRemainingMs = remaining.Milliseconds()
			}
		}
	} else {
		for _, t := range g.pendingTimers {
			t.Stop()
		}
		g.pendingTimers = nil
		g.tickerRemainingMs = -1
	}

	g.mutex.Unlock()

	g.emitEvent(GameEvent{
		Type: EventGamePaused,
		Data: map[string]interface{}{
			keyCountdownSecs: graceSecs,
		},
	})
}

func (g *BaseGame) resumeFromPause() {
	g.mutex.Lock()

	if !g.paused || g.status != model.GameStatusInProgress {
		g.mutex.Unlock()
		return
	}

	pauseDuration := time.Since(g.pausedAt)
	g.paused = false
	g.pausedAt = time.Time{}

	for _, p := range g.players {
		if !p.CurrentQuestionSentAt.IsZero() {
			p.CurrentQuestionSentAt = p.CurrentQuestionSentAt.Add(pauseDuration)
		}
	}

	needsSchedule := g.tickerRemainingMs < 0
	remainingMs := g.tickerRemainingMs
	currentQ := g.currentQ
	g.tickerRemainingMs = 0

	if remainingMs > 0 {
		g.ticker = time.NewTicker(time.Duration(remainingMs) * time.Millisecond)
	}

	g.mutex.Unlock()

	g.emitEvent(GameEvent{
		Type: EventGameResumed,
		Data: map[string]interface{}{},
	})

	if needsSchedule {
		g.scheduleNextQuestion(currentQ)
	}
}
