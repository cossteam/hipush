package push

import (
	"context"
)

// Message 消息接口
type Message interface {
	GetTitle() string
	GetContent() string
	GetCategory() string
	GetPriority() Priority
	GetNotifyType() int
	GetIcon() string
	GetSound() string
	GetTTL() int
	GetForeground() bool
	GetClickAction() ClickAction
	GetData() map[string]interface{}
}

type MessageValidator interface {
	Validate(message Message) error
}

type MessageBuilder interface {
	Build(token string, message Message) interface{}
}

type ProvisionHandler interface {
	// Push 推送消息
	Push(ctx context.Context, token string, message Message, opts ...PushOption) error
}

// Provisioner 推送服务提供商接口
type Provisioner interface {
	ProvisionHandler
	// GetType 获取服务提供商类型
	GetType() string
	// GetAppId 获取应用id
	GetAppId() string
	// GetAppName 获取应用名称
	GetAppName() string
}
