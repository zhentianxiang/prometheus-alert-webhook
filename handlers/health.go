package handlers

import (
	"net/http"
	"prometheus-webhook/models"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	config models.Config
}

func NewHealthHandler(config models.Config) *HealthHandler {
	return &HealthHandler{config: config}
}

func (hh *HealthHandler) HealthCheck(c *gin.Context) {
	config := hh.config

	enabledWebhooks := make(map[string]string)
	if config.Webhooks.Feishu.Enable {
		enabledWebhooks["feishu"] = config.Webhooks.Feishu.WebhookURL
	}
	if config.Webhooks.Dingding.Enable {
		enabledWebhooks["dingding"] = config.Webhooks.Dingding.WebhookURL
	}
	if config.Webhooks.Weixin.Enable {
		enabledWebhooks["weixin"] = config.Webhooks.Weixin.WebhookURL
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "healthy",
		"timestamp":        time.Now().Format(time.RFC3339),
		"version":          "1.0.0",
		"enabled_webhooks": enabledWebhooks,
	})
}

func NewServer(port string, timeout time.Duration, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}
}
