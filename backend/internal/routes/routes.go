// backend/internal/routes/routes.go

package routes

import (
	"net/http"

	"github.com/Culturae-org/culturae/internal/config"
	"github.com/Culturae-org/culturae/internal/handler"
	"github.com/Culturae-org/culturae/internal/handler/admin"
	"github.com/Culturae-org/culturae/internal/middleware"
	"github.com/Culturae-org/culturae/internal/pkg/httputil"
	"github.com/Culturae-org/culturae/internal/version"

	"github.com/gin-gonic/gin"
)

type Dependencies interface {
	GetConfig() *config.Config
	GetAuthMiddleware() *middleware.AuthMiddleware
	GetRateLimitMiddleware() *middleware.RateLimitMiddleware
	GetAuthHandler() *handler.AuthHandler
	GetAvatarHandler() *handler.AvatarHandler
	GetUserHandler() *handler.UserHandler
	GetProfileHandler() *handler.ProfileHandler
	GetFriendsHandler() *handler.FriendsHandler
	GetWebSocketHandler() *handler.WebSocketHandler
	GetLogsHandler() *admin.AdminLogsHandler
	GetServicesHandler() *admin.AdminServicesHandler
	GetImportsHandler() *admin.AdminImportsHandler
	GetGeographyHandler() *handler.GeographyHandler
	GetGamesHandler() *handler.GamesHandler
	// ----------------------------------------------------
	// ADMIN CONTROLLERS
	// ----------------------------------------------------
	GetAdminFriendsHandler() *admin.AdminFriendsHandler
	GetAdminQuestionHandler() *admin.AdminQuestionHandler
	GetAdminGeographyHandler() *admin.AdminGeographyHandler
	GetAdminGamesHandler() *admin.AdminGamesHandler
	GetAdminAuthHandler() *admin.AdminAuthHandler
	GetAdminUserHandler() *admin.AdminUserHandler
	GetAdminAvatarHandler() *admin.AdminAvatarHandler
	GetAdminDatasetsHandler() *admin.AdminDatasetsHandler
	GetReportsHandler() *handler.ReportsHandler
	GetMatchmakingHandler() *handler.MatchmakingHandler
	GetLeaderboardHandler() *handler.LeaderboardHandler
	GetAdminReportsHandler() *admin.AdminReportsHandler
	GetAdminMatchmakingHandler() *admin.AdminMatchmakingHandler
	GetAdminSettingsHandler() *admin.AdminSettingsHandler
	GetMaintenanceMiddleware() *middleware.MaintenanceMiddleware
	GetAdminGameTemplatesHandler() *admin.AdminGameTemplatesHandler
	GetLobbyHandler() *handler.LobbyHandler
	GetHealthHandler() *handler.HealthHandler
}

func RegisterRoutes(r *gin.Engine, deps Dependencies) {
	mode := deps.GetConfig().AppMode

	if deps.GetHealthHandler() != nil {
		r.GET("/health", deps.GetHealthHandler().Health)
	} else {
		r.GET("/health", healthCheck())
	}

	if mode != config.AppModeHeadless {
		registerDashboardRoutes(r)
	}

	rateLimitMiddleware := deps.GetRateLimitMiddleware()
	maintenanceMiddleware := deps.GetMaintenanceMiddleware()

	registerPublicRoutes(r, deps, rateLimitMiddleware, maintenanceMiddleware)
	registerProtectedRoutes(r, deps, rateLimitMiddleware, maintenanceMiddleware, mode)

	if mode != config.AppModeHeadless {
		registerAdminRoutes(r, deps)
	}
}

func registerDashboardRoutes(r *gin.Engine) {
	dashboardHandler := admin.ServeDashboard()
	r.GET("/console", dashboardHandler)
	r.GET("/console/*path", dashboardHandler)
	r.GET("/login", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/console/login")
	})
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/console/")
	})
}

func registerPublicRoutes(r *gin.Engine, deps Dependencies, rateLimitMiddleware *middleware.RateLimitMiddleware, maintenanceMiddleware *middleware.MaintenanceMiddleware) {
	public := r.Group("/api/v1")
	public.Use(rateLimitMiddleware.RateLimit())
	public.Use(maintenanceMiddleware.CheckMaintenance())
	{
		auth := public.Group("/auth")
		auth.Use(rateLimitMiddleware.AuthRateLimit())
		{
			auth.POST("/register", deps.GetAuthHandler().Register)
			auth.POST("/login", deps.GetAuthHandler().Login)
			if deps.GetConfig().AppMode != config.AppModeHeadless {
				auth.POST("/login-admin", deps.GetAdminAuthHandler().LoginAdmin)
			}
			auth.POST("/refresh", deps.GetAuthHandler().RefreshToken)
		}

		public.GET("/lobby/stats", deps.GetLobbyHandler().GetLobbyStats)
	}
}

func registerProtectedRoutes(r *gin.Engine, deps Dependencies, rateLimitMiddleware *middleware.RateLimitMiddleware, maintenanceMiddleware *middleware.MaintenanceMiddleware, mode string) {
	protected := r.Group("/api/v1")
	protected.Use(rateLimitMiddleware.RateLimit())
	protected.Use(deps.GetAuthMiddleware().JWTAuthWithSessions())
	protected.Use(maintenanceMiddleware.CheckMaintenance())
	{
		auth := protected.Group("/auth")
		{
			auth.POST("/logout", deps.GetAuthHandler().Logout)
		}

		protected.GET("/me", deps.GetProfileHandler().GetCurrentUser)

		profile := protected.Group("/profile")
		{
			profile.GET("", deps.GetProfileHandler().UserProfile)
			profile.PUT("", deps.GetProfileHandler().UpdateProfile)
			profile.DELETE("", deps.GetProfileHandler().DeleteAccount)
			profile.POST("/regenerate-id", deps.GetProfileHandler().RegeneratePublicID)
			profile.POST("/change-password", deps.GetProfileHandler().ChangePassword)
			profile.GET("/stats", deps.GetProfileHandler().GetUserStats)
		}

		avatar := protected.Group("/avatar")
		{
			avatar.POST("", deps.GetAvatarHandler().UploadAvatar)
			avatar.DELETE("", deps.GetAvatarHandler().DeleteAvatar)
			avatar.GET("/:publicID", deps.GetAvatarHandler().GetAvatar)
		}

		users := protected.Group("/users")
		{
			users.GET("/search", deps.GetUserHandler().SearchPublicProfiles)
			users.GET("/profile/:publicID", deps.GetUserHandler().GetUserProfileWithRelationship)
		}

		friendships := protected.Group("/friends")
		{
			friendships.POST("/request/:toUserPublicID", deps.GetFriendsHandler().SendFriendRequest)
			friendships.DELETE("/request/:requestID", deps.GetFriendsHandler().CancelFriendRequest)

			friendships.POST("/accept/:requestID", deps.GetFriendsHandler().AcceptFriendRequest)
			friendships.POST("/reject/:requestID", deps.GetFriendsHandler().RejectFriendRequest)

			friendships.POST("/block/:requestID", deps.GetFriendsHandler().BlockFriendRequest)
			friendships.GET("/requests", deps.GetFriendsHandler().ListFriendRequests)
			friendships.GET("", deps.GetFriendsHandler().ListFriends)
			friendships.DELETE("/:friendUserPublicID", deps.GetFriendsHandler().RemoveFriend)

			friendships.GET("/blocked", deps.GetFriendsHandler().GetBlockedUsers)
			friendships.POST("/block-user/:userPublicID", deps.GetFriendsHandler().BlockUser)
			friendships.DELETE("/blocked/:userPublicID", deps.GetFriendsHandler().UnblockUser)
		}

		protected.GET("/game-templates", deps.GetGamesHandler().GetGameTemplates)

		games := protected.Group("/games")
		{
			games.POST("/create", deps.GetGamesHandler().CreateGame)

			games.POST("/:gameID/start", deps.GetGamesHandler().StartGame)

			games.POST("/:gameID/invite/:toUserPublicID", deps.GetGamesHandler().InviteToGame)
			games.POST("/invite/:inviteID/accept", deps.GetGamesHandler().AcceptGameInvite)
			games.POST("/invite/:inviteID/reject", deps.GetGamesHandler().RejectGameInvite)

			games.POST("/:gameID/join", deps.GetGamesHandler().JoinGame)
			games.POST("/:gameID/leave", deps.GetGamesHandler().LeaveGame)
			games.POST("/:gameID/cancel", deps.GetGamesHandler().CancelGame)

			games.POST("/:gameID/answer", deps.GetGamesHandler().SubmitAnswer)

			games.GET("/:gameID/status", deps.GetGamesHandler().GetGameStatus)
			games.GET("/active", deps.GetGamesHandler().GetActiveGames)
			games.GET("/invites", deps.GetGamesHandler().GetUserGameInvites)
			games.GET("/history", deps.GetGamesHandler().GetGameHistory)

			games.POST("/:gameID/questions/:number/report", deps.GetReportsHandler().CreateReportFromGame)

			matchmaking := games.Group("/matchmaking")
			{
				matchmaking.POST("/join", deps.GetMatchmakingHandler().JoinQueue)
				matchmaking.POST("/leave", deps.GetMatchmakingHandler().LeaveQueue)
			}
		}

		geography := protected.Group("/geography")
		{
			geography.GET("/flags/:country_code", deps.GetGeographyHandler().GetFlag)
			geography.GET("/flags/:country_code/url", deps.GetGeographyHandler().GetFlagURL)
			geography.GET("/flags/:country_code/png/:format", deps.GetGeographyHandler().GetFlagPNG)

			geography.GET("/countries", deps.GetGeographyHandler().GetCountries)
			geography.GET("/continents", deps.GetGeographyHandler().GetContinents)
		}

		protected.GET("/leaderboard", deps.GetLeaderboardHandler().GetLeaderboard)

		if mode != config.AppModeAdmin {
			protected.GET("/realtime", deps.GetWebSocketHandler().HandleWebSocket)
		}
	}
}

func registerAdminRoutes(r *gin.Engine, deps Dependencies) {
	protected := r.Group("/api/v1")
	protected.Use(deps.GetRateLimitMiddleware().RateLimit())
	protected.Use(deps.GetAuthMiddleware().JWTAuthWithSessions())
	protected.Use(deps.GetMaintenanceMiddleware().CheckMaintenance())

	adminGrp := protected.Group("/admin")
	adminGrp.Use(deps.GetAuthMiddleware().AdminRequired())
	{
		adminGrp.GET("/openapi.yaml", serveAdminOpenAPI())
		// admin.GET("/asyncapi.yaml", serveAdminAsyncAPI()) - waiting for asyncAPI scalar support

		adminGrp.GET("/info", platformInfo(deps.GetConfig().Env))

		adminGrp.GET("/realtime", deps.GetWebSocketHandler().HandleAdminWebSocket)
		adminGrp.GET("/me", deps.GetAdminUserHandler().GetCurrentUser)

		users := adminGrp.Group("/users")
		{
			users.GET("", deps.GetAdminUserHandler().GetAllUsers)
			users.GET("/count", deps.GetAdminUserHandler().GetUserCount)
			users.GET("/count/online", deps.GetAdminUserHandler().GetUserOnlineCount)
			users.GET("/count/active/weekly", deps.GetAdminUserHandler().GetWeeklyActiveUserCount)
			users.GET("/search", deps.GetAdminUserHandler().SearchUsers)
			users.GET("/:id", deps.GetAdminUserHandler().GetUserByID)

			users.PUT("/:id", deps.GetAdminUserHandler().UpdateUser)
			users.PUT("/:id/password", deps.GetAdminUserHandler().UpdateUserPassword)
			users.POST("", deps.GetAdminUserHandler().CreateUser)

			users.PATCH("/:id/deactivate", deps.GetAdminUserHandler().DeactivateUser)
			users.PATCH("/:id/status", deps.GetAdminUserHandler().UpdateUserStatus)
			users.DELETE("/:id", deps.GetAdminUserHandler().DeleteUser)
			users.POST("/:id/regenerate-id", deps.GetAdminUserHandler().RegeneratePublicID)
			users.POST("/:id/ban", deps.GetAdminUserHandler().BanUser)
			users.POST("/:id/unban", deps.GetAdminUserHandler().UnbanUser)

			users.GET("/level-stats", deps.GetAdminUserHandler().GetUserLevelStats)
			users.GET("/role-stats", deps.GetAdminUserHandler().GetUserRoleStats)
			users.GET("/creation-dates", deps.GetAdminUserHandler().GetUserCreationDates)
		}

		avatar := adminGrp.Group("/avatar")
		{
			avatar.POST("/:userID", deps.GetAdminAvatarHandler().UploadUserAvatar)
			avatar.DELETE("/:userID", deps.GetAdminAvatarHandler().DeleteUserAvatar)
			avatar.GET("/:userID", deps.GetAdminAvatarHandler().GetUserAvatar)
		}

		logs := adminGrp.Group("/logs")
		{
			logs.GET("/connections/:id", deps.GetAdminUserHandler().GetUserConnectionLogs)
			logs.GET("/user-actions/:id", deps.GetLogsHandler().GetUserActionLogsByID)
			logs.GET("/active-sessions/:id", deps.GetAdminUserHandler().GetUserActiveSessions)

			logs.GET("/user-actions", deps.GetLogsHandler().GetAllUserActionLogs)
			logs.GET("/admin-actions", deps.GetLogsHandler().GetAdminActionLogs)
			logs.GET("/connections", deps.GetLogsHandler().GetConnectionLogs)

			logs.GET("/system-metrics", deps.GetServicesHandler().GetSystemMetrics)
			logs.GET("/service-status", deps.GetServicesHandler().GetServiceStatus)

			logs.GET("/api-requests/stats", deps.GetLogsHandler().GetAPIRequestStats)
			logs.GET("/api-requests/timestamps", deps.GetLogsHandler().GetAPIRequestTimestamps)
			logs.GET("/admin-actions/stats", deps.GetLogsHandler().GetAdminActionStats)
			logs.GET("/user-actions/stats", deps.GetLogsHandler().GetUserActionStats)
		}

		questions := adminGrp.Group("/questions")
		{
			questions.POST("", deps.GetAdminQuestionHandler().CreateQuestion)
			questions.GET("", deps.GetAdminQuestionHandler().ListQuestions)
			questions.GET("/:id", deps.GetAdminQuestionHandler().GetQuestion)
			questions.GET("/slug/:slug", deps.GetAdminQuestionHandler().GetQuestionBySlug)
			questions.PUT("/:id", deps.GetAdminQuestionHandler().UpdateQuestion)
			questions.DELETE("/:id", deps.GetAdminQuestionHandler().DeleteQuestion)

			questions.GET("/export", deps.GetAdminQuestionHandler().ExportQuestionsClean)
		}

		friendships := adminGrp.Group("/friends")
		{
			friendships.GET("/requests/:userID", deps.GetAdminFriendsHandler().ListFriendRequestsForUser)
			friendships.GET("/:userID", deps.GetAdminFriendsHandler().ListFriendsForUser)
		}

		imports := adminGrp.Group("/imports")
		{
			imports.GET("", deps.GetImportsHandler().ListImportJobs)
			imports.GET("/stats", deps.GetImportsHandler().GetImportStats)
			imports.GET("/:id", deps.GetImportsHandler().GetImportJob)
			imports.GET("/:id/logs", deps.GetImportsHandler().GetImportJobLogs)
		}

		datasets := adminGrp.Group("/datasets")
		{
			datasets.GET("", deps.GetAdminDatasetsHandler().ListDatasets)
			datasets.GET("/default", deps.GetAdminDatasetsHandler().GetDefaultDataset)
			datasets.GET("/history", deps.GetAdminDatasetsHandler().GetHistory)
			datasets.GET("/:id", deps.GetAdminDatasetsHandler().GetDataset)
			datasets.GET("/slug/:slug", deps.GetAdminDatasetsHandler().GetDatasetBySlug)
			datasets.GET("/:id/questions", deps.GetAdminDatasetsHandler().GetDatasetQuestions)
			datasets.GET("/:id/stats", deps.GetAdminDatasetsHandler().GetDatasetStatistics)
			datasets.POST("", deps.GetAdminDatasetsHandler().CreateDataset)
			datasets.PATCH("/:id", deps.GetAdminDatasetsHandler().UpdateDataset)
			datasets.DELETE("/:id", deps.GetAdminDatasetsHandler().DeleteDataset)
			datasets.POST("/:id/set-default", deps.GetAdminDatasetsHandler().SetDefaultDataset)
			datasets.POST("/check-updates", deps.GetAdminDatasetsHandler().CheckForUpdates)
			datasets.POST("/import", deps.GetAdminDatasetsHandler().ImportDataset)
			datasets.POST("/:id/update-stats", deps.GetAdminDatasetsHandler().UpdateDatasetStatistics)
		}

		gameTemplates := adminGrp.Group("/game-templates")
		{
			gameTemplates.GET("", deps.GetAdminGameTemplatesHandler().List)
			gameTemplates.GET("/:id", deps.GetAdminGameTemplatesHandler().Get)
			gameTemplates.POST("", deps.GetAdminGameTemplatesHandler().Create)
			gameTemplates.POST("/seed-defaults", deps.GetAdminGameTemplatesHandler().SeedDefaults)
			gameTemplates.PATCH("/:id", deps.GetAdminGameTemplatesHandler().Update)
			gameTemplates.DELETE("/:id", deps.GetAdminGameTemplatesHandler().Delete)
		}

		geography := adminGrp.Group("/geography")
		{
			geography.GET("/slug/:slug", deps.GetAdminGeographyHandler().GetGeographyDatasetBySlug)
			geography.GET("/flags/:country_code", deps.GetAdminGeographyHandler().GetFlag)

			geography.GET("/:id", deps.GetAdminGeographyHandler().GetGeographyDataset)
			geography.GET("/:id/stats", deps.GetAdminGeographyHandler().GetGeographyDatasetStatistics)
			geography.DELETE("/:id", deps.GetAdminGeographyHandler().DeleteGeographyDataset)
			geography.POST("/:id/set-default", deps.GetAdminGeographyHandler().SetDefaultGeographyDataset)

			geography.GET("/:id/countries", deps.GetAdminGeographyHandler().ListCountries)
			geography.GET("/:id/countries/search", deps.GetAdminGeographyHandler().SearchCountries)
			geography.GET("/:id/countries/:slug", deps.GetAdminGeographyHandler().GetCountry)
			geography.PATCH("/:id/countries/:slug", deps.GetAdminGeographyHandler().UpdateCountry)
			geography.GET("/:id/countries/continent/:continent", deps.GetAdminGeographyHandler().ListCountriesByContinent)
			geography.GET("/:id/countries/region/:region", deps.GetAdminGeographyHandler().ListCountriesByRegion)

			geography.GET("/:id/continents", deps.GetAdminGeographyHandler().ListContinents)
			geography.GET("/:id/continents/:slug", deps.GetAdminGeographyHandler().GetContinent)
			geography.PATCH("/:id/continents/:slug", deps.GetAdminGeographyHandler().UpdateContinent)

			geography.GET("/:id/regions", deps.GetAdminGeographyHandler().ListRegions)
			geography.GET("/:id/regions/:slug", deps.GetAdminGeographyHandler().GetRegion)
			geography.PATCH("/:id/regions/:slug", deps.GetAdminGeographyHandler().UpdateRegion)
			geography.GET("/:id/regions/continent/:continent", deps.GetAdminGeographyHandler().ListRegionsByContinent)

			geography.GET("/:id/flags/:country_code", deps.GetAdminGeographyHandler().GetFlag)
			geography.GET("/:id/flags/:country_code/url", deps.GetAdminGeographyHandler().GetFlagURL)
		}

		games := adminGrp.Group("/games")
		{
			games.GET("", deps.GetAdminGamesHandler().ListGames)
			games.GET("/stats", deps.GetAdminGamesHandler().GetGameStats)

			games.GET("/:id", deps.GetAdminGamesHandler().GetGameByID)
			games.GET("/:id/players", deps.GetAdminGamesHandler().GetGamePlayers)
			games.GET("/:id/questions", deps.GetAdminGamesHandler().GetGameQuestions)
			games.GET("/:id/answers", deps.GetAdminGamesHandler().GetGameAnswers)
			games.GET("/:id/events", deps.GetAdminGamesHandler().GetGameEventLogs)
			games.POST("/:id/cancel", deps.GetAdminGamesHandler().AdminCancelGame)
			games.POST("/:id/archive", deps.GetAdminGamesHandler().ArchiveGame)
			games.POST("/:id/unarchive", deps.GetAdminGamesHandler().UnarchiveGame)
			games.DELETE("/:id", deps.GetAdminGamesHandler().DeleteGameByID)

			games.GET("/invites", deps.GetAdminGamesHandler().ListGameInvites)
			games.GET("/invites/pending", deps.GetAdminGamesHandler().ListPendingInvites)
			games.DELETE("/invites/:inviteID", deps.GetAdminGamesHandler().DeleteGameInvite)
			games.POST("/invites/:inviteID/cancel", deps.GetAdminGamesHandler().CancelGameInvite)

			games.GET("/stats/modes", deps.GetAdminGamesHandler().GetGameModeStats)
			games.GET("/stats/daily", deps.GetAdminGamesHandler().GetDailyGameStats)
			games.GET("/stats/users/:userID", deps.GetAdminGamesHandler().GetUserGameStats)
			games.GET("/users/:userID", deps.GetAdminGamesHandler().GetUserGameHistory)
			games.GET("/stats/performance", deps.GetAdminGamesHandler().GetGamePerformanceStats)
			games.POST("/cleanup", deps.GetAdminGamesHandler().CleanupAbandonedGames)
			games.POST("/maintenance", deps.GetAdminGamesHandler().RunGameMaintenance)
		}

		reports := adminGrp.Group("/reports")
		{
			reports.GET("", deps.GetAdminReportsHandler().ListReports)
			reports.GET("/:id", deps.GetAdminReportsHandler().GetReport)
			reports.PATCH("/:id/status", deps.GetAdminReportsHandler().UpdateReportStatus)
		}

		matchmaking := adminGrp.Group("/matchmaking")
		{
			matchmaking.GET("/stats", deps.GetAdminMatchmakingHandler().GetQueueStats)
			matchmaking.DELETE("/queue/:mode", deps.GetAdminMatchmakingHandler().ClearQueue)
		}

		settings := adminGrp.Group("/settings")
		{
			settings.GET("/maintenance", deps.GetAdminSettingsHandler().GetMaintenanceStatus)
			settings.POST("/maintenance", deps.GetAdminSettingsHandler().SetMaintenanceMode)
			settings.GET("/rate-limit", deps.GetAdminSettingsHandler().GetRateLimitConfig)
			settings.PUT("/rate-limit", deps.GetAdminSettingsHandler().UpdateRateLimitConfig)
			settings.POST("/cache/clear", deps.GetAdminSettingsHandler().ClearCache)
			settings.GET("/websocket", deps.GetAdminSettingsHandler().GetWebSocketConfig)
			settings.PUT("/websocket", deps.GetAdminSettingsHandler().UpdateWebSocketConfig)
			settings.GET("/avatar", deps.GetAdminSettingsHandler().GetAvatarConfig)
			settings.PUT("/avatar", deps.GetAdminSettingsHandler().UpdateAvatarConfig)
			settings.GET("/xp", deps.GetAdminSettingsHandler().GetXPConfig)
			settings.PUT("/xp", deps.GetAdminSettingsHandler().UpdateXPConfig)
			settings.GET("/elo", deps.GetAdminSettingsHandler().GetELOConfig)
			settings.PUT("/elo", deps.GetAdminSettingsHandler().UpdateELOConfig)
			settings.GET("/games", deps.GetAdminSettingsHandler().GetGamesConfig)
			settings.PUT("/games", deps.GetAdminSettingsHandler().UpdateGamesConfig)
			settings.GET("/system", deps.GetAdminSettingsHandler().GetSystemConfig)
			settings.PUT("/system", deps.GetAdminSettingsHandler().UpdateSystemConfig)
			settings.GET("/games-countdown", deps.GetAdminSettingsHandler().GetGameCountdownConfig)
			settings.PUT("/games-countdown", deps.GetAdminSettingsHandler().UpdateGameCountdownConfig)
			settings.GET("/auth", deps.GetAdminSettingsHandler().GetAuthConfig)
			settings.PUT("/auth", deps.GetAdminSettingsHandler().UpdateAuthConfig)
			settings.GET("/version-status", deps.GetAdminSettingsHandler().GetVersionStatus)
		}
	}
}

func healthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	}
}

func platformInfo(env string) gin.HandlerFunc {
	return func(c *gin.Context) {
		httputil.Success(c, http.StatusOK, gin.H{
			"service":     "culturae-server",
			"version":     version.Version,
			"build_time":  version.BuildTime,
			"environment": env,
		})
	}
}

func serveAdminOpenAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.File("docs/admin-openapi.yaml")
	}
}

// func serveAdminAsyncAPI() gin.HandlerFunc {
//	 return func(c *gin.Context) {
//		c.File("docs/asyncapi.yaml")
// 	 }
//  }
