package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"prometheus-webhook/models"
)

type Service struct {
	httpClient *http.Client
}

func NewService() *Service {
	return &Service{
		httpClient: &http.Client{},
	}
}

func (s *Service) SendMessage(providerConfig models.WebhookProvider, message string) error {
	// 解析消息，可能是单个卡片或卡片数组
	var feishuMessages []models.FeishuInteractiveMessage
	var singleMsg models.FeishuInteractiveMessage

	// 首先尝试解析为数组
	if err := json.Unmarshal([]byte(message), &feishuMessages); err != nil {
		// 如果解析为数组失败，尝试解析为单个消息
		if err := json.Unmarshal([]byte(message), &singleMsg); err != nil {
			return fmt.Errorf("解析模板JSON失败: %w", err)
		}
		feishuMessages = []models.FeishuInteractiveMessage{singleMsg}
	}

	// 发送每个独立的卡片消息
	for msgIndex, feishuMsg := range feishuMessages {
		jsonData, err := json.Marshal(feishuMsg)
		if err != nil {
			log.Printf("序列化第 %d 个消息失败: %v", msgIndex+1, err)
			continue
		}

		success := false
		for i := 0; i < providerConfig.RetryCount; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), providerConfig.Timeout)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, "POST", providerConfig.WebhookURL, bytes.NewBuffer(jsonData))
			if err != nil {
				log.Printf("创建第 %d 个消息请求失败: %v", msgIndex+1, err)
				break
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := s.httpClient.Do(req)
			if err != nil {
				log.Printf("发送第 %d 个飞书消息失败 (尝试 %d/%d): %v", msgIndex+1, i+1, providerConfig.RetryCount, err)
				if i < providerConfig.RetryCount-1 {
					time.Sleep(time.Second * time.Duration(i+1))
				}
				continue
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("读取第 %d 个消息响应失败 (尝试 %d/%d): %v", msgIndex+1, i+1, providerConfig.RetryCount, err)
				if i < providerConfig.RetryCount-1 {
					time.Sleep(time.Second * time.Duration(i+1))
				}
				continue
			}

			if resp.StatusCode == http.StatusOK {
				var result map[string]interface{}
				if err := json.Unmarshal(body, &result); err != nil {
					log.Printf("解析第 %d 个消息响应失败 (尝试 %d/%d): %v", msgIndex+1, i+1, providerConfig.RetryCount, err)
					if i < providerConfig.RetryCount-1 {
						time.Sleep(time.Second * time.Duration(i+1))
					}
					continue
				}
				// 检查飞书返回的具体业务错误码
				if code, ok := result["code"].(float64); ok && code == 0 {
					log.Printf("第 %d 个飞书消息发送成功到: %s", msgIndex+1, providerConfig.WebhookURL)
					success = true
					break
				} else {
					log.Printf("第 %d 个飞书API返回错误 (尝试 %d/%d): %v", msgIndex+1, i+1, providerConfig.RetryCount, result)
				}
			} else {
				log.Printf("发送第 %d 个飞书消息失败 (尝试 %d/%d), 状态码: %d, 响应: %s", msgIndex+1, i+1, providerConfig.RetryCount, resp.StatusCode, string(body))
			}

			if i < providerConfig.RetryCount-1 {
				time.Sleep(time.Second * time.Duration(i+1))
			}
		}

		if !success {
			log.Printf("第 %d 个飞书消息发送失败，重试 %d 次后仍然失败", msgIndex+1, providerConfig.RetryCount)
		}
	}

	return nil
}
