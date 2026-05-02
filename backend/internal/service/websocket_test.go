package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/model"
	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const testGameID1 = "game-1"

func init() {
	gin.SetMode(gin.TestMode)
}

type mockRedis struct {
	mu  sync.Mutex
	cfg *model.WebSocketConfig
}

func (m *mockRedis) GetJSON(_ context.Context, _ string, dest interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cfg == nil {
		return errors.New("not found")
	}
	data, _ := json.Marshal(m.cfg)
	return json.Unmarshal(data, dest)
}

func (m *mockRedis) Ping(_ context.Context) error                                              { return nil }
func (m *mockRedis) Close() error                                                              { return nil }
func (m *mockRedis) Set(_ context.Context, _ string, _ interface{}, _ time.Duration) error    { return nil }
func (m *mockRedis) Get(_ context.Context, _ string) (string, error)                          { return "", nil }
func (m *mockRedis) Delete(_ context.Context, _ string) error                                 { return nil }
func (m *mockRedis) Exists(_ context.Context, _ string) (bool, error)                         { return false, nil }
func (m *mockRedis) SetJSON(_ context.Context, _ string, _ interface{}, _ time.Duration) error { return nil }
func (m *mockRedis) SetCache(_ context.Context, _ string, _ interface{}, _ time.Duration) error { return nil }
func (m *mockRedis) GetCache(_ context.Context, _ string, _ interface{}) error                { return nil }
func (m *mockRedis) DeleteCache(_ context.Context, _ string) error                            { return nil }
func (m *mockRedis) Publish(_ context.Context, _ string, _ interface{}) error                 { return nil }
func (m *mockRedis) PublishRaw(_ context.Context, _ string, _ string) error                      { return nil }
func (m *mockRedis) Subscribe(_ context.Context, _ ...string) *redis.PubSub                   { return nil }
func (m *mockRedis) GetInfo(_ context.Context) (map[string]interface{}, error)                { return nil, nil }
func (m *mockRedis) Scan(_ context.Context, _ string) ([]string, error)                       { return nil, nil }
func (m *mockRedis) Expire(_ context.Context, _ string, _ time.Duration) error                { return nil }
func (m *mockRedis) RPush(_ context.Context, _ string, _ ...interface{}) error                { return nil }
func (m *mockRedis) LRem(_ context.Context, _ string, _ int64, _ interface{}) error           { return nil }
func (m *mockRedis) LLen(_ context.Context, _ string) (int64, error)                          { return 0, nil }
func (m *mockRedis) LPop(_ context.Context, _ string) (string, error)                         { return "", nil }
func (m *mockRedis) LPush(_ context.Context, _ string, _ ...interface{}) error                { return nil }
func (m *mockRedis) ZAdd(_ context.Context, _ string, _ ...cache.ZSetMember) error            { return nil }
func (m *mockRedis) ZRem(_ context.Context, _ string, _ ...string) error                      { return nil }
func (m *mockRedis) ZRangeByScore(_ context.Context, _ string, _ cache.ZRangeByScoreOptions) ([]string, error) {
	return nil, nil
}
func (m *mockRedis) ZGetScore(_ context.Context, _ string, _ string) (float64, error)                    { return 0, nil }
func (m *mockRedis) MGet(_ context.Context, _ ...string) ([]string, error)                               { return nil, nil }
func (m *mockRedis) SetNX(_ context.Context, _ string, _ interface{}, _ time.Duration) (bool, error)     { return true, nil }
func (m *mockRedis) Eval(_ context.Context, _ string, _ []string, _ ...interface{}) (interface{}, error) { return nil, nil }
func (m *mockRedis) NativeClient() redis.UniversalClient                                                 { return nil }
func (m *mockRedis) HIncrBy(_ context.Context, _ string, _ string, _ int64) (int64, error)           { return 1, nil }
func (m *mockRedis) HLen(_ context.Context, _ string) (int64, error)                                 { return 0, nil }
func (m *mockRedis) HDel(_ context.Context, _ string, _ ...string) error                                    { return nil }
func (m *mockRedis) HSet(_ context.Context, _, _ string, _ interface{}) error                               { return nil }
func (m *mockRedis) HGetAll(_ context.Context, _ string) (map[string]string, error)                         { return nil, nil }
func (m *mockRedis) HGet(_ context.Context, _, _ string) (string, error)                                    { return "", nil }
func (m *mockRedis) HDelete(_ context.Context, _ string, _ ...string) error                                 { return nil }
func (m *mockRedis) CheckRateLimit(_ context.Context, _ string, _ int64, _ time.Duration) (bool, int64, int64, error) {
	return false, 0, 0, nil
}

type mockStatusUpdater struct {
	mu    sync.Mutex
	calls []struct {
		userID   uuid.UUID
		isOnline bool
	}
	err error
}

func (m *mockStatusUpdater) UpdateUserOnlineStatus(userID uuid.UUID, isOnline bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, struct {
		userID   uuid.UUID
		isOnline bool
	}{userID, isOnline})
	return m.err
}

type mockGameHandler struct {
	mu sync.Mutex

	submitAnswerErr error
	playerReadyErr  error
	startGameErr    error
	leaveGameErr    error

	currentQuestion map[string]interface{}
	currentQErr     error
	gameState       map[string]interface{}
	gameStateErr    error

	disconnectedPlayers []uuid.UUID
	reconnectedPlayers  []uuid.UUID
	leaveGameCalls      int
}

func (m *mockGameHandler) HandleSubmitAnswer(_ uuid.UUID, _ string, _ map[string]interface{}) error {
	return m.submitAnswerErr
}
func (m *mockGameHandler) HandlePlayerReady(_ uuid.UUID, _ string, _ bool) error {
	return m.playerReadyErr
}
func (m *mockGameHandler) HandleStartGame(_ uuid.UUID, _ string) error { return m.startGameErr }
func (m *mockGameHandler) HandleLeaveGame(_ uuid.UUID, _ string) error {
	m.mu.Lock()
	m.leaveGameCalls++
	m.mu.Unlock()
	return m.leaveGameErr
}
func (m *mockGameHandler) GetCurrentQuestionPayload(_ string) (map[string]interface{}, error) {
	return m.currentQuestion, m.currentQErr
}
func (m *mockGameHandler) GetGameStateForReconnect(_ string) (map[string]interface{}, error) {
	return m.gameState, m.gameStateErr
}
func (m *mockGameHandler) MarkPlayerDisconnected(userID uuid.UUID, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.disconnectedPlayers = append(m.disconnectedPlayers, userID)
	return nil
}
func (m *mockGameHandler) MarkPlayerReconnected(userID uuid.UUID, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reconnectedPlayers = append(m.reconnectedPlayers, userID)
	return nil
}
func (m *mockGameHandler) OnUserConnected(_ uuid.UUID) {}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func newService(t *testing.T, cfg *model.WebSocketConfig) *WebSocketService {
	t.Helper()
	return NewWebSocketServiceWithMode(nil, &mockRedis{cfg: cfg}, zap.NewNop(), false)
}

// newTestServer creates an httptest.Server that routes /ws (regular) and /admin/ws (admin).
// The userID is read from the "uid" query parameter; if absent, a random UUID is used.
func newTestServer(t *testing.T, svc *WebSocketService) *httptest.Server {
	t.Helper()
	r := gin.New()
	r.GET("/ws", func(c *gin.Context) {
		uid := parseUID(c)
		_ = svc.UpgradeConnection(c, uid)
	})
	r.GET("/admin/ws", func(c *gin.Context) {
		uid := parseUID(c)
		_ = svc.UpgradeAdminConnection(c, uid)
	})
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

func parseUID(c *gin.Context) uuid.UUID {
	if id, err := uuid.Parse(c.Query("uid")); err == nil {
		return id
	}
	return uuid.New()
}

func dial(t *testing.T, srv *httptest.Server, path string, uid uuid.UUID) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + path + "?uid=" + uid.String()
	conn, _, err := websocket.Dial(ctx, url, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.CloseNow() })
	return conn
}

func readMsg(t *testing.T, conn *websocket.Conn) map[string]interface{} {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, data, err := conn.Read(ctx)
	require.NoError(t, err, "timed out waiting for a message")
	var msg map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &msg))
	return msg
}

func skipHello(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	msg := readMsg(t, conn)
	require.Equal(t, "hello", msg[keyType], "first message must be hello")
}

func send(t *testing.T, conn *websocket.Conn, payload interface{}) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	data, _ := json.Marshal(payload)
	require.NoError(t, conn.Write(ctx, websocket.MessageText, data))
}

// pingPong sends a ping and asserts the immediate response is a pong.
// Use this to verify the connection is alive and the message queue is otherwise empty.
func pingPong(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	send(t, conn, map[string]string{keyType: msgPing})
	msg := readMsg(t, conn)
	assert.Equal(t, "pong", msg[keyType], "expected pong in response to ping")
}

// ─── Connection & handshake ───────────────────────────────────────────────────

func TestConnect_HelloMessage_Fields(t *testing.T) {
	svc := newService(t, nil)
	srv := newTestServer(t, svc)

	conn := dial(t, srv, "/ws", uuid.New())
	msg := readMsg(t, conn)

	assert.Equal(t, "hello", msg[keyType])
	assert.Equal(t, float64(1), msg["protocol_version"])
	assert.NotEmpty(t, msg["server_time"])
	assert.Equal(t, float64(30000), msg["heartbeat_interval"])
}

func TestConnect_IncrementsAndDecrementsClientCount(t *testing.T) {
	svc := newService(t, nil)
	srv := newTestServer(t, svc)

	assert.Equal(t, 0, svc.GetConnectedClients())

	c1 := dial(t, srv, "/ws", uuid.New())
	c2 := dial(t, srv, "/ws", uuid.New())
	skipHello(t, c1)
	skipHello(t, c2)
	assert.Equal(t, 2, svc.GetConnectedClients())

	_ = c1.CloseNow()
	require.Eventually(t, func() bool { return svc.GetConnectedClients() == 1 }, time.Second, 10*time.Millisecond)

	_ = c2.CloseNow()
	require.Eventually(t, func() bool { return svc.GetConnectedClients() == 0 }, time.Second, 10*time.Millisecond)
}

func TestConnect_SameUser_TwoDevices_CountedOnce(t *testing.T) {
	svc := newService(t, nil)
	srv := newTestServer(t, svc)
	uid := uuid.New()

	c1 := dial(t, srv, "/ws", uid)
	c2 := dial(t, srv, "/ws", uid)
	skipHello(t, c1)
	skipHello(t, c2)

	assert.Equal(t, 2, svc.GetConnectedClients())
	assert.Equal(t, 1, svc.GetOnlineUsers(), "two connections from the same user count as one online user")
}

// ─── Ping / pong ─────────────────────────────────────────────────────────────

func TestPing_ReturnsPong(t *testing.T) {
	svc := newService(t, nil)
	srv := newTestServer(t, svc)
	conn := dial(t, srv, "/ws", uuid.New())
	skipHello(t, conn)

	pingPong(t, conn)
}

// ─── Unknown / malformed messages ────────────────────────────────────────────

func TestUnknownMessageType_AckWithError(t *testing.T) {
	svc := newService(t, nil)
	srv := newTestServer(t, svc)
	conn := dial(t, srv, "/ws", uuid.New())
	skipHello(t, conn)

	send(t, conn, map[string]string{
		keyType:           "not_a_real_type",
		keyCorrelationID: "cid-001",
	})

	msg := readMsg(t, conn)
	assert.Equal(t, msgAck, msg[keyType])
	assert.Equal(t, "not_a_real_type", msg["action"])
	assert.Equal(t, false, msg["success"])
	assert.Contains(t, msg["error"], "unknown message type")
	assert.Equal(t, "cid-001", msg[keyCorrelationID])
}

func TestInvalidJSON_SilentlyDropped_ConnectionSurvives(t *testing.T) {
	// Invalid JSON must not close the connection or produce an ack.
	svc := newService(t, nil)
	srv := newTestServer(t, svc)
	conn := dial(t, srv, "/ws", uuid.New())
	skipHello(t, conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	require.NoError(t, conn.Write(ctx, websocket.MessageText, []byte("not { valid json at all")))

	// Connection must still be alive.
	pingPong(t, conn)
}

func TestMessageMissingTypeField_SilentlyDropped(t *testing.T) {
	svc := newService(t, nil)
	srv := newTestServer(t, svc)
	conn := dial(t, srv, "/ws", uuid.New())
	skipHello(t, conn)

	// Valid JSON but no keyType field.
	send(t, conn, map[string]string{"action": msgJoinGame, keyGamePublicID: "x"})

	pingPong(t, conn)
}

// ─── join_game ────────────────────────────────────────────────────────────────

func TestJoinGame_AckSuccess(t *testing.T) {
	svc := newService(t, nil)
	srv := newTestServer(t, svc)
	conn := dial(t, srv, "/ws", uuid.New())
	skipHello(t, conn)

	send(t, conn, map[string]string{
		keyType:           msgJoinGame,
		keyGamePublicID: testGameID1,
		keyCorrelationID: "cid-join",
	})

	msg := readMsg(t, conn)
	assert.Equal(t, msgAck, msg[keyType])
	assert.Equal(t, msgJoinGame, msg["action"])
	assert.Equal(t, true, msg["success"])
	assert.Equal(t, "cid-join", msg[keyCorrelationID])
}

func TestJoinGame_Idempotent_NoDuplicateInRoom(t *testing.T) {
	// Joining the same game twice must not duplicate the client in the room.
	// If it did, the client would receive every broadcast twice.
	svc := newService(t, nil)
	srv := newTestServer(t, svc)
	conn := dial(t, srv, "/ws", uuid.New())
	skipHello(t, conn)

	for range 3 {
		send(t, conn, map[string]string{keyType: msgJoinGame, keyGamePublicID: "game-dup"})
		msg := readMsg(t, conn)
		assert.Equal(t, true, msg["success"])
	}

	svc.mutex.RLock()
	count := len(svc.gameClients["game-dup"])
	svc.mutex.RUnlock()
	assert.Equal(t, 1, count, "client must appear exactly once regardless of how many join_game messages were sent")
}

func TestJoinGame_SendsCurrentQuestion_WhenGameIsActive(t *testing.T) {
	// If a question is already running when a player joins, they must receive it immediately
	// so they don't miss the round.
	svc := newService(t, nil)
	svc.SetGameActionHandler(&mockGameHandler{
		currentQuestion: map[string]interface{}{
			"question_number": float64(3),
			"text":            "What is the capital of Japan?",
		},
	})
	srv := newTestServer(t, svc)
	conn := dial(t, srv, "/ws", uuid.New())
	skipHello(t, conn)

	send(t, conn, map[string]string{keyType: msgJoinGame, keyGamePublicID: "game-live"})

	ack := readMsg(t, conn)
	assert.Equal(t, msgJoinGame, ack["action"])
	assert.Equal(t, true, ack["success"])

	question := readMsg(t, conn)
	assert.Equal(t, "question_sent", question[keyType])
	assert.Equal(t, "game-live", question["public_id"])
}

func TestJoinGame_NoCurrentQuestion_NoExtraMessage(t *testing.T) {
	// If there is no active question, only the ack must be sent — nothing more.
	svc := newService(t, nil)
	svc.SetGameActionHandler(&mockGameHandler{
		currentQErr: errors.New("no active question"),
	})
	srv := newTestServer(t, svc)
	conn := dial(t, srv, "/ws", uuid.New())
	skipHello(t, conn)

	send(t, conn, map[string]string{keyType: msgJoinGame, keyGamePublicID: "game-lobby"})
	msg := readMsg(t, conn) // ack
	assert.Equal(t, msgJoinGame, msg["action"])

	// The next message must be the pong we ask for — not some stray extra message.
	pingPong(t, conn)
}

// ─── leave_game ───────────────────────────────────────────────────────────────

func TestLeaveGame_WhenNotInAnyGame_AckError(t *testing.T) {
	svc := newService(t, nil)
	srv := newTestServer(t, svc)
	conn := dial(t, srv, "/ws", uuid.New())
	skipHello(t, conn)

	send(t, conn, map[string]interface{}{
		keyType:           msgLeaveGame,
		keyCorrelationID: "cid-leave",
	})

	msg := readMsg(t, conn)
	assert.Equal(t, false, msg["success"])
	assert.Equal(t, "not in a game", msg["error"])
	assert.Equal(t, "cid-leave", msg[keyCorrelationID])
}

func TestLeaveGame_RemovesClientFromRoom(t *testing.T) {
	svc := newService(t, nil)
	srv := newTestServer(t, svc)
	conn := dial(t, srv, "/ws", uuid.New())
	skipHello(t, conn)

	send(t, conn, map[string]string{keyType: msgJoinGame, keyGamePublicID: "game-bye"})
	readMsg(t, conn) // ack

	send(t, conn, map[string]string{keyType: msgLeaveGame, keyGamePublicID: "game-bye"})
	msg := readMsg(t, conn)
	assert.Equal(t, true, msg["success"])

	svc.mutex.RLock()
	_, stillExists := svc.gameClients["game-bye"]
	svc.mutex.RUnlock()
	assert.False(t, stillExists, "game room must be deleted when the last client leaves")
}

func TestLeaveGame_SpecificGameID_LeavesCorrectGame(t *testing.T) {
	// Client is in game-A; explicitly leaves game-B → must stay in game-A.
	svc := newService(t, nil)
	srv := newTestServer(t, svc)
	conn := dial(t, srv, "/ws", uuid.New())
	skipHello(t, conn)

	send(t, conn, map[string]string{keyType: msgJoinGame, keyGamePublicID: "game-A"})
	readMsg(t, conn)

	send(t, conn, map[string]string{keyType: msgLeaveGame, keyGamePublicID: "game-B"})
	msg := readMsg(t, conn)
	// game-B: client was never in it, handler will be called but game-B room is empty — fine.
	// More importantly, game-A must be untouched.
	_ = msg

	svc.mutex.RLock()
	inA := len(svc.gameClients["game-A"])
	svc.mutex.RUnlock()
	assert.Equal(t, 1, inA, "leaving game-B must not remove client from game-A")
}

// ─── submit_answer ────────────────────────────────────────────────────────────

func TestSubmitAnswer_NoGameHandler_AckError(t *testing.T) {
	svc := newService(t, nil) // no handler
	srv := newTestServer(t, svc)
	conn := dial(t, srv, "/ws", uuid.New())
	skipHello(t, conn)

	send(t, conn, map[string]interface{}{
		keyType:           msgSubmitAnswer,
		keyGamePublicID: testGameID1,
		keyAnswer:         map[string]interface{}{keyType: answerTypeMCQ, keyAnswer: "Paris"},
		keyCorrelationID: "cid-ans",
	})

	msg := readMsg(t, conn)
	assert.Equal(t, msgAck, msg[keyType])
	assert.Equal(t, msgSubmitAnswer, msg["action"])
	assert.Equal(t, false, msg["success"])
	assert.Equal(t, "cid-ans", msg[keyCorrelationID])
}

func TestSubmitAnswer_MissingGamePublicID_AckError(t *testing.T) {
	svc := newService(t, nil)
	svc.SetGameActionHandler(&mockGameHandler{})
	srv := newTestServer(t, svc)
	conn := dial(t, srv, "/ws", uuid.New())
	skipHello(t, conn)

	send(t, conn, map[string]interface{}{
		keyType:   msgSubmitAnswer,
		keyAnswer: map[string]interface{}{keyType: answerTypeMCQ, keyAnswer: "Paris"},
		// game_public_id intentionally absent
	})

	msg := readMsg(t, conn)
	assert.Equal(t, false, msg["success"])
	assert.Contains(t, msg["error"], "game_public_id required")
}

func TestSubmitAnswer_HandlerError_AckError(t *testing.T) {
	svc := newService(t, nil)
	svc.SetGameActionHandler(&mockGameHandler{
		submitAnswerErr: errors.New("already answered this question"),
	})
	srv := newTestServer(t, svc)
	conn := dial(t, srv, "/ws", uuid.New())
	skipHello(t, conn)

	send(t, conn, map[string]interface{}{
		keyType:           msgSubmitAnswer,
		keyGamePublicID: testGameID1,
		keyAnswer:         map[string]interface{}{keyType: answerTypeMCQ, keyAnswer: "Berlin"},
	})

	msg := readMsg(t, conn)
	assert.Equal(t, false, msg["success"])
	assert.Contains(t, msg["error"], "already answered")
}

// ─── Broadcasting ─────────────────────────────────────────────────────────────

func TestBroadcastToGame_ClientReceives(t *testing.T) {
	svc := newService(t, nil)
	srv := newTestServer(t, svc)
	conn := dial(t, srv, "/ws", uuid.New())
	skipHello(t, conn)

	send(t, conn, map[string]string{keyType: msgJoinGame, keyGamePublicID: "game-bcast"})
	readMsg(t, conn) // ack

	require.NoError(t, svc.SendToGame("game-bcast", map[string]interface{}{
		keyType:  "score_updated",
		"score": 99,
	}))

	msg := readMsg(t, conn)
	assert.Equal(t, "score_updated", msg[keyType])
	assert.Equal(t, float64(99), msg["score"])
}

func TestBroadcastToGame_ExcludesSpecifiedUser(t *testing.T) {
	// userA is excluded from the broadcast; userB must receive it; userA must not.
	svc := newService(t, nil)
	srv := newTestServer(t, svc)

	userA := uuid.New()
	userB := uuid.New()
	connA := dial(t, srv, "/ws", userA)
	connB := dial(t, srv, "/ws", userB)
	skipHello(t, connA)
	skipHello(t, connB)

	for _, tc := range []struct {
		conn *websocket.Conn
		uid  uuid.UUID
	}{{connA, userA}, {connB, userB}} {
		send(t, tc.conn, map[string]string{keyType: msgJoinGame, keyGamePublicID: "game-ex"})
		readMsg(t, tc.conn) // ack
	}

	require.NoError(t, svc.BroadcastToGame("game-ex", map[string]string{keyType: "event_x"}, &userA))

	// userB must receive the broadcast.
	msgB := readMsg(t, connB)
	assert.Equal(t, "event_x", msgB[keyType])

	// userA must NOT — use ping/pong to verify nothing is queued ahead of the pong.
	pingPong(t, connA)
}

func TestBroadcastToAll_ReachesEveryClient(t *testing.T) {
	svc := newService(t, nil)
	srv := newTestServer(t, svc)

	c1 := dial(t, srv, "/ws", uuid.New())
	c2 := dial(t, srv, "/ws", uuid.New())
	skipHello(t, c1)
	skipHello(t, c2)

	require.NoError(t, svc.BroadcastToAll(map[string]string{keyType: "lobby_stats"}))

	assert.Equal(t, "lobby_stats", readMsg(t, c1)[keyType])
	assert.Equal(t, "lobby_stats", readMsg(t, c2)[keyType])
}

func TestBroadcastToAdmins_OnlyAdminsReceive(t *testing.T) {
	svc := newService(t, nil)
	srv := newTestServer(t, svc)

	admin := uuid.New()
	regular := uuid.New()
	connAdmin := dial(t, srv, "/admin/ws", admin)
	connRegular := dial(t, srv, "/ws", regular)
	skipHello(t, connAdmin)
	skipHello(t, connRegular)

	require.NoError(t, svc.BroadcastToAdmins(map[string]string{keyType: "admin_event"}))

	// Admin receives it.
	assert.Equal(t, "admin_event", readMsg(t, connAdmin)[keyType])

	// Regular user must not — verify with ping/pong.
	pingPong(t, connRegular)
}

func TestSendToUser_WhenOffline_NoError(t *testing.T) {
	svc := newService(t, nil)
	err := svc.SendToUser(uuid.New(), map[string]string{keyType: "test"})
	assert.NoError(t, err, "sending to a disconnected user must not return an error")
}

func TestSendToUser_SameUser_TwoConnections_BothReceive(t *testing.T) {
	svc := newService(t, nil)
	srv := newTestServer(t, svc)
	uid := uuid.New()

	c1 := dial(t, srv, "/ws", uid)
	c2 := dial(t, srv, "/ws", uid)
	skipHello(t, c1)
	skipHello(t, c2)

	require.NoError(t, svc.SendToUser(uid, map[string]string{keyType: "direct_msg"}))

	assert.Equal(t, "direct_msg", readMsg(t, c1)[keyType])
	assert.Equal(t, "direct_msg", readMsg(t, c2)[keyType])
}

// ─── Rate limiting ────────────────────────────────────────────────────────────

func TestRateLimit_BlocksAfterLimit(t *testing.T) {
	cfg := model.DefaultWebSocketConfig()
	cfg.MessageRateLimit = 3
	cfg.MessageRateWindowSeconds = 60

	svc := newService(t, &cfg)
	srv := newTestServer(t, svc)
	conn := dial(t, srv, "/ws", uuid.New())
	skipHello(t, conn)

	// Fire 4 pings back-to-back.
	for range 4 {
		send(t, conn, map[string]string{keyType: msgPing})
	}

	pongs := 0
	limited := 0
	for range 4 {
		msg := readMsg(t, conn)
		switch msg[keyType] {
		case "pong":
			pongs++
		case msgAck:
			limited++
			assert.Equal(t, false, msg["success"])
			assert.Contains(t, msg["error"], "rate limit exceeded")
		}
	}

	assert.Equal(t, 3, pongs, "first 3 messages must pass the rate limit")
	assert.Equal(t, 1, limited, "4th message must be rate-limited")
}

func TestRateLimit_WindowReset_AllowsAgain(t *testing.T) {
	cfg := model.DefaultWebSocketConfig()
	cfg.MessageRateLimit = 2
	cfg.MessageRateWindowSeconds = 1 // 1-second window for test speed

	svc := newService(t, &cfg)
	srv := newTestServer(t, svc)
	conn := dial(t, srv, "/ws", uuid.New())
	skipHello(t, conn)

	// Use up the 2-message limit.
	for range 2 {
		send(t, conn, map[string]string{keyType: msgPing})
		assert.Equal(t, "pong", readMsg(t, conn)[keyType])
	}

	// 3rd message within the window must be rate-limited.
	send(t, conn, map[string]string{keyType: msgPing})
	msg := readMsg(t, conn)
	assert.Equal(t, msgAck, msg[keyType])
	assert.Equal(t, false, msg["success"])

	// Wait for the 1-second window to expire.
	time.Sleep(1100 * time.Millisecond)

	// Now the window has reset — should work again.
	send(t, conn, map[string]string{keyType: msgPing})
	assert.Equal(t, "pong", readMsg(t, conn)[keyType], "rate limit window must have reset")
}

// ─── trySend edge cases ───────────────────────────────────────────────────────

func TestTrySend_FullChannel_DropsWithoutBlocking(t *testing.T) {
	svc := newService(t, nil)
	client := &WSClient{
		ID:       uuid.New(),
		SendChan: make(chan []byte, 1),
	}
	client.SendChan <- []byte(`{keyType:"original"}`)

	done := make(chan struct{})
	go func() {
		svc.trySend(client, []byte(`{keyType:"dropped"}`))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("trySend blocked on a full channel — potential deadlock")
	}

	// Original message must still be there; the new one was dropped.
	assert.Equal(t, `{keyType:"original"}`, string(<-client.SendChan))
	assert.Equal(t, 0, len(client.SendChan))
}

func TestTrySend_ClosedClient_IgnoresMessage(t *testing.T) {
	svc := newService(t, nil)
	client := &WSClient{
		ID:       uuid.New(),
		SendChan: make(chan []byte, 10),
	}
	client.IsClosed.Store(true)

	svc.trySend(client, []byte(`{keyType:"test"}`))

	assert.Equal(t, 0, len(client.SendChan), "closed client must not receive any message")
}

// ─── Disconnect idempotency ───────────────────────────────────────────────────

func TestDisconnect_Idempotent_NoPanic(t *testing.T) {
	// disconnectClient is called from both handleConnection and writePump goroutines.
	// The second call must be a no-op and must not panic.
	svc := newService(t, nil)
	srv := newTestServer(t, svc)
	uid := uuid.New()

	conn := dial(t, srv, "/ws", uid)
	skipHello(t, conn)

	require.Eventually(t, func() bool { return svc.GetConnectedClients() == 1 }, time.Second, 10*time.Millisecond)

	_ = conn.CloseNow()

	// Wait for the server to process the disconnect (both goroutines will call disconnectClient).
	require.Eventually(t, func() bool { return svc.GetConnectedClients() == 0 }, time.Second, 10*time.Millisecond)

	// No panic means the test passed — the atomic CompareAndSwap guard worked.
}

// ─── Presence (online status) ─────────────────────────────────────────────────

func TestConnect_OnlineStatus_UpdatedOnConnect(t *testing.T) {
	updater := &mockStatusUpdater{}
	svc := NewWebSocketService(updater, &mockRedis{}, zap.NewNop())
	srv := newTestServer(t, svc)
	uid := uuid.New()

	conn := dial(t, srv, "/ws", uid)
	skipHello(t, conn)

	require.Eventually(t, func() bool {
		updater.mu.Lock()
		defer updater.mu.Unlock()
		for _, c := range updater.calls {
			if c.userID == uid && c.isOnline {
				return true
			}
		}
		return false
	}, time.Second, 10*time.Millisecond, "UpdateUserOnlineStatus(true) must be called on connect")
}

func TestConnect_OnlineStatus_UpdatedOnDisconnect(t *testing.T) {
	updater := &mockStatusUpdater{}
	svc := NewWebSocketService(updater, &mockRedis{}, zap.NewNop())
	srv := newTestServer(t, svc)
	uid := uuid.New()

	conn := dial(t, srv, "/ws", uid)
	skipHello(t, conn)
	_ = conn.CloseNow()

	require.Eventually(t, func() bool {
		updater.mu.Lock()
		defer updater.mu.Unlock()
		for _, c := range updater.calls {
			if c.userID == uid && !c.isOnline {
				return true
			}
		}
		return false
	}, 3*time.Second, 50*time.Millisecond, "UpdateUserOnlineStatus(false) must be called after disconnect")
}

func TestConnect_OnlineStatus_NotSetOffline_WhenSecondConnectionStillActive(t *testing.T) {
	// Same user connects from two devices. One disconnects.
	// The user must NOT be set offline because the second connection is still alive.
	updater := &mockStatusUpdater{}
	svc := NewWebSocketService(updater, &mockRedis{}, zap.NewNop())
	srv := newTestServer(t, svc)
	uid := uuid.New()

	c1 := dial(t, srv, "/ws", uid)
	c2 := dial(t, srv, "/ws", uid)
	skipHello(t, c1)
	skipHello(t, c2)

	_ = c1.CloseNow()
	time.Sleep(300 * time.Millisecond) // let the offline timer logic run

	updater.mu.Lock()
	wentOffline := false
	for _, c := range updater.calls {
		if c.userID == uid && !c.isOnline {
			wentOffline = true
		}
	}
	updater.mu.Unlock()

	assert.False(t, wentOffline, "user must not go offline when a second connection is still active")
	assert.Equal(t, 1, svc.GetConnectedClients())
}
