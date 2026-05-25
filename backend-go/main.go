package main

import (
	"fmt"
	"log/slog"
	"os"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/database"
	"gpt-image-playground/backend/handler"
	glog "gpt-image-playground/backend/log"
	"gpt-image-playground/backend/middleware"
	"gpt-image-playground/backend/util"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logger (text format for dev, JSON for production)
	glog.Init(false)

	util.EnsureRuntimeDirs(config.App.DataDir, config.App.UploadDir)

	if err := database.Init(); err != nil {
		slog.Error("初始化数据库失败", "error", err)
		panic(err)
	}
	if sqlDB, err := database.DB.DB(); err == nil {
		defer sqlDB.Close()
	}

	r := gin.Default()

	r.Use(middleware.RequestLogger())

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
	}))

	r.MaxMultipartMemory = 50 << 20

	r.GET("/api/health", handler.Health)
	r.GET("/api/announcement", handler.AnnouncementPublic)
	r.GET("/api/changelog/latest", handler.ChangelogLatestPublic)
	r.GET("/api/changelog", handler.ChangelogListPublic)

	auth := r.Group("/api/auth")
	auth.POST("/login", handler.AuthLogin)
	auth.GET("/me", middleware.AuthMiddleware(), handler.AuthMe)
	auth.POST("/redeem", middleware.AuthMiddleware(), handler.AuthRedeem)
	auth.POST("/login-password", handler.AuthLoginPassword)
	auth.POST("/register", handler.AuthRegister)
	auth.POST("/migrate", middleware.AuthMiddleware(), handler.AuthMigrate)
	auth.POST("/change-password", middleware.AuthMiddleware(), handler.AuthChangePassword)
	auth.PUT("/username", middleware.AuthMiddleware(), handler.AuthChangeUsername)
	auth.PUT("/invite-code", middleware.AuthMiddleware(), handler.AuthSetInviteCode)
	auth.GET("/invite-code", middleware.AuthMiddleware(), handler.AuthGetInviteCode)
	auth.GET("/invited-users", middleware.AuthMiddleware(), handler.AuthGetInvitedUsers)

	cfg := r.Group("/api/config")
	cfg.GET("/public", handler.ConfigPublic)

	images := r.Group("/api/images", middleware.AuthMiddleware())
	images.POST("", handler.ImagesUpload)
	images.GET("/:id", handler.ImagesGet)
	images.DELETE("/:id", handler.ImagesDelete)

	tasks := r.Group("/api/tasks", middleware.AuthMiddleware())
	tasks.GET("", handler.TasksList)
	tasks.GET("/:id/stream", handler.TaskStream)
	tasks.PUT("/:id", handler.TasksUpdate)
	tasks.DELETE("/:id", handler.TasksDelete)
	tasks.DELETE("/", handler.TasksClear)

	generate := r.Group("/api", middleware.AuthMiddleware())
	generate.POST("/generate", handler.GenerateImage)
	generate.POST("/edit", handler.GenerateImage)
	generate.POST("/feedback", handler.FeedbackCreate)

	// Admin API
	admin := r.Group("/api/admin")
	admin.POST("/login", handler.AdminLogin)
	adminAuth := r.Group("/api/admin", middleware.AdminMiddleware())
	adminAuth.GET("/users", handler.AdminListUsers)
	adminAuth.PUT("/users/:id/quota", handler.AdminUpdateQuota)
	adminAuth.PUT("/users/:id/status", handler.AdminToggleStatus)
		adminAuth.PUT("/users/:id/unlimited", handler.AdminToggleUnlimited)
	adminAuth.DELETE("/users/:id", handler.AdminDeleteUser)
	adminAuth.DELETE("/users", handler.AdminDeleteUsers)
	adminAuth.POST("/codes", handler.AdminCreateCode)
	adminAuth.GET("/codes", handler.AdminListCodes)
	adminAuth.DELETE("/codes", handler.AdminDeleteCodes)
	adminAuth.GET("/config/endpoints", handler.AdminGetEndpoints)
	adminAuth.PUT("/config/endpoints", handler.AdminUpdateEndpoints)
	adminAuth.GET("/config/pricing", handler.AdminGetPricingConfig)
	adminAuth.PUT("/config/pricing", handler.AdminUpdatePricingConfig)
	adminAuth.GET("/announcement", handler.AdminGetAnnouncement)
	adminAuth.PUT("/announcement", handler.AdminUpdateAnnouncement)
	adminAuth.GET("/feedback", handler.AdminListFeedbacks)
	adminAuth.PUT("/feedback/:id/status", handler.AdminUpdateFeedbackStatus)
	adminAuth.GET("/changelog", handler.AdminListChangelogs)
	adminAuth.POST("/changelog", handler.AdminCreateChangelog)
	adminAuth.PUT("/changelog/:id", handler.AdminUpdateChangelog)
	adminAuth.DELETE("/changelog/:id", handler.AdminDeleteChangelog)
	adminAuth.GET("/analytics/summary", handler.AdminBillingSummary)
	adminAuth.GET("/analytics/trend", handler.AdminBillingTrend)
	adminAuth.GET("/analytics/endpoints", handler.AdminBillingEndpointBreakdown)
	adminAuth.GET("/analytics/users", handler.AdminBillingUserBreakdown)
	adminAuth.PUT("/users/:id/password", handler.AdminResetPassword)
	adminAuth.GET("/invite-config", handler.AdminGetInviteConfig)
	adminAuth.PUT("/invite-config", handler.AdminUpdateInviteConfig)
	adminAuth.GET("/invites", handler.AdminListInvites)

	addr := fmt.Sprintf(":%d", config.App.Port)
	slog.Info("后端服务启动", "addr", addr)
	if err := r.Run(addr); err != nil {
		slog.Error("启动服务失败", "error", err)
		panic(err)
	}
}
