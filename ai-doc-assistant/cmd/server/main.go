package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"ai-doc-assistant/internal/config"
	"ai-doc-assistant/internal/handler"
	"ai-doc-assistant/internal/repository"
	"ai-doc-assistant/internal/service"
	"ai-doc-assistant/pkg/logger"
	"ai-doc-assistant/pkg/middleware"
)

// @title AI文档助手API
// @version 1.0
// @description 基于Eino框架的智能文档问答系统
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}

	// 2. 初始化日志
	logger.Init(cfg.Log.Level, cfg.Log.File)
	defer logger.Sync()

	zap.S().Info("🚀 启动AI文档助手服务器...")

	// 3. 初始化数据库
	db, err := repository.NewDatabase(cfg.Database)
	if err != nil {
		zap.S().Fatalf("❌ 数据库初始化失败: %v", err)
	}
	defer db.Close()

	// 4. 初始化Eino服务
	einoService, err := service.NewEinoService(cfg)
	if err != nil {
		zap.S().Fatalf("❌ Eino服务初始化失败: %v", err)
	}
	defer einoService.Close()
	
	// 设置数据库连接
	einoService.SetDatabase(db)

	// 5. 初始化Handler层
	handlers := handler.NewHandlers(einoService)

	// 6. 设置路由
	router := setupRoutes(cfg, handlers)

	// 7. 启动HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// 8. 优雅启动
	go func() {
		zap.S().Infof("🌟 服务器启动成功，监听端口: %d", cfg.Server.Port)
		zap.S().Infof("📱 Web界面: http://localhost:%d", cfg.Server.Port)
		zap.S().Infof("📚 API文档: http://localhost:%d/swagger/index.html", cfg.Server.Port)
		zap.S().Infof("❤️  健康检查: http://localhost:%d/health", cfg.Server.Port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.S().Fatalf("❌ HTTP服务器启动失败: %v", err)
		}
	}()

	// 9. 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zap.S().Info("🛑 正在关闭服务器...")

	// 10. 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		zap.S().Errorf("❌ 服务器强制关闭: %v", err)
	}

	zap.S().Info("✅ 服务器已安全关闭")
}

func setupRoutes(cfg *config.Config, handlers *handler.Handlers) *gin.Engine {
	// 设置Gin模式
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	// 中间件
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())

	// 静态文件服务
	router.Static("/static", "./web/static")
	router.Static("/uploads", "./uploads")
	router.LoadHTMLGlob("web/templates/*")

	// 首页
	router.GET("/", handlers.Home)

	// 健康检查
	router.GET("/health", handlers.HealthCheck)

	// API路由组
	v1 := router.Group("/api/v1")
	{
		// 文档管理
		docs := v1.Group("/documents")
		{
			docs.POST("", handlers.UploadDocument)           // 上传文档
			docs.GET("", handlers.ListDocuments)             // 文档列表
			docs.GET("/:id", handlers.GetDocument)           // 获取文档详情
			docs.PUT("/:id", handlers.UpdateDocument)        // 更新文档
			docs.DELETE("/:id", handlers.DeleteDocument)     // 删除文档
			docs.POST("/batch", handlers.BatchUpload)        // 批量上传
		}

		// 问答功能
		qa := v1.Group("/qa")
		{
			qa.POST("/ask", handlers.AskQuestion)            // 提问
			qa.GET("/history", handlers.QueryHistory)       // 查询历史
			qa.POST("/feedback", handlers.SubmitFeedback)   // 反馈
		}

		// 用户管理
		users := v1.Group("/users")
		{
			users.POST("", handlers.CreateUser)             // 创建用户
			users.GET("", handlers.ListUsers)               // 用户列表
			users.GET("/:id", handlers.GetUser)             // 获取用户
			users.PUT("/:id", handlers.UpdateUser)          // 更新用户
		}

		// 统计信息
		stats := v1.Group("/stats")
		{
			stats.GET("/overview", handlers.GetOverview)    // 总览统计
			stats.GET("/usage", handlers.GetUsageStats)     // 使用统计
			stats.GET("/performance", handlers.GetPerformanceStats) // 性能统计
		}
	}

	// Swagger文档(仅开发和测试环境)
	if cfg.Server.Mode != "production" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 监控端点
	router.GET("/metrics", handlers.Metrics)

	return router
}