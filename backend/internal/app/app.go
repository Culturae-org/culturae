// backend/internal/app/app.go
//                ___________________________
// cmd/main.go -> | internal/app.SetupApp() | -> App.Run() -> Serveur HTTP Gin

package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"encoding/json"
	"io"

	"github.com/Culturae-org/culturae/internal/config"
	"github.com/Culturae-org/culturae/internal/database"
	"github.com/Culturae-org/culturae/internal/game"
	"github.com/Culturae-org/culturae/internal/handler"
	"github.com/Culturae-org/culturae/internal/handler/admin"
	"github.com/Culturae-org/culturae/internal/infrastructure/cache"
	"github.com/Culturae-org/culturae/internal/infrastructure/storage"
	"github.com/Culturae-org/culturae/internal/middleware"
	"github.com/Culturae-org/culturae/internal/model"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/repository"
	adminRepo "github.com/Culturae-org/culturae/internal/repository/admin"
	"github.com/Culturae-org/culturae/internal/routes"
	"github.com/Culturae-org/culturae/internal/service"
	"github.com/Culturae-org/culturae/internal/token"
	"github.com/Culturae-org/culturae/internal/usecase"
	adminUsecase "github.com/Culturae-org/culturae/internal/usecase/admin"
	"github.com/Culturae-org/culturae/internal/version"

	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Deps struct {
	RedisService        cache.RedisClientInterface
	MinIOService        storage.MinIOClientInterface
	UserCacheService    service.UserCacheServiceInterface
	SessionService      service.SessionServiceInterface
	WebSocketService    service.WebSocketServiceInterface
	JWTService          *token.JWTService
	SessionConfig       *model.SessionConfig
	PodDiscoveryService *service.PodDiscoveryService
}

type App struct {
	Config *config.Config
	Router *gin.Engine
	DB     *gorm.DB
	Server *http.Server
	Logger *zap.Logger

	Deps *Deps

	UserRepository      repository.UserRepositoryInterface
	SessionRepository   repository.SessionRepositoryInterface
	AdminUserRepository adminRepo.AdminUserRepositoryInterface
	AdminLogsRepository adminRepo.AdminLogsRepositoryInterface
	QuestionRepository  repository.QuestionRepositoryInterface
	FriendsRepository   repository.FriendsRepositoryInterface
	GameRepository      repository.GameRepositoryInterface
	GeographyRepository repository.GeographyRepositoryInterface

	UserUsecase               *usecase.UserUsecase
	AvatarUsecase             *usecase.AvatarUsecase
	AdminUserUsecase          *adminUsecase.AdminUserUsecase
	LogsUsecase               *adminUsecase.AdminLogsUsecase
	FriendsUsecase            *usecase.FriendsUsecase
	GameUsecase               *usecase.GameUsecase
	GeographyUsecase          *usecase.GeographyUsecase
	AdminDatasetsUsecase      *adminUsecase.AdminDatasetsUsecase
	AdminGameUsecase          *adminUsecase.AdminGameUsecase
	AdminGameTemplatesUsecase *adminUsecase.AdminGameTemplatesUsecase

	GameManager          game.GameManagerInterface
	MatchmakingService   *game.MatchmakingService
	GameEventBroadcaster *service.GameEventBroadcaster

	LobbyBroadcaster *service.LobbyBroadcaster

	appCtx    context.Context
	appCancel context.CancelFunc

	schedulerMu   sync.Mutex
	schedulerCron *cron.Cron

	AdminAuthHandler          *admin.AdminAuthHandler
	AdminUserHandler          *admin.AdminUserHandler
	AdminAvatarHandler        *admin.AdminAvatarHandler
	AdminQuestionHandler      *admin.AdminQuestionHandler
	AdminGeographyHandler     *admin.AdminGeographyHandler
	AdminGamesHandler         *admin.AdminGamesHandler
	AdminLogsHandler          *admin.AdminLogsHandler
	AdminImportsHandler       *admin.AdminImportsHandler
	AdminDatasetsHandler      *admin.AdminDatasetsHandler
	AdminGameTemplatesHandler *admin.AdminGameTemplatesHandler
	AdminFriendsHandler       *admin.AdminFriendsHandler
	AdminReportsHandler       *admin.AdminReportsHandler
	AdminMatchmakingHandler   *admin.AdminMatchmakingHandler
	AdminSettingsHandler      *admin.AdminSettingsHandler
	PodsHandler               *admin.PodsHandler

	AuthHandler          *handler.AuthHandler
	AvatarHandler        *handler.AvatarHandler
	UserHandler          *handler.UserHandler
	ProfileHandler       *handler.ProfileHandler
	FriendsHandler       *handler.FriendsHandler
	WebSocketHandler     *handler.WebSocketHandler
	AdminServicesHandler *admin.AdminServicesHandler
	GeographyHandler     *handler.GeographyHandler
	GamesHandler         *handler.GamesHandler
	ReportsHandler       *handler.ReportsHandler
	MatchmakingHandler   *handler.MatchmakingHandler
	LeaderboardHandler   *handler.LeaderboardHandler
	LobbyHandler         *handler.LobbyHandler
	HealthHandler        *handler.HealthHandler

	AuthMiddleware        *middleware.AuthMiddleware
	APILoggingMiddleware  *middleware.APILoggingMiddleware
	RateLimitMiddleware   *middleware.RateLimitMiddleware
	MaintenanceMiddleware *middleware.MaintenanceMiddleware
}

func SetupApp() (*App, error) {
	logger := setupLogger()

	cfg, err := setupConfig()
	if err != nil {
		logger.Error("Failed to setup config", zap.Error(err))
		return nil, fmt.Errorf("failed to setup config: %w", err)
	}

	db, err := setupDatabase(cfg)
	if err != nil {
		logger.Error("Failed to setup database", zap.Error(err))
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	dependencies, err := setupDependencies(cfg, db, logger)
	if err != nil {
		logger.Error("Failed to setup dependencies", zap.Error(err))
		return nil, fmt.Errorf("failed to setup dependencies: %w", err)
	}

	appCtx, appCancel := context.WithCancel(context.Background())

	app, err := buildApp(appCtx, cfg, db, dependencies, logger)
	if err != nil {
		appCancel()
		logger.Error("Failed to build app", zap.Error(err))
		return nil, fmt.Errorf("failed to build app: %w", err)
	}

	app.appCancel = appCancel
	app.Logger = logger

	httputil.SetLogger(logger)

	if err := app.initialize(); err != nil {
		logger.Error("Failed to initialize app", zap.Error(err))
		return nil, fmt.Errorf("failed to initialize app: %w", err)
	}

	logger.Info("--------------------------------------------------")
	logger.Info("             Culturae Platform OK !             ")
	logger.Info(fmt.Sprintf("             Mode: %s             ", cfg.AppMode))
	logger.Info("--------------------------------------------------")
	return app, nil
}

func (a *App) GetConfig() *config.Config {
	return a.Config
}

func (a *App) GetAuthMiddleware() *middleware.AuthMiddleware {
	return a.AuthMiddleware
}

func (a *App) GetRateLimitMiddleware() *middleware.RateLimitMiddleware {
	return a.RateLimitMiddleware
}

func (a *App) GetAuthHandler() *handler.AuthHandler {
	return a.AuthHandler
}

func (a *App) GetAvatarHandler() *handler.AvatarHandler {
	return a.AvatarHandler
}

func (a *App) GetUserHandler() *handler.UserHandler {
	return a.UserHandler
}

func (a *App) GetProfileHandler() *handler.ProfileHandler {
	return a.ProfileHandler
}

func (a *App) GetFriendsHandler() *handler.FriendsHandler {
	return a.FriendsHandler
}

func (a *App) GetWebSocketHandler() *handler.WebSocketHandler {
	return a.WebSocketHandler
}

func (a *App) GetLogsHandler() *admin.AdminLogsHandler {
	return a.AdminLogsHandler
}

func (a *App) GetServicesHandler() *admin.AdminServicesHandler {
	return a.AdminServicesHandler
}

func (a *App) GetImportsHandler() *admin.AdminImportsHandler {
	return a.AdminImportsHandler
}

func (a *App) GetGeographyHandler() *handler.GeographyHandler {
	return a.GeographyHandler
}

func (a *App) GetGamesHandler() *handler.GamesHandler {
	return a.GamesHandler
}

func (a *App) GetAdminFriendsHandler() *admin.AdminFriendsHandler {
	return a.AdminFriendsHandler
}

func (a *App) GetAdminQuestionHandler() *admin.AdminQuestionHandler {
	return a.AdminQuestionHandler
}

func (a *App) GetAdminGeographyHandler() *admin.AdminGeographyHandler {
	return a.AdminGeographyHandler
}

func (a *App) GetAdminGamesHandler() *admin.AdminGamesHandler {
	return a.AdminGamesHandler
}

func (a *App) GetAdminAuthHandler() *admin.AdminAuthHandler {
	return a.AdminAuthHandler
}

func (a *App) GetAdminUserHandler() *admin.AdminUserHandler {
	return a.AdminUserHandler
}

func (a *App) GetAdminAvatarHandler() *admin.AdminAvatarHandler {
	return a.AdminAvatarHandler
}

func (a *App) GetAdminDatasetsHandler() *admin.AdminDatasetsHandler {
	return a.AdminDatasetsHandler
}

func (a *App) GetReportsHandler() *handler.ReportsHandler {
	return a.ReportsHandler
}

func (a *App) GetMatchmakingHandler() *handler.MatchmakingHandler {
	return a.MatchmakingHandler
}

func (a *App) GetLeaderboardHandler() *handler.LeaderboardHandler {
	return a.LeaderboardHandler
}

func (a *App) GetAdminReportsHandler() *admin.AdminReportsHandler {
	return a.AdminReportsHandler
}

func (a *App) GetAdminMatchmakingHandler() *admin.AdminMatchmakingHandler {
	return a.AdminMatchmakingHandler
}

func (a *App) GetAdminSettingsHandler() *admin.AdminSettingsHandler {
	return a.AdminSettingsHandler
}

func (a *App) GetPodsHandler() *admin.PodsHandler {
	return a.PodsHandler
}

func (a *App) GetMaintenanceMiddleware() *middleware.MaintenanceMiddleware {
	return a.MaintenanceMiddleware
}

func (a *App) GetAdminGameTemplatesHandler() *admin.AdminGameTemplatesHandler {
	return a.AdminGameTemplatesHandler
}

func (a *App) GetLobbyHandler() *handler.LobbyHandler {
	return a.LobbyHandler
}

func (a *App) GetHealthHandler() *handler.HealthHandler {
	return a.HealthHandler
}

func (a *App) Run() {
	a.Server = &http.Server{
		Addr:         ":" + a.Config.ServerPort,
		Handler:      a.Router,
		ReadTimeout:  a.Config.ReadTimeout,
		WriteTimeout: a.Config.WriteTimeout,
		IdleTimeout:  a.Config.IdleTimeout,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		a.Logger.Info("Server starting", zap.String("port", a.Config.ServerPort))
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	<-quit
	a.Logger.Info("Shutting down server...")

	a.Shutdown()
}

func (a *App) Shutdown() {
	if a.appCancel != nil {
		a.appCancel()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if a.Server != nil {
		if err := a.Server.Shutdown(ctx); err != nil {
			a.Logger.Error("Server forced to shutdown", zap.Error(err))
		}
	}

	a.Close()
}

func (a *App) Close() {

	if a.Deps.WebSocketService != nil {
		a.Deps.WebSocketService.StopRelay()
		a.Logger.Info("WebSocket Pub/Sub relay stopped")
	}

	if a.Deps.PodDiscoveryService != nil {
		a.Deps.PodDiscoveryService.Stop()
		a.Logger.Info("Pod discovery service stopped")
	}

	if a.GameEventBroadcaster != nil {
		a.GameEventBroadcaster.Stop()
		a.Logger.Info("Game event broadcaster stopped")
	}

	if a.LobbyBroadcaster != nil {
		a.LobbyBroadcaster.Stop()
		a.Logger.Info("Lobby broadcaster stopped")
	}

	if a.Deps.RedisService != nil {
		if err := a.Deps.RedisService.Close(); err != nil {
			a.Logger.Error("Error closing Redis connection", zap.Error(err))
		}
	}

	if a.DB != nil {
		sqlDB, err := a.DB.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				a.Logger.Error("Error closing database connection", zap.Error(err))
			}
		}
	}

	a.Logger.Info("All connections closed successfully")
}

func (a *App) StartQuestionTimeoutChecker() {
	rgm, ok := a.GameManager.(*game.RedisGameManager)
	if !ok {
		a.Logger.Warn("GameManager is not a *RedisGameManager — question timeout checker not started")
		return
	}
	rgm.StartQuestionTimeoutChecker(a.appCtx)
	a.Logger.Info("Question timeout checker started")
}

func (a *App) RestartSchedulers() {
	a.schedulerMu.Lock()
	defer a.schedulerMu.Unlock()

	if a.schedulerCron != nil {
		a.schedulerCron.Stop()
	}

	cfg := a.loadSystemConfig()
	c := cron.New()

	if cfg.GameCleanupEnabled && a.AdminGameUsecase != nil {
		expr := cfg.GameCleanupCron
		if expr == "" {
			expr = defaultCronExpressionEvery5Minutes
		}
		if _, err := c.AddFunc(expr, func() {
			if _, err := a.AdminGameUsecase.CleanupAbandonedGames(uuid.Nil); err != nil {
				a.Logger.Error("Game cleanup error", zap.Error(err))
			}
		}); err != nil {
			a.Logger.Error("Invalid game cleanup cron expression", zap.String("expr", expr), zap.Error(err))
		} else {
			a.Logger.Info("Game cleanup scheduler registered", zap.String("cron", expr))
		}
	}

	if cfg.SessionCleanupEnabled && a.Deps.SessionService != nil {
		expr := cfg.SessionCleanupCron
		if expr == "" {
			expr = defaultCronExpressionHourly
		}
		if _, err := c.AddFunc(expr, func() {
			if err := a.Deps.SessionService.CleanupExpiredSessions(); err != nil {
				a.Logger.Error("Session cleanup error", zap.Error(err))
			}
		}); err != nil {
			a.Logger.Error("Invalid session cleanup cron expression", zap.String("expr", expr), zap.Error(err))
		} else {
			a.Logger.Info("Session cleanup scheduler registered", zap.String("cron", expr))
		}
	}

	if cfg.DatasetCheckEnabled && a.AdminDatasetsUsecase != nil {
		expr := cfg.DatasetCheckCron
		if expr == "" {
			expr = defaultCronExpressionHourly
		}
		if _, err := c.AddFunc(expr, func() {
			a.runDatasetUpdateCheck()
		}); err != nil {
			a.Logger.Error("Invalid dataset check cron expression", zap.String("expr", expr), zap.Error(err))
		} else {
			a.Logger.Info("Dataset update checker registered", zap.String("cron", expr))
		}
	}

	if cfg.VersionCheckEnabled && a.Deps.RedisService != nil {
		expr := cfg.VersionCheckCron
		if expr == "" {
			expr = defaultCronExpressionHourly
		}
		if _, err := c.AddFunc(expr, func() {
			a.runVersionCheck()
		}); err != nil {
			a.Logger.Error("Invalid version check cron expression", zap.String("expr", expr), zap.Error(err))
		} else {
			a.Logger.Info("Version checker registered", zap.String("cron", expr))
			go a.runVersionCheck()
		}
	}

	c.Start()
	a.schedulerCron = c
}

func (a *App) loadSystemConfig() model.SystemConfig {
	var cfg model.SystemConfig
	if a.Deps.RedisService == nil {
		return model.DefaultSystemConfig()
	}
	cfgCtx, cfgCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cfgCancel()
	if err := a.Deps.RedisService.GetJSON(cfgCtx, "system:config", &cfg); err != nil ||
		cfg.UserCacheTTLMinutes <= 0 {
		return model.DefaultSystemConfig()
	}
	return cfg
}

func (a *App) runDatasetUpdateCheck() {
	if _, err := a.AdminDatasetsUsecase.CheckAllForUpdates(); err != nil {
		a.Logger.Error("Dataset update check failed", zap.Error(err))
	}
}

const githubReleaseURL = "https://api.github.com/repos/Culturae-org/culturae/releases/latest"
const versionStatusKey = "system:version:status"
const defaultCronExpressionHourly = "0 * * * *"
const defaultCronExpressionEvery5Minutes = "*/5 * * * *"

type versionStatus struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	IsUpToDate     bool   `json:"is_up_to_date"`
	CheckedAt      string `json:"checked_at"`
}

func (a *App) runVersionCheck() {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, githubReleaseURL, nil)
	if err != nil {
		a.Logger.Warn("Version check: failed to build request", zap.Error(err))
		return
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "culturae-server")

	resp, err := client.Do(req)
	if err != nil {
		a.Logger.Warn("Version check: GitHub API unreachable", zap.Error(err))
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			a.Logger.Warn("Version check: failed to close response body", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		a.Logger.Warn("Version check: unexpected status", zap.Int("status", resp.StatusCode))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		a.Logger.Warn("Version check: failed to read response", zap.Error(err))
		return
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &release); err != nil || release.TagName == "" {
		a.Logger.Warn("Version check: failed to parse response", zap.Error(err))
		return
	}

	status := versionStatus{
		CurrentVersion: version.Version,
		LatestVersion:  release.TagName,
		IsUpToDate:     version.Version == release.TagName,
		CheckedAt:      time.Now().UTC().Format(time.RFC3339),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.Deps.RedisService.SetJSON(ctx, versionStatusKey, status, 2*time.Hour); err != nil {
		a.Logger.Warn("Version check: failed to store status in Redis", zap.Error(err))
		return
	}

	if !status.IsUpToDate {
		a.Logger.Info("New version available",
			zap.String("current", status.CurrentVersion),
			zap.String("latest", status.LatestVersion),
		)
	}

	a.Deps.WebSocketService.BroadcastAdminNotification(service.AdminNotification{
		Event: "version_status_updated",
		Data: map[string]interface{}{
			"current_version": status.CurrentVersion,
			"latest_version":  status.LatestVersion,
			"is_up_to_date":   status.IsUpToDate,
			"checked_at":      status.CheckedAt,
		},
	})
}

func setupLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		panic("failed to initialize zap logger: " + err.Error())
	}
	return logger
}

func setupConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}

func setupDatabase(cfg *config.Config) (*gorm.DB, error) {
	db, err := database.Connect(cfg.GetPostgresDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if err := database.RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	return db, nil
}

func setupDependencies(cfg *config.Config, db *gorm.DB, logger *zap.Logger) (*Deps, error) {
	services, err := setupServices(cfg, db, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to setup services: %w", err)
	}

	deps := &Deps{
		RedisService:     services.RedisService,
		MinIOService:     services.MinIOService,
		UserCacheService: services.UserCacheService,
		SessionService:   services.SessionService,
		WebSocketService: services.WebSocketService,
		JWTService:       services.JWTService,
		SessionConfig:    setupSessionConfig(),
	}

	return deps, nil
}

type AppServices struct {
	RedisService     cache.RedisClientInterface
	MinIOService     storage.MinIOClientInterface
	UserCacheService service.UserCacheServiceInterface
	SessionService   service.SessionServiceInterface
	WebSocketService service.WebSocketServiceInterface
	JWTService       *token.JWTService
}

func setupServices(cfg *config.Config, db *gorm.DB, logger *zap.Logger) (*AppServices, error) {
	redisService := cache.NewRedisClient(cfg.GetRedisAddr(), cfg.RedisPassword, cfg.RedisDB, logger)
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := redisService.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	minioService, err := storage.NewMinIOClient(
		cfg.MinIOEndpoint,
		cfg.MinIOAccessKey,
		cfg.MinIOSecretKey,
		cfg.MinIOBucketName,
		cfg.MinIOUseSSL,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MinIO service: %w", err)
	}

	userCacheService := service.NewUserCacheService(
		redisService,
	)

	sessionConfig := setupSessionConfig()

	jwtService := token.NewJWTService(
		cfg.JWTSecret,
	)

	userRepository := repository.NewUserRepository(
		db,
		userCacheService,
		logger,
	)

	sessionRepository := repository.NewSessionRepository(
		db,
		logger,
	)

	sessionService := service.NewSessionService(
		sessionRepository,
		redisService,
		sessionConfig,
		logger,
	)

	webSocketService := service.NewWebSocketService(
		userRepository,
		redisService,
		logger,
	)

	return &AppServices{
		RedisService:     redisService,
		MinIOService:     minioService,
		UserCacheService: userCacheService,
		SessionService:   sessionService,
		WebSocketService: webSocketService,
		JWTService:       jwtService,
	}, nil
}

func setupSessionConfig() *model.SessionConfig {
	sessionConfig := model.DefaultSessionConfig()
	sessionConfig.AccessTokenDuration = time.Minute * 15
	sessionConfig.RefreshTokenDuration = time.Hour * 24 * 7
	sessionConfig.MaxActiveSessions = 5
	return sessionConfig
}

type AppComponents struct {
	Repositories struct {
		User            repository.UserRepositoryInterface
		Session         repository.SessionRepositoryInterface
		Logs            adminRepo.AdminLogsRepositoryInterface
		Imports         adminRepo.ImportJobRepositoryInterface
		Metrics         adminRepo.MetricsRepositoryInterface
		Question        repository.QuestionRepositoryInterface
		QuestionDataset repository.QuestionDatasetRepositoryInterface
		AdminQuestion   adminRepo.AdminQuestionRepositoryInterface
		AdminDataset    adminRepo.AdminDatasetRepositoryInterface
		Friends         repository.FriendsRepositoryInterface
		Game            repository.GameRepositoryInterface
		AdminGame       adminRepo.AdminGameRepositoryInterface
		Geography       repository.GeographyRepositoryInterface
		AdminGeography  adminRepo.AdminGeographyRepositoryInterface
		Report          repository.ReportRepositoryInterface
		AdminReport     adminRepo.AdminReportRepositoryInterface
		GameEventLog    adminRepo.GameEventLogRepositoryInterface
		AdminUser       adminRepo.AdminUserRepositoryInterface
		AdminFriends    adminRepo.AdminFriendsRepositoryInterface
		GameTemplate    repository.GameTemplateRepositoryInterface
		AvatarStorage   repository.AvatarStorageRepositoryInterface
		AvatarCache     repository.AvatarCacheRepositoryInterface
		FlagStorage     repository.FlagStorageRepositoryInterface
		CacheHealth     repository.CacheHealthRepositoryInterface
		StorageHealth   repository.StorageHealthRepositoryInterface
	}
	UseCases struct {
		User               *usecase.UserUsecase
		Avatar             *usecase.AvatarUsecase
		Friends            *usecase.FriendsUsecase
		Game               *usecase.GameUsecase
		Geography          *usecase.GeographyUsecase
		Report             *usecase.ReportUsecase
		Admin              *adminUsecase.AdminUserUsecase
		AdminImports       *adminUsecase.AdminImportsUsecase
		AdminLogs          *adminUsecase.AdminLogsUsecase
		AdminGeography     *adminUsecase.AdminGeographyUsecase
		AdminDatasets      *adminUsecase.AdminDatasetsUsecase
		AdminQuestions     *adminUsecase.AdminQuestionsUsecase
		AdminReport        *adminUsecase.AdminReportUsecase
		AdminGame          *adminUsecase.AdminGameUsecase
		AdminGameTemplates *adminUsecase.AdminGameTemplatesUsecase
		AdminFriends       *adminUsecase.AdminFriendsUsecase
	}
	Handlers struct {
		Auth               *handler.AuthHandler
		Avatar             *handler.AvatarHandler
		User               *handler.UserHandler
		Profile            *handler.ProfileHandler
		WebSocket          *handler.WebSocketHandler
		Services           *admin.AdminServicesHandler
		Friends            *handler.FriendsHandler
		Games              *handler.GamesHandler
		Geography          *handler.GeographyHandler
		Reports            *handler.ReportsHandler
		Matchmaking        *handler.MatchmakingHandler
		Leaderboard        *handler.LeaderboardHandler
		Lobby              *handler.LobbyHandler
		Health             *handler.HealthHandler
		AdminAuth          *admin.AdminAuthHandler
		AdminUser          *admin.AdminUserHandler
		AdminAvatar        *admin.AdminAvatarHandler
		AdminQuestion      *admin.AdminQuestionHandler
		AdminGeography     *admin.AdminGeographyHandler
		AdminGames         *admin.AdminGamesHandler
		AdminImports       *admin.AdminImportsHandler
		AdminDatasets      *admin.AdminDatasetsHandler
		AdminGameTemplates *admin.AdminGameTemplatesHandler
		AdminLogs          *admin.AdminLogsHandler
		AdminReports       *admin.AdminReportsHandler
		AdminMatchmaking   *admin.AdminMatchmakingHandler
		AdminSettings      *admin.AdminSettingsHandler
		AdminFriends       *admin.AdminFriendsHandler
		Pods               *admin.PodsHandler
	}
	Middlewares struct {
		Auth        *middleware.AuthMiddleware
		APILogging  *middleware.APILoggingMiddleware
		RateLimit   *middleware.RateLimitMiddleware
		Maintenance *middleware.MaintenanceMiddleware
	}
	GameManager          game.GameManagerInterface
	MatchmakingService   *game.MatchmakingService
	GameEventBroadcaster *service.GameEventBroadcaster
	LobbyBroadcaster     *service.LobbyBroadcaster
	LoggingService       service.LoggingServiceInterface
}

func buildApp(appCtx context.Context, cfg *config.Config, db *gorm.DB, deps *Deps, logger *zap.Logger) (*App, error) {

	loggingRepo := adminRepo.NewServiceLoggingRepository(db)
	loggingService := service.NewLoggingService(loggingRepo, deps.RedisService, deps.MinIOService, logger)

	components := &AppComponents{LoggingService: loggingService}

	if err := setupRepositoriesInComponents(appCtx, db, deps.UserCacheService, deps.WebSocketService, deps.RedisService, components, logger); err != nil {
		return nil, fmt.Errorf("failed to setup repositories: %w", err)
	}

	setupUseCasesInComponents(components, db, deps.RedisService, deps.MinIOService, deps.WebSocketService, logger)

	setupHandlersAndMiddlewares(cfg, deps, components, logger)

	router := setupRouter(cfg, components.Middlewares.APILogging)

	app := &App{
		appCtx:                    appCtx,
		Config:                    cfg,
		Router:                    router,
		DB:                        db,
		Deps:                      deps,
		UserRepository:            components.Repositories.User,
		SessionRepository:         components.Repositories.Session,
		AdminUserRepository:       components.Repositories.AdminUser,
		AdminLogsRepository:       components.Repositories.Logs,
		QuestionRepository:        components.Repositories.Question,
		FriendsRepository:         components.Repositories.Friends,
		GameRepository:            components.Repositories.Game,
		GeographyRepository:       components.Repositories.Geography,
		UserUsecase:               components.UseCases.User,
		AdminUserUsecase:          components.UseCases.Admin,
		LogsUsecase:               components.UseCases.AdminLogs,
		FriendsUsecase:            components.UseCases.Friends,
		GameUsecase:               components.UseCases.Game,
		GeographyUsecase:          components.UseCases.Geography,
		AdminDatasetsUsecase:      components.UseCases.AdminDatasets,
		AdminGameUsecase:          components.UseCases.AdminGame,
		AdminGameTemplatesUsecase: components.UseCases.AdminGameTemplates,
		GameManager:               components.GameManager,
		MatchmakingService:        components.MatchmakingService,
		GameEventBroadcaster:      components.GameEventBroadcaster,
		AuthHandler:               components.Handlers.Auth,
		AvatarHandler:             components.Handlers.Avatar,
		UserHandler:               components.Handlers.User,
		ProfileHandler:            components.Handlers.Profile,
		FriendsHandler:            components.Handlers.Friends,
		GamesHandler:              components.Handlers.Games,
		AdminAuthHandler:          components.Handlers.AdminAuth,
		AdminUserHandler:          components.Handlers.AdminUser,
		AdminAvatarHandler:        components.Handlers.AdminAvatar,
		AdminQuestionHandler:      components.Handlers.AdminQuestion,
		AdminGeographyHandler:     components.Handlers.AdminGeography,
		AdminGamesHandler:         components.Handlers.AdminGames,
		AdminImportsHandler:       components.Handlers.AdminImports,
		AdminLogsHandler:          components.Handlers.AdminLogs,
		AdminDatasetsHandler:      components.Handlers.AdminDatasets,
		AdminGameTemplatesHandler: components.Handlers.AdminGameTemplates,
		GeographyHandler:          components.Handlers.Geography,
		WebSocketHandler:          components.Handlers.WebSocket,
		ReportsHandler:            components.Handlers.Reports,
		MatchmakingHandler:        components.Handlers.Matchmaking,
		LeaderboardHandler:        components.Handlers.Leaderboard,
		LobbyHandler:              components.Handlers.Lobby,
		HealthHandler:             components.Handlers.Health,
		LobbyBroadcaster:          components.LobbyBroadcaster,
		AdminFriendsHandler:       components.Handlers.AdminFriends,
		AdminReportsHandler:       components.Handlers.AdminReports,
		AdminMatchmakingHandler:   components.Handlers.AdminMatchmaking,
		AdminSettingsHandler:      components.Handlers.AdminSettings,
		AdminServicesHandler:      components.Handlers.Services,
		PodsHandler:               components.Handlers.Pods,
		AuthMiddleware:            components.Middlewares.Auth,
		APILoggingMiddleware:      components.Middlewares.APILogging,
		RateLimitMiddleware:       components.Middlewares.RateLimit,
		MaintenanceMiddleware:     components.Middlewares.Maintenance,
	}

	if app.AdminSettingsHandler != nil {
		app.AdminSettingsHandler.RestartSchedulers = app.RestartSchedulers
	}

	return app, nil
}

func setupRepositoriesInComponents(
	appCtx context.Context,
	db *gorm.DB,
	userCacheService service.UserCacheServiceInterface,
	wsService service.WebSocketServiceInterface,
	redisService cache.RedisClientInterface,
	components *AppComponents,
	logger *zap.Logger,
) error {
	userRepo, sessionRepo, adminUserRepo, logsRepo, questionRepo, friendsRepo, gameRepo, err := setupRepositories(db, userCacheService, logger)
	if err != nil {
		return err
	}
	components.Repositories.User = userRepo
	components.Repositories.Session = sessionRepo
	components.Repositories.AdminUser = adminUserRepo
	components.Repositories.Logs = logsRepo
	components.Repositories.Question = questionRepo
	components.Repositories.QuestionDataset = repository.NewQuestionDatasetRepository(db)
	components.Repositories.AdminQuestion = adminRepo.NewAdminQuestionRepository(db)
	components.Repositories.AdminDataset = adminRepo.NewAdminDatasetRepository(db)
	components.Repositories.Friends = friendsRepo
	components.Repositories.Game = gameRepo
	components.Repositories.AdminFriends = adminRepo.NewAdminFriendsRepository(db)
	concreteGameRepo, ok := gameRepo.(*repository.GameRepository)
	if !ok {
		return fmt.Errorf("gameRepo is not *repository.GameRepository")
	}
	components.Repositories.AdminGame = concreteGameRepo
	components.Repositories.Geography = repository.NewGeographyRepository(db)
	components.Repositories.AdminGeography = adminRepo.NewAdminGeographyRepository(db)

	reportRepo := repository.NewReportRepository(db)
	components.Repositories.Report = reportRepo
	components.Repositories.AdminReport = reportRepo
	components.Repositories.GameEventLog = adminRepo.NewGameEventLogRepository(db)
	components.Repositories.GameTemplate = repository.NewGameTemplateRepository(db)

	components.GameManager = game.NewRedisGameManager(
		appCtx,
		redisService,
		components.Repositories.User,
		components.Repositories.Game,
		components.Repositories.Logs,
		logger,
	)

	gameEventChan := components.GameManager.GetEventChannel()
	serviceEventChan := make(chan service.RawGameEvent, 100)

	go func() {
		for gameEvent := range gameEventChan {
			serviceEventChan <- service.RawGameEvent{
				Type:      gameEvent.Type,
				GameID:    gameEvent.GameID,
				PublicID:  gameEvent.PublicID,
				Data:      gameEvent.Data,
				Timestamp: gameEvent.Timestamp,
			}
		}
		close(serviceEventChan)
	}()

	components.GameEventBroadcaster = service.NewGameEventBroadcaster(
		serviceEventChan,
		wsService,
		logger,
	)
	components.GameEventBroadcaster.SetEventLogRepo(components.Repositories.GameEventLog)

	components.MatchmakingService = game.NewMatchmakingService(
		appCtx,
		redisService,
		components.Repositories.User,
		logger,
	)
	components.MatchmakingService.SetUserNotifier(wsService)
	if rgm, ok := components.GameManager.(*game.RedisGameManager); ok {
		rgm.SetUserNotifier(wsService)
	}

	components.LobbyBroadcaster = service.NewLobbyBroadcaster(
		components.GameManager,
		components.MatchmakingService,
		wsService,
		logger,
	)

	return nil
}

func setupUseCasesInComponents(
	components *AppComponents,
	db *gorm.DB,
	redisService cache.RedisClientInterface,
	minioService storage.MinIOClientInterface,
	wsService service.WebSocketServiceInterface,
	logger *zap.Logger,
) {
	components.UseCases.User = usecase.NewUserUsecase(
		components.Repositories.User,
		components.Repositories.User,
		components.Repositories.Session,
	)

	components.Repositories.AvatarStorage = repository.NewAvatarStorageAdapter(minioService)
	components.Repositories.AvatarCache = repository.NewAvatarCacheAdapter(redisService)

	components.UseCases.Avatar = usecase.NewAvatarUsecase(
		components.Repositories.User,
		components.Repositories.AvatarStorage,
		components.Repositories.AvatarCache,
		components.LoggingService,
	)

	components.UseCases.Admin = adminUsecase.NewAdminUserUsecase(
		components.Repositories.AdminUser,
		components.Repositories.Session,
		components.LoggingService,
	)

	components.Repositories.Imports = adminRepo.NewImportJobRepository(db)
	components.Repositories.Metrics = adminRepo.NewMetricsRepository(db)
	components.Repositories.FlagStorage = repository.NewFlagStorageAdapter(minioService, logger)
	components.Repositories.CacheHealth = repository.NewCacheHealthAdapter(redisService)
	components.Repositories.StorageHealth = repository.NewStorageHealthAdapter(minioService)

	components.UseCases.AdminLogs = adminUsecase.NewAdminLogsUsecase(
		components.Repositories.Logs,
		components.Repositories.Metrics,
		components.Repositories.CacheHealth,
		components.Repositories.StorageHealth,
	)

	components.UseCases.AdminQuestions = adminUsecase.NewAdminQuestionsUsecase(
		components.Repositories.Question,
		components.Repositories.AdminQuestion,
		components.Repositories.AdminDataset,
		logger,
		components.Repositories.Imports,
	)

	components.UseCases.Friends = usecase.NewFriendsUsecase(
		components.Repositories.Friends,
		components.Repositories.User,
		components.LoggingService,
		wsService,
	)

	components.UseCases.Geography = usecase.NewGeographyUsecase(
		components.Repositories.Geography,
		logger,
		components.Repositories.FlagStorage,
	)

	components.UseCases.AdminGeography = adminUsecase.NewAdminGeographyUsecase(
		components.Repositories.AdminGeography,
		components.Repositories.Geography,
		logger,
		components.Repositories.FlagStorage,
		components.Repositories.Imports,
	)

	components.UseCases.AdminImports = adminUsecase.NewAdminImportsUsecase(
		logger,
		components.Repositories.Imports,
		components.Repositories.User,
		components.Repositories.Session,
		components.UseCases.AdminQuestions,
		components.UseCases.AdminGeography,
	)

	components.UseCases.Report = usecase.NewReportUsecase(
		components.Repositories.Report,
		components.Repositories.Game,
		wsService,
	)

	components.UseCases.AdminReport = adminUsecase.NewAdminReportUsecase(
		components.Repositories.AdminReport,
		wsService,
	)

	components.UseCases.AdminDatasets = adminUsecase.NewAdminDatasetsUsecase(
		components.Repositories.QuestionDataset,
		components.Repositories.Question,
		components.Repositories.AdminQuestion,
		components.Repositories.AdminGeography,
		components.Repositories.Logs,
		logger,
	)

	components.UseCases.AdminGeography.SetAdminDatasetsUsecase(components.UseCases.AdminDatasets)

	datasetReader := usecase.NewDatasetReaderAdapter(components.UseCases.AdminDatasets)

	xpCalc := game.NewXPCalculator(redisService)
	eloCalc := game.NewELOCalculator(redisService)

	components.UseCases.Game = usecase.NewGameUsecase(
		components.Repositories.Game,
		components.Repositories.Question,
		components.Repositories.AdminQuestion,
		components.Repositories.Geography,
		components.Repositories.User,
		components.Repositories.Friends,
		components.Repositories.GameTemplate,
		components.GameManager,
		components.LoggingService,
		wsService,
		datasetReader,
		xpCalc,
		eloCalc,
		logger,
	)
	components.UseCases.AdminGame = adminUsecase.NewAdminGameUsecase(
		components.Repositories.AdminGame,
		components.Repositories.User,
		components.Repositories.Question,
		components.GameManager,
		components.LoggingService,
		logger,
	)
	components.UseCases.AdminGame.SetEventLogRepo(components.Repositories.GameEventLog)

	components.UseCases.AdminGameTemplates = adminUsecase.NewAdminGameTemplatesUsecase(
		components.Repositories.GameTemplate,
		components.LoggingService,
		logger,
	)
	components.MatchmakingService.SetMatchFoundCallback(components.UseCases.Game.CreateMatchmakedGame)
}

func setupHandlersAndMiddlewares(
	cfg *config.Config,
	deps *Deps,
	components *AppComponents,
	logger *zap.Logger,
) {

	components.Handlers.Auth = handler.NewAuthHandler(
		cfg,
		components.UseCases.User,
		cfg.JWTSecret,
		deps.RedisService,
		deps.SessionService,
		deps.JWTService,
		deps.SessionConfig,
		components.LoggingService,
		deps.WebSocketService,
		logger,
	)

	components.Handlers.Avatar = handler.NewAvatarHandler(
		components.UseCases.User,
		components.UseCases.Avatar,
		components.UseCases.Friends,
		components.LoggingService,
		logger,
	)

	components.Handlers.Profile = handler.NewProfileHandler(
		components.UseCases.User,
		components.UseCases.Game,
		deps.SessionService,
		deps.SessionConfig,
		components.LoggingService,
	)

	components.Handlers.User = handler.NewUserHandler(
		components.UseCases.User,
		components.UseCases.Friends,
	)

	components.Handlers.AdminAuth = admin.NewAdminAuthHandler(
		cfg,
		components.UseCases.User,
		cfg.JWTSecret,
		deps.RedisService,
		deps.SessionService,
		deps.JWTService,
		deps.SessionConfig,
		components.LoggingService,
	)

	components.UseCases.AdminFriends = adminUsecase.NewAdminFriendsUsecase(components.Repositories.AdminFriends)
	components.Handlers.AdminFriends = admin.NewAdminFriendsHandler(components.UseCases.AdminFriends)

	components.Handlers.AdminUser = admin.NewAdminUserHandler(
		components.UseCases.Admin,
		components.UseCases.User,
		components.UseCases.Game,
		components.LoggingService,
		deps.WebSocketService,
	)

	components.Handlers.AdminAvatar = admin.NewAdminAvatarHandler(
		components.UseCases.User,
		components.UseCases.Avatar,
		components.LoggingService,
	)

	components.Handlers.AdminGeography = admin.NewAdminGeographyHandler(
		components.UseCases.AdminGeography,
		components.UseCases.Geography,
		components.LoggingService,
		deps.WebSocketService,
		logger,
	)

	components.Handlers.AdminGames = admin.NewAdminGamesHandler(
		components.UseCases.AdminGame,
		components.LoggingService,
		logger,
	)

	components.Handlers.AdminReports = admin.NewAdminReportsHandler(
		components.UseCases.AdminReport,
		components.LoggingService,
		logger,
	)

	components.Handlers.AdminLogs = admin.NewAdminLogsHandler(
		components.UseCases.AdminLogs,
	)

	components.Handlers.Services = admin.NewAdminServicesHandler(
		components.UseCases.AdminLogs,
	)

	components.Handlers.AdminQuestion = admin.NewAdminQuestionHandler(
		components.UseCases.AdminQuestions,
		components.UseCases.AdminDatasets,
		components.LoggingService,
		deps.WebSocketService,
		logger,
	)

	components.Handlers.Friends = handler.NewFriendsHandler(
		components.UseCases.Friends,
	)

	components.Handlers.Games = handler.NewGamesHandler(
		components.UseCases.Game,
		logger,
	)

	components.Handlers.AdminDatasets = admin.NewAdminDatasetsHandler(
		components.UseCases.AdminDatasets,
		components.UseCases.AdminImports,
		components.UseCases.AdminQuestions,
		components.UseCases.AdminGeography,
		components.LoggingService,
		deps.WebSocketService,
		logger,
	)

	components.Handlers.AdminGameTemplates = admin.NewAdminGameTemplatesHandler(
		components.UseCases.AdminGameTemplates,
		components.LoggingService,
		logger,
	)

	components.Handlers.AdminImports = admin.NewAdminImportsHandler(
		components.UseCases.AdminImports,
		logger,
	)

	components.Handlers.Geography = handler.NewGeographyHandler(
		components.UseCases.Geography,
		logger,
	)

	components.Handlers.AdminMatchmaking = admin.NewAdminMatchmakingHandler(
		components.MatchmakingService,
		components.LoggingService,
		logger,
	)

	components.Handlers.Reports = handler.NewReportsHandler(
		components.UseCases.Report,
		logger,
	)

	components.Handlers.Matchmaking = handler.NewMatchmakingHandler(
		components.MatchmakingService,
		logger,
	)

	components.Handlers.Leaderboard = handler.NewLeaderboardHandler(
		usecase.NewLeaderboardUsecase(
			components.Repositories.User,
			components.Repositories.Game,
		),
		deps.RedisService,
	)

	components.Handlers.Lobby = handler.NewLobbyHandler(
		components.GameManager,
		components.MatchmakingService,
		deps.WebSocketService,
	)

	components.Handlers.Health = handler.NewHealthHandler()

	components.Handlers.AdminSettings = admin.NewAdminSettingsHandler(
		deps.RedisService,
		cfg,
		components.LoggingService,
		logger,
		nil,
	)

	components.Middlewares.Maintenance = middleware.NewMaintenanceMiddleware(
		deps.RedisService,
	)

	var podDiscoveryService *service.PodDiscoveryService
	if deps.RedisService != nil {
		podType := "main"
		if cfg.AppMode == config.AppModeHeadless {
			podType = "headless"
		}
		podDiscoveryService = service.NewPodDiscoveryService(
			podType,
			deps.RedisService,
			logger,
			deps.WebSocketService,
			components.GameManager,
		)
		deps.PodDiscoveryService = podDiscoveryService

		components.MatchmakingService.SetPodDiscovery(podDiscoveryService, podDiscoveryService.PodID())
	}

	components.Handlers.Pods = admin.NewPodsHandler(
		podDiscoveryService,
		deps.WebSocketService,
	)

	components.Middlewares.Auth = middleware.NewAuthMiddleware(
		deps.JWTService,
		deps.SessionService,
		cfg.JWTSecret,
	)

	components.Middlewares.APILogging = middleware.NewAPILoggingMiddleware(
		components.LoggingService,
		logger,
	)

	components.Middlewares.RateLimit = middleware.NewRateLimitMiddleware(
		deps.RedisService,
		cfg,
	)

	components.Handlers.WebSocket = handler.NewWebSocketHandler(
		deps.WebSocketService,
		components.LoggingService,
		logger,
	)

	deps.WebSocketService.SetGameActionHandler(components.UseCases.Game)
}

func (a *App) initialize() error {
	if err := a.validate(); err != nil {
		return fmt.Errorf("app validation failed: %w", err)
	}

	routes.RegisterRoutes(a.Router, a)

	mode := a.Config.AppMode
	isAdmin := mode == config.AppModeAdmin
	isHeadless := mode == config.AppModeHeadless

	if a.Deps.WebSocketService != nil {
		a.Deps.WebSocketService.StartRelay(a.appCtx)
		a.Logger.Info("WebSocket Pub/Sub relay started", zap.String("mode", mode))
	}

	if a.Deps.PodDiscoveryService != nil {
		go a.Deps.PodDiscoveryService.Start(a.appCtx)
		a.Logger.Info("Pod discovery service started",
			zap.String("pod_id", a.Deps.PodDiscoveryService.PodID()),
			zap.String("pod_type", a.Deps.PodDiscoveryService.PodType()),
		)
	}

	if a.MatchmakingService != nil {
		go a.MatchmakingService.Start(a.appCtx)
		a.Logger.Info("Matchmaking delegate poller started")
	}

	if !isAdmin {
		a.StartQuestionTimeoutChecker()
	}

	if !isHeadless {
		a.RestartSchedulers()
	}

	if a.GameEventBroadcaster != nil {
		a.GameEventBroadcaster.Start()
		a.Logger.Info("Game event broadcaster started")
	}

	if !isAdmin {
		if a.LobbyBroadcaster != nil {
			a.LobbyBroadcaster.Start()
			a.Logger.Info("Lobby broadcaster started")
		}
	}

	if err := a.setupDefaultAdmin(); err != nil {
		return fmt.Errorf("failed to setup default admin: %w", err)
	}

	if a.Config.SeedGameTemplates && a.AdminGameTemplatesUsecase != nil {
		if count, err := a.AdminGameTemplatesUsecase.SeedDefaultGameTemplates(); err != nil {
			a.Logger.Warn("Failed to seed default game templates", zap.Error(err))
		} else if count > 0 {
			a.Logger.Info("Seeded default game templates", zap.Int("created", count))
		}
	}

	return nil
}

func (a *App) validate() error {
	if a.Config == nil {
		return errors.New("config is required")
	}
	if a.DB == nil {
		return errors.New("database connection is required")
	}
	if a.Router == nil {
		return errors.New("router is required")
	}
	if a.Deps.JWTService == nil {
		return errors.New("JWT service is required")
	}
	if a.Deps.RedisService == nil {
		return errors.New("redis service is required")
	}
	if a.Deps.SessionService == nil {
		return errors.New("session service is required")
	}
	if a.Deps.WebSocketService == nil {
		return errors.New("websocket service is required")
	}
	if a.AuthMiddleware == nil {
		return errors.New("auth middleware is required")
	}
	if a.GameManager == nil {
		return errors.New("game manager is required")
	}
	if a.UserRepository == nil {
		return errors.New("user repository is required")
	}
	if a.AdminUserRepository == nil {
		return errors.New("admin user repository is required")
	}
	if a.GameRepository == nil {
		return errors.New("game repository is required")
	}
	if a.QuestionRepository == nil {
		return errors.New("question repository is required")
	}
	return nil
}

func (a *App) setupDefaultAdmin() error {
	exists, err := a.AdminUserUsecase.CheckAdminExists()
	if err != nil {
		return fmt.Errorf("error checking for admin users: %w", err)
	}

	if !exists {
		if err := a.AdminUserUsecase.CreateDefaultAdmin(); err != nil {
			return fmt.Errorf("error creating default admin: %w", err)
		}
	}
	return nil
}

func safeUserCacheAssertion(
	userCacheService service.UserCacheServiceInterface,
) (*service.UserCacheService, error) {
	if uc, ok := userCacheService.(*service.UserCacheService); ok {
		return uc, nil
	}
	return nil, fmt.Errorf("invalid user cache service type")
}

func setupRepositories(
	db *gorm.DB,
	userCacheService service.UserCacheServiceInterface,
	logger *zap.Logger,
) (
	repository.UserRepositoryInterface,
	repository.SessionRepositoryInterface,
	adminRepo.AdminUserRepositoryInterface,
	adminRepo.AdminLogsRepositoryInterface,
	repository.QuestionRepositoryInterface,
	repository.FriendsRepositoryInterface,
	repository.GameRepositoryInterface,
	error) {
	ucService, err := safeUserCacheAssertion(userCacheService)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	userRepository := repository.NewUserRepository(db, ucService, logger)
	sessionRepository := repository.NewSessionRepository(db, logger)
	adminUserRepository := adminRepo.NewAdminUserRepository(db, ucService, logger)
	logsRepository := adminRepo.NewAdminLogsRepository(db)
	questionRepository := repository.NewQuestionRepository(db)
	friendsRepository := repository.NewFriendsRepository(db, logger)
	gameRepository := repository.NewGameRepository(db, logger)
	return userRepository, sessionRepository, adminUserRepository, logsRepository, questionRepository, friendsRepository, gameRepository, nil
}

func setupRouter(cfg *config.Config, apiLoggingMiddleware *middleware.APILoggingMiddleware) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	if len(cfg.TrustedProxies) > 0 {
		_ = router.SetTrustedProxies(cfg.TrustedProxies)
	} else {
		_ = router.SetTrustedProxies(nil)
	}

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(apiLoggingMiddleware.LogAPIRequest())

	router.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if cfg.Env == "production" {
			if origin != "" {
				if cfg.IsOriginAllowed(origin) {
					c.Header("Access-Control-Allow-Origin", origin)
				} else {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Origin not allowed"})
					return
				}
			}
		} else {
			if origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
			} else {
				c.Header("Access-Control-Allow-Origin", "*")
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset, Retry-After")
		c.Header("Access-Control-Max-Age", "86400")

		if origin != "" {
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	return router
}
