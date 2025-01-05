package honor

import (
	"context"
	"fmt"

	honorClient "github.com/cossteam/hipush/pkg/client/push"
	"go.uber.org/zap"

	"github.com/cossteam/hipush/pkg/config"
	"github.com/cossteam/hipush/pkg/log"
	"github.com/cossteam/hipush/pkg/provisioner"
	"github.com/cossteam/hipush/pkg/push"
)

const (
	provisionerName = "honor"
)

func Register(plugins *provisioner.Provisioners) {
	plugins.Register(provisionerName, func(logger *log.Logger, provisionerConfig config.ProvisionerConfig) (push.Provisioner, error) {
		cfg := &Config{
			AppName:      provisionerConfig.AppName,
			AppId:        provisionerConfig.AppConfig["appId"],
			ClientId:     provisionerConfig.AppConfig["clientId"],
			ClientSecret: provisionerConfig.AppConfig["clientSecret"],
		}

		if cfg.AppId == "" || cfg.ClientId == "" || cfg.ClientSecret == "" || cfg.AppName == "" {
			logger.Logger.Error("honor provisioner config is invalid", zap.Any("config", provisionerConfig))
			return nil, fmt.Errorf("honor provisioner config is invalid")
		}
		client := honorClient.NewHonorPush(cfg.ClientId, cfg.ClientSecret)

		return &Provisioner{
			cfg:    cfg,
			logger: logger,
			client: client,
		}, nil
	})
}

// Config huawei 推送配置
type Config struct {
	AppName      string
	AppId        string
	ClientId     string
	ClientSecret string
}

var (
	_ push.Provisioner      = &Provisioner{}
	_ push.MessageValidator = &Provisioner{}
	_ push.MessageBuilder   = &Provisioner{}
)

type Provisioner struct {
	cfg    *Config
	logger *log.Logger
	client *honorClient.HonorPushClient
}

func (p *Provisioner) Validate(message push.Message) error {
	return nil
}

func (p *Provisioner) Build(token string, message push.Message) interface{} {
	return p.buildNotification(token, message)
}

func (p *Provisioner) Push(ctx context.Context, token string, message push.Message, opts ...push.PushOption) error {
	pushOption := &push.PushOptions{}
	pushOption.ApplyOptions(opts)

	if err := p.Validate(message); err != nil {
		return err
	}

	p.logger.Info("Pushing huawei notification", zap.String("token", token), zap.Any("message", message))

	notification := p.buildNotification(token, message)

	if p.client == nil {
		return fmt.Errorf("ios provisioner not initialized")
	}

	_, err := p.client.SendMessage(ctx, p.cfg.AppId, notification)
	if err != nil {
		return fmt.Errorf("failed to push huawei notification: %w", err)
	}

	return nil
}

func (p *Provisioner) GetType() string {
	return provisionerName
}

func (p *Provisioner) GetAppId() string {
	return p.cfg.AppId
}

func (p *Provisioner) GetAppName() string {
	return p.cfg.AppName
}

func (p *Provisioner) buildNotification(token string, message push.Message) *honorClient.SendMessageRequest {
	// 构建通知栏消息
	notification := &honorClient.Notification{
		Title: message.GetTitle(),
		Body:  message.GetContent(),
		Image: message.GetIcon(),
	}

	clickAction := message.GetClickAction()
	// Action 点击动作
	// 1 打开应用内页（activity的action标签名）
	// 2 打开特定url
	// 3 打开应用
	action := 1
	switch clickAction.Action {
	case push.OpenApp:
		action = 3
	case push.OpenActivity:
		action = 1
	case push.OpenURL:
		action = 2
	}

	// 构建 Android 平台的通知消息
	androidNotification := &honorClient.AndroidNotification{
		Title: message.GetTitle(),
		Body:  message.GetContent(),
		Image: message.GetIcon(),
		//NotifyID: req.NotifyId,
		//Badge: &hClient.BadgeNotification{
		//	AddNum:     req.Badge.AddNum,
		//	SetNum:     req.Badge.SetNum,
		//	BadgeClass: req.Badge.BadgeClass,
		//},

		ClickAction: &honorClient.ClickAction{
			Type:   action,
			Intent: clickAction.Activity,
			URL:    clickAction.URL,
			Action: clickAction.Activity,
		},
	}

	var targetUserType int

	//if so.Development {
	//	targetUserType = 1
	//}

	// 构建 Android 平台消息推送配置
	androidConfig := &honorClient.AndroidConfig{
		// TTL:            message.GetTTL(),      // 设置消息缓存时间
		BiTag: "", // 设置批量任务消息标识
		// Data:           string(data), // 设置自定义消息负载
		Notification:   androidNotification,
		TargetUserType: targetUserType, // 设置目标用户类型
	}

	// 构建发送消息请求
	sendMessageReq := &honorClient.SendMessageRequest{
		// Data:         string(data), // 设置自定义消息负载
		Notification: notification,
		Android:      androidConfig,
		Token:        []string{token},
	}

	return sendMessageReq
}
