package dingding

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"prometheus-webhook/models"
	"time"
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
	var dingTalkMsg map[string]interface{}
	if err := json.Unmarshal([]byte(message), &dingTalkMsg); err != nil {
		return fmt.Errorf("解析模板JSON失败: %w", err)
	}

	jsonData, err := json.Marshal(dingTalkMsg)
	if err != nil {
		return err
	}

	webhookURL := providerConfig.WebhookURL
	if providerConfig.Secret != "" {
		timestamp := time.Now().UnixNano() / 1e6
		signature := s.generateSignature(providerConfig.Secret, timestamp)
		webhookURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhookURL, timestamp, signature)
	}

	for i := 0; i < providerConfig.RetryCount; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), providerConfig.Timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("创建请求失败: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			log.Printf("发送钉钉消息失败 (尝试 %d/%d): %v", i+1, providerConfig.RetryCount, err)
			if i < providerConfig.RetryCount-1 {
				time.Sleep(time.Second * time.Duration(i+1))
			}
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("读取响应失败 (尝试 %d/%d): %v", i+1, providerConfig.RetryCount, err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			if err := json.Unmarshal(body, &result); err != nil {
				log.Printf("解析响应失败 (尝试 %d/%d): %v", i+1, providerConfig.RetryCount, err)
				continue
			}

			if errcode, ok := result["errcode"].(float64); ok && errcode == 0 {
				log.Printf("钉钉消息发送成功到: %s", providerConfig.WebhookURL)
				return nil
			} else {
				log.Printf("钉钉API返回错误 (尝试 %d/%d): %v", i+1, providerConfig.RetryCount, result)
			}
		} else {
			log.Printf("发送钉钉消息失败 (尝试 %d/%d), 状态码: %d, 响应: %s", i+1, providerConfig.RetryCount, resp.StatusCode, string(body))
		}

		if i < providerConfig.RetryCount-1 {
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}

	return fmt.Errorf("发送钉钉消息失败，重试 %d 次后仍然失败", providerConfig.RetryCount)
}

func (s *Service) generateSignature(secret string, timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return url.QueryEscape(signature)
}
