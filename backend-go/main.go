package main

import (
	"fmt"
	"log"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/database"
	"gpt-image-playground/backend/handler"
	"gpt-image-playground/backend/middleware"
	"gpt-image-playground/backend/util"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	util.EnsureRuntimeDirs(config.App.DataDir, config.App.UploadDir)

	if err := database.Init(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.DB.Close()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
	}))

	r.MaxMultipartMemory = 50 << 20

	r.GET("/api/health", handler.Health)

	auth := r.Group("/api/auth")
	auth.POST("/login", handler.AuthLogin)
	auth.GET("/me", middleware.AuthMiddleware(), handler.AuthMe)

	cfg := r.Group("/api/config")
	cfg.GET("/public", middleware.AuthMiddleware(), handler.ConfigPublic)

	images := r.Group("/api/images", middleware.AuthMiddleware())
	images.POST("", handler.ImagesUpload)
	images.GET("/:id", handler.ImagesGet)
	images.DELETE("/:id", handler.ImagesDelete)

	tasks := r.Group("/api/tasks", middleware.AuthMiddleware())
	tasks.GET("", handler.TasksList)
	tasks.PUT("/:id", handler.TasksUpdate)
	tasks.DELETE("/:id", handler.TasksDelete)
	tasks.DELETE("/", handler.TasksClear)

	generate := r.Group("/api", middleware.AuthMiddleware())
	generate.POST("/generate", handler.GenerateImage)
	generate.POST("/edit", handler.GenerateImage)
	generate.POST("/responses-generate", handler.GenerateResponses)

	addr := fmt.Sprintf(":%d", config.App.Port)
	log.Printf("Backend server listening on http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}
