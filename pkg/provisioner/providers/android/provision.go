package android

import (
	"context"
	"fmt"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"go.uber.org/zap"
	"google.golang.org/api/option"

	"github.com/cossteam/hipush/pkg/config"
	"github.com/cossteam/hipush/pkg/log"
	"github.com/cossteam/hipush/pkg/provisioner"
	"github.com/cossteam/hipush/pkg/push"
)

const (
	provisionerName = "android"
)

func Register(plugins *provisioner.Provisioners) {
	plugins.Register(provisionerName, func(logger *log.Logger, provisionerConfig config.ProvisionerConfig) (push.Provisioner, error) {
		cfg := &Config{
			AppName: provisionerConfig.AppName,
			AppId:   provisionerConfig.AppConfig["appId"],
			AppKey:  provisionerConfig.AppConfig["appKey"],
			KeyPath: provisionerConfig.AppConfig["keyPath"],
		}

		opt := option.WithCredentialsFile(cfg.KeyPath)
		app, err := firebase.NewApp(context.Background(), nil, opt)
		if err != nil {
			logger.Logger.Error("android provisioner config is invalid", zap.Any("config", provisionerConfig))
			return nil, fmt.Errorf("android provisioner config is invalid")
		}

		client, err := app.Messaging(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to create firebase messaging client: %w", err)
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
	AppName string `yaml:"appName"`
	AppId   string `yaml:"appId"`
	AppKey  string `yaml:"appKey"`
	KeyPath string `yaml:"keyPath"`
}

var (
	_ push.Provisioner      = &Provisioner{}
	_ push.MessageValidator = &Provisioner{}
	_ push.MessageBuilder   = &Provisioner{}
)

type Provisioner struct {
	cfg    *Config
	logger *log.Logger
	client *messaging.Client
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

	_, err := p.client.Send(ctx, notification)
	if err != nil {
		return fmt.Errorf("failed to push android notification: %w", err)
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

// buildNotification use for define Android notification.
// HTTP Connection Server Reference for Android
// https://firebase.google.com/docs/cloud-messaging/http-server-ref
func (p *Provisioner) buildNotification(token string, message push.Message) *messaging.Message {
	notification := &messaging.Message{
		Token: token,
		// Topic:     message.GetTopic(),
		// Condition: message.GetCondition(),
		Android: &messaging.AndroidConfig{},
	}

	//if len(req.GetToken()) > 0 {
	//notification.Token = ""
	//notification.Token = req.Tokens[0]
	//}

	//if message.GetCategory() != "" {
	//notification.Android.CollapseKey = message.GetCollapseID()
	//}

	if message.GetPriority() == "high" || message.GetPriority() == "normal" {
		notification.Android.Priority = string(message.GetPriority())
	}

	customData := message.GetData()

	// Add another field
	if len(customData) > 0 {
		data := make(map[string]string)
		for k, v := range customData {
			data[k] = fmt.Sprintf("%v", v)
		}
		notification.Data = data
	}

	duration := time.Duration(message.GetTTL()) * time.Second
	notification.Android.TTL = &duration

	n := &messaging.Notification{}
	isNotificationSet := false

	if len(message.GetContent()) > 0 {
		isNotificationSet = true
		n.Body = message.GetContent()
	}

	if len(message.GetTitle()) > 0 {
		isNotificationSet = true
		n.Title = message.GetTitle()
	}

	if len(message.GetIcon()) > 0 {
		isNotificationSet = true
		n.ImageURL = message.GetIcon()
	}

	//if req.GetSound() != "" {
	//	isNotificationSet = true
	//	//n.Sound = req.Sound
	//	sound, ok := req.GetSound().(string)
	//	if ok {
	//		notification.Android.Notification.Sound = sound
	//	}
	//}

	if isNotificationSet {
		notification.Notification = n
	}

	// handle iOS apns in fcm
	//if len(req.Apns) > 0 {
	// Handle iOS APNS
	//}

	return notification
}
