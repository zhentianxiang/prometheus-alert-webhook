package models

import "time"

// Config 配置文件结构体
type Config struct {
	Server struct {
		Port    string        `yaml:"port"`
		Timeout time.Duration `yaml:"timeout"`
	} `yaml:"server"`

	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`

	Template struct {
		Timezone string `yaml:"timezone"`
	} `yaml:"template"`

	Webhooks struct {
		Feishu   WebhookProvider `yaml:"feishu"`
		Dingding WebhookProvider `yaml:"dingding"`
		Weixin   WebhookProvider `yaml:"weixin"`
	} `yaml:"webhooks"`
}

// WebhookProvider 定义了单个 webhook 提供商的配置
type WebhookProvider struct {
	Enable     bool          `yaml:"enable"`
	WebhookURL string        `yaml:"webhook_url"`
	Secret     string        `yaml:"secret,omitempty"` // 用于钉钉签名
	Timeout    time.Duration `yaml:"timeout"`
	RetryCount int           `yaml:"retry_count"`
	Template   string        `yaml:"template"`
}
