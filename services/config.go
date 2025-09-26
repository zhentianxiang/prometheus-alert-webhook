package services

import (
	"fmt"
	"os"
	"prometheus-webhook/models"
	"time"

	"gopkg.in/yaml.v3"
)

type ConfigService struct {
	config models.Config
}

func NewConfigService() *ConfigService {
	return &ConfigService{}
}

func (cs *ConfigService) LoadConfig(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cs.config); err != nil {
		return err
	}

	// 设置默认值
	cs.setDefaults()

	// 验证配置
	if err := cs.validateConfig(); err != nil {
		return err
	}

	return nil
}

func (cs *ConfigService) setDefaults() {
	if cs.config.Server.Port == "" {
		cs.config.Server.Port = "8080"
	}
	if cs.config.Server.Timeout == 0 {
		cs.config.Server.Timeout = 30 * time.Second
	}

	cs.setWebhookProviderDefaults(&cs.config.Webhooks.Feishu)
	cs.setWebhookProviderDefaults(&cs.config.Webhooks.Dingding)
	cs.setWebhookProviderDefaults(&cs.config.Webhooks.Weixin)

	if cs.config.Logging.Level == "" {
		cs.config.Logging.Level = "info"
	}
	if cs.config.Template.Timezone == "" {
		cs.config.Template.Timezone = "Asia/Shanghai"
	}
}

func (cs *ConfigService) setWebhookProviderDefaults(provider *models.WebhookProvider) {
	if provider.Timeout == 0 {
		provider.Timeout = 10 * time.Second
	}
	if provider.RetryCount == 0 {
		provider.RetryCount = 3
	}
}

func (cs *ConfigService) validateConfig() error {
	if cs.config.Webhooks.Feishu.Enable {
		if err := cs.validateWebhookProvider("feishu", cs.config.Webhooks.Feishu); err != nil {
			return err
		}
	}
	if cs.config.Webhooks.Dingding.Enable {
		if err := cs.validateWebhookProvider("dingding", cs.config.Webhooks.Dingding); err != nil {
			return err
		}
	}
	if cs.config.Webhooks.Weixin.Enable {
		if err := cs.validateWebhookProvider("weixin", cs.config.Webhooks.Weixin); err != nil {
			return err
		}
	}
	return nil
}

func (cs *ConfigService) validateWebhookProvider(name string, provider models.WebhookProvider) error {
	if provider.WebhookURL == "" {
		return fmt.Errorf("必须为启用的 webhook '%s' 配置 webhook_url", name)
	}
	if provider.Template == "" {
		return fmt.Errorf("必须为启用的 webhook '%s' 配置 template", name)
	}
	return nil
}

func (cs *ConfigService) GetConfig() models.Config {
	return cs.config
}
