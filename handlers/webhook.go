package handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"prometheus-webhook/models"
	"prometheus-webhook/services"

	"github.com/gin-gonic/gin"
)

// MessageHandler 定义了发送消息服务的通用接口
type MessageHandler interface {
	SendMessage(providerConfig models.WebhookProvider, message string) error
}

type WebhookHandler struct {
	messageHandler  MessageHandler
	providerConfig  models.WebhookProvider
	templateService *services.TemplateService
}

func NewWebhookHandler(handler MessageHandler, providerConfig models.WebhookProvider, templateService *services.TemplateService) *WebhookHandler {
	return &WebhookHandler{
		messageHandler:  handler,
		providerConfig:  providerConfig,
		templateService: templateService,
	}
}

func (wh *WebhookHandler) Handle(c *gin.Context) {
	var webhookData models.AlertmanagerWebhook

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("读取请求体失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取请求"})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	log.Printf("接收到原始告警 Webhook: %s", string(bodyBytes))

	if err := c.BindJSON(&webhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的JSON数据"})
		return
	}

	status := webhookData.Status
	if status == "" && len(webhookData.Alerts) > 0 {
		status = webhookData.Alerts[0].Status
	}
	log.Printf("接收到告警: %d 条告警, 状态: %s", len(webhookData.Alerts), status)

	// 渲染模板
	var messageBuf bytes.Buffer
	data := wh.prepareTemplateData(webhookData)

	tmpl, err := wh.templateService.GetTemplate(wh.providerConfig.Template)
	if err != nil {
		log.Printf("获取模板 '%s' 失败: %v", wh.providerConfig.Template, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "模板加载失败"})
		return
	}

	templateBaseName := filepath.Base(wh.providerConfig.Template)
	templateName := strings.TrimSuffix(templateBaseName, filepath.Ext(templateBaseName)) + "_message"

	if err := tmpl.ExecuteTemplate(&messageBuf, templateName, data); err != nil {
		log.Printf("模板渲染失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "模板渲染失败"})
		return
	}

	// 发送消息
	if err := wh.messageHandler.SendMessage(wh.providerConfig, messageBuf.String()); err != nil {
		log.Printf("发送消息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送消息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "告警处理成功",
		"sent_to": wh.providerConfig.WebhookURL,
		"alerts":  len(webhookData.Alerts),
	})
}

func (wh *WebhookHandler) prepareTemplateData(webhookData models.AlertmanagerWebhook) map[string]interface{} {
	data := map[string]interface{}{
		"alerts": webhookData.Alerts,
	}

	var feishuAlerts []map[string]interface{}
	for _, alert := range webhookData.Alerts {
		feishuAlert := map[string]interface{}{
			"Status":      alert.Status,
			"Labels":      alert.Labels,
			"Annotations": alert.Annotations,
			"StartsAt":    alert.StartsAt,
			"EndsAt":      alert.EndsAt,
			"Fields":      getAlertFields(alert),
		}
		feishuAlerts = append(feishuAlerts, feishuAlert)
	}
	data["alerts"] = feishuAlerts
	return data
}

// getAlertFields 提取告警中的标签用于飞书卡片展示
func getAlertFields(alert models.Alert) []map[string]string {
	var fields []map[string]string
	fieldMapping := map[string]string{
		"namespace":  "🏷️ **命名空间:**",
		"pod":        "🐳 **Pod名称:**",
		"pod_ip":     "🌐 **Pod IP:**",
		"node":       "🖥️ **节点名称:**",
		"owner_kind": "🔄 **控制器类型:**",
		"owner_name": "🔧 **控制器名称:**",
	}

	// 保持一个固定的顺序
	orderedKeys := []string{"namespace", "pod", "pod_ip", "node", "owner_kind", "owner_name"}

	for _, key := range orderedKeys {
		if value, ok := alert.Labels[key]; ok {
			fields = append(fields, map[string]string{
				"key":   fieldMapping[key],
				"value": value,
			})
		}
	}
	return fields
}
