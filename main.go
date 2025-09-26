package main

import (
	"log"
	"os"

	"prometheus-webhook/handlers"
	"prometheus-webhook/internal/provider/dingding"
	"prometheus-webhook/internal/provider/feishu"
	"prometheus-webhook/internal/provider/weixin"
	"prometheus-webhook/models"
	"prometheus-webhook/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	configPath := "config/config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	configService := services.NewConfigService()
	if err := configService.LoadConfig(configPath); err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	config := configService.GetConfig()

	// 设置时区
	location, err := services.SetTimezone(config.Template.Timezone)
	if err != nil {
		log.Fatalf("加载时区失败: %v", err)
	}

	// 加载模板服务
	templateService := services.NewTemplateService(location)

	// 设置Gin模式
	if config.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()
	router.Use(gin.Recovery())

	// 初始化处理器
	healthHandler := handlers.NewHealthHandler(config)
	router.GET("/health", healthHandler.HealthCheck)

	// 为每个启用的 webhook 创建路由
	setupWebhookRoutes(router, &config, templateService)

	// 启动服务器
	server := handlers.NewServer(config.Server.Port, config.Server.Timeout, router)

	log.Printf("Prometheus Webhook服务启动在端口 %s", config.Server.Port)
	log.Printf("可用的webhook端点:")
	if config.Webhooks.Feishu.Enable {
		log.Printf("  POST http://127.0.0.1:%s/feishu", config.Server.Port)
	}
	if config.Webhooks.Dingding.Enable {
		log.Printf("  POST http://127.0.0.1:%s/dingding", config.Server.Port)
	}
	if config.Webhooks.Weixin.Enable {
		log.Printf("  POST http://127.0.0.1:%s/weixin", config.Server.Port)
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

func setupWebhookRoutes(router *gin.Engine, config *models.Config, templateService *services.TemplateService) {
	if config.Webhooks.Feishu.Enable {
		feishuService := feishu.NewService()
		webhookHandler := handlers.NewWebhookHandler(feishuService, config.Webhooks.Feishu, templateService)
		router.POST("/feishu", webhookHandler.Handle)
		log.Printf("注册路由: POST /feishu -> %s", config.Webhooks.Feishu.WebhookURL)
	}

	if config.Webhooks.Dingding.Enable {
		dingdingService := dingding.NewService()
		webhookHandler := handlers.NewWebhookHandler(dingdingService, config.Webhooks.Dingding, templateService)
		router.POST("/dingding", webhookHandler.Handle)
		log.Printf("注册路由: POST /dingding -> %s", config.Webhooks.Dingding.WebhookURL)
	}

	if config.Webhooks.Weixin.Enable {
		weixinService := weixin.NewService()
		webhookHandler := handlers.NewWebhookHandler(weixinService, config.Webhooks.Weixin, templateService)
		router.POST("/weixin", webhookHandler.Handle)
		log.Printf("注册路由: POST /weixin -> %s", config.Webhooks.Weixin.WebhookURL)
	}
}
