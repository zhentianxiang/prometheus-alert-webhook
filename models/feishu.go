package models

// FeishuMessage 飞书消息结构
type FeishuMessage struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

// FeishuInteractiveMessage 飞书消息卡片结构
type FeishuInteractiveMessage struct {
	MsgType string      `json:"msg_type"`
	Card    interface{} `json:"card"`
}
