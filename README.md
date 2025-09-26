# Prometheus Webhook

一个灵活的 Webhook 服务，用于接收 Prometheus Alertmanager 的告警，并将它们转发到飞书、钉钉和企业微信。

## 特性

- **多渠道支持**: 同时将告警发送到飞书、钉钉和企业微信。
- **高度可定制**: 通过 Go 模板，可以为不同渠道定制丰富的告警消息格式。
- **动态路由**: 根据配置文件自动启用 `/feishu`, `/dingding`, `/weixin` 等 Webhook 端点。
- **高性能**: 基于 Gin 框架构建，轻量且高效。
- **容器化部署**: 提供 `Dockerfile` 和 Kubernetes 部署示例，易于部署和扩展。

## 快速开始

### 1. 配置

复制或重命名 `config/config.yaml.example` 为 `config/config.yaml`，并根据你的需求进行修改。

```yaml
# 服务器配置
server:
  port: "8080"
  timeout: 30s

# 日志配置
logging:
  level: "info" # debug, info, warn, error
  format: "json" # json, text

# 通用模板配置
template:
  timezone: "Asia/Shanghai"

# Webhook 提供商设置
webhooks:
  feishu:
    enable: true
    webhook_url: "your-feishu-webhook-url"
    timeout: 30s
    retry_count: 3
    template: "templates/feishu.tmpl"
  
  dingding:
    enable: true
    webhook_url: "your-dingtalk-webhook-url"
    secret: "your-dingtalk-secret" # 如果启用了加签，请填入密钥
    timeout: 10s
    retry_count: 3
    template: "templates/dingding.tmpl"
    
  weixin:
    enable: false
    webhook_url: "your-weixin-webhook-url"
    timeout: 10s
    retry_count: 3
    template: "templates/weixin.tmpl"
```

### 2. 在 Alertmanager 中配置 Webhook

修改 Alertmanager 的配置文件 (`alertmanager.yml`)，添加 `webhook_configs`，指向你启用的端点。

```yaml
receivers:
- name: 'webhook-receiver'
  webhook_configs:
  - url: 'http://<your-webhook-service-address>:8080/feishu'
    send_resolved: true
  - url: 'http://<your-webhook-service-address>:8080/dingding'
    send_resolved: true
```

### 3. 运行

#### 本地运行

```bash
# 运行服务
go run main.go
```

#### 使用 Docker

```bash
# 构建 Docker 镜像
docker build -t prometheus-webhook .

# 运行容器
docker run -d -p 8080:8080 -v $(pwd)/config:/app/config --name prometheus-webhook prometheus-webhook
```

#### 使用 Makefile

我们提供了一个 `Makefile` 来简化操作。

```bash
# 运行
make run

# 构建所有平台的二进制文件
make all

# 清理构建产物
make clean
```

## 模板定制

你可以通过修改 `templates/` 目录下的 `.tmpl` 文件来定制你自己的告警消息格式。

- `feishu.tmpl`: 飞书消息卡片模板。
- `dingding.tmpl`: 钉钉 Markdown 消息模板。
- `weixin.tmpl`: 企业微信 Markdown 消息模板。

模板中可以使用 `getCSTtime` (格式化时间) 和 `sub` (减法) 等自定义函数。

## 测试

你可以使用以下 `curl` 命令来模拟 Prometheus 发送告警，以测试你的 Webhook 端点是否正常工作。

### 发送告警触发 (firing)

```bash
curl -X POST http://localhost:8080/dingding \
  -H "Content-Type: application/json" \
  -d '{
    "version": "4",
    "groupKey": "test-group",
    "status": "firing",
    "alerts": [
      {
        "status": "firing",
        "labels": {
          "alertname": "Pod 处于非 Running 状态",
          "severity": "critical",
          "namespace": "k8s-app",
          "pod": "prometheus-feishu-webhook-dcf7cccf-f2dp5",
          "pod_ip":"192.18.159.152",
          "owner_name":"prometheus-feishu-webhook-dcf7cccf",
          "owner_kind":"ReplicaSet",
          "node":"k8s-master1"
        },
        "annotations": {
          "message": "这是一个测试告警",
          "summary": "测试告警"
        },
        "startsAt": "2023-07-01T10:00:00Z",
        "endsAt": "2023-07-01T10:05:00Z"
      }
    ]
  }'
```

### 发送告警恢复 (resolved)

```bash
curl -X POST http://localhost:8080/dingding \
  -H "Content-Type: application/json" \
  -d '{
    "version": "4",
    "groupKey": "test-group",
    "status": "resolved",
    "alerts": [
      {
        "status": "resolved",
        "labels": {
          "alertname": "Pod 处于非 Running 状态",
          "severity": "critical",
          "namespace": "k8s-app",
          "pod": "prometheus-feishu-webhook-dcf7cccf-f2dp5",
          "pod_ip":"192.18.159.152",
          "owner_name":"prometheus-feishu-webhook-dcf7cccf",
          "owner_kind":"ReplicaSet",
          "node":"k8s-master1"
        },
        "annotations": {
          "message": "这是一个测试告警",
          "summary": "测试告警"
        },
        "startsAt": "2023-07-01T10:00:00Z",
        "endsAt": "2023-07-01T10:05:00Z"
      }
    ]
  }'
```

**注意**: 你需要将 `http://localhost:8080/dingding` 替换为你想要测试的具体端点，例如 `/feishu` 或 `/weixin`。
