package huawei

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	hmsConfig "github.com/cossim/go-hms-push/push/config"
	hmsClient "github.com/cossim/go-hms-push/push/core"
	"github.com/cossim/go-hms-push/push/model"
	"go.uber.org/zap"

	"github.com/cossteam/hipush/pkg/config"
	"github.com/cossteam/hipush/pkg/log"
	"github.com/cossteam/hipush/pkg/provisioner"
	"github.com/cossteam/hipush/pkg/push"
)

const (
	provisionerName = "huawei"

	DefaultAuthUrl = "https://oauth-login.cloud.huawei.com/oauth2/v3/token"
	DefaultPushUrl = "https://push-api.cloud.huawei.com"
)

func Register(plugins *provisioner.Provisioners) {
	plugins.Register(provisionerName, func(logger *log.Logger, provisionerConfig config.ProvisionerConfig) (push.Provisioner, error) {
		cfg := &Config{
			AppName:   provisionerConfig.AppName,
			AppId:     provisionerConfig.AppConfig["appId"],
			AppSecret: provisionerConfig.AppConfig["appSecret"],
			AuthUrl:   provisionerConfig.AppConfig["authUrl"],
			PushUrl:   provisionerConfig.AppConfig["pushUrl"],
		}

		if cfg.AuthUrl == "" {
			cfg.AuthUrl = DefaultAuthUrl
		}
		if cfg.PushUrl == "" {
			cfg.PushUrl = DefaultPushUrl
		}
		if cfg.AppId == "" || cfg.AppSecret == "" {
			logger.Logger.Error("huawei provisioner config is invalid", zap.Any("config", provisionerConfig))
			return nil, fmt.Errorf("huawei provisioner config is invalid")
		}
		client, err := hmsClient.NewHttpClient(&hmsConfig.Config{
			AppId:     cfg.AppId,
			AppSecret: cfg.AppSecret,
			AuthUrl:   cfg.AuthUrl,
			PushUrl:   cfg.PushUrl,
		})
		if err != nil {
			return nil, fmt.Errorf("create huawei client error: %v", err)
		}

		return &Provisioner{
			cfg:    cfg,
			logger: logger,
			client: client,
		}, nil
	})
}

// Config huawei 推送配置
type Config struct {
	AppName   string
	AppId     string
	AppSecret string
	AuthUrl   string
	PushUrl   string
}

var (
	_ push.Provisioner      = &Provisioner{}
	_ push.MessageValidator = &Provisioner{}
	_ push.MessageBuilder   = &Provisioner{}
)

type Provisioner struct {
	cfg    *Config
	logger *log.Logger
	client *hmsClient.HMSClient
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

	notification := p.buildNotification(token, message)

	if p.client == nil {
		return fmt.Errorf("ios provisioner not initialized")
	}

	p.logger.Debug("Pushing huawei notification", zap.String("token", token), zap.Any("notification", notification))

	_, err := p.client.SendMessage(ctx, notification)
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

func (p *Provisioner) buildNotification(token string, message push.Message) *model.MessageRequest {
	msgRequest := model.NewNotificationMsgRequest()

	msgRequest.Message.Android = model.GetDefaultAndroid()
	msgRequest.Message.Token = []string{token}

	//if len(req.GetTopic()) > 0 {
	//	msgRequest.Message.Topic = req.GetTopic()
	//}

	//if len(req.GetCondition()) > 0 {
	//	msgRequest.Message.Condition = req.GetCondition()
	//}

	//if len(message.GetPriority()) > 0 {
	//msgRequest.Message.Android.Urgency = message.GetPriority()
	//}

	if len(message.GetCategory()) > 0 {
		msgRequest.Message.Android.Category = message.GetCategory()
	}

	if message.GetTTL() > 0 {
		duration := time.Duration(message.GetTTL()) * time.Second
		msgRequest.Message.Android.TTL = duration.String()
	}

	//if len(req.BiTag) > 0 {
	//	msgRequest.Message.Android.BiTag = req.BiTag
	//}

	// msgRequest.Message.Android.FastAppTarget = req.FastAppTarget

	// Add data fields
	if len(message.GetData()) > 0 {
		jsonBytes, err := json.Marshal(message.GetData())
		if err == nil {
			msgRequest.Message.Data = string(jsonBytes)
		}
	}

	// Notification Content
	//if req.MessageRequest.Message.Android.Notification != nil {
	//	msgRequest.Message.Android.Notification = req.MessageRequest.Message.Android.Notification
	//
	//	if msgRequest.Message.Android.Notification.ClickAction == nil {
	//		msgRequest.Message.Android.Notification.ClickAction = model.GetDefaultClickAction()
	//	}
	//}

	setDefaultAndroidNotification := func() {
		if msgRequest.Message.Android.Notification == nil {
			msgRequest.Message.Android.Notification = model.GetDefaultAndroidNotification()
		}
	}

	if len(message.GetContent()) > 0 {
		setDefaultAndroidNotification()
		msgRequest.Message.Android.Notification.Body = message.GetContent()
	}

	if len(message.GetTitle()) > 0 {
		setDefaultAndroidNotification()
		msgRequest.Message.Android.Notification.Title = message.GetTitle()
	}

	if len(message.GetIcon()) > 0 {
		setDefaultAndroidNotification()
		msgRequest.Message.Android.Notification.Image = message.GetIcon()
	}

	//if v, ok := req.Sound.(string); ok && len(v) > 0 {
	//	setDefaultAndroidNotification()
	//	msgRequest.Message.Android.Notification.Sound = v
	//} else if msgRequest.Message.Android.Notification != nil {
	//	msgRequest.Message.Android.Notification.DefaultSound = true
	//}
	msgRequest.Message.Android.Notification.DefaultSound = true

	//if so.Development {
	//	msgRequest.Message.Android.TargetUserType = 1
	//}

	return msgRequest
}
