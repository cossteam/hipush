package ios

import (
	"context"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/cossteam/hipush/pkg/config"
	"github.com/cossteam/hipush/pkg/log"
	"github.com/cossteam/hipush/pkg/provisioner"
	"github.com/cossteam/hipush/pkg/push"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
	"go.uber.org/zap"
)

const provisionerName = "ios"

// Register registers a provisioner
func Register(plugins *provisioner.Provisioners) {
	plugins.Register(provisionerName, func(logger *log.Logger, provisionerConfig config.ProvisionerConfig) (push.Provisioner, error) {
		cfg := &Config{
			AppName:    provisionerConfig.AppName,
			BundleID:   provisionerConfig.AppConfig["bundleId"],
			KeyID:      provisionerConfig.AppConfig["keyId"],
			TeamID:     provisionerConfig.AppConfig["teamId"],
			PrivateKey: provisionerConfig.AppConfig["privateKey"],
			IsSandbox:  provisionerConfig.AppConfig["environment"] == "sandbox",
		}

		p := &Provisioner{
			logger: logger,
			cfg:    cfg,
		}

		// 解析私钥
		block, _ := pem.Decode([]byte(cfg.PrivateKey))
		if block == nil {
			return nil, fmt.Errorf("failed to decode private key")
		}

		key, err := token.AuthKeyFromBytes(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		// 创建 JWT token
		authToken := &token.Token{
			AuthKey: key,
			KeyID:   cfg.KeyID,
			TeamID:  cfg.TeamID,
		}

		// 创建 APNS 客户端
		if cfg.IsSandbox {
			p.client = apns2.NewTokenClient(authToken).Development()
		} else {
			p.client = apns2.NewTokenClient(authToken).Production()
		}

		return p, nil
	})
}

// Config iOS 推送配置
type Config struct {
	AppName    string // 应用名称
	BundleID   string // 应用的 Bundle ID
	KeyID      string // APNs Key ID
	TeamID     string // Apple Developer Team ID
	PrivateKey string // APNs 私钥（P8 格式）
	IsSandbox  bool   // 是否为沙箱环境
}

var (
	_ push.Provisioner                    = &Provisioner{}
	_ provisioner.InitializationValidator = &Provisioner{}
	_ push.MessageValidator               = &Provisioner{}
	_ push.MessageBuilder                 = &Provisioner{}
)

type Provisioner struct {
	cfg    *Config
	client *apns2.Client
	logger *log.Logger
}

func (p *Provisioner) ValidateInitialization() error {
	return nil
}

func (p *Provisioner) GetType() string {
	return provisionerName
}

func (p *Provisioner) GetAppId() string {
	return p.cfg.AppName
}

func (p *Provisioner) GetAppName() string {
	return p.cfg.AppName
}

func (p *Provisioner) Validate(message push.Message) error {
	return nil
}

func (p *Provisioner) Build(token string, message push.Message) interface{} {
	return p.buildNotification(token, message)
}

func (p *Provisioner) Push(ctx context.Context, token string, message push.Message, opts ...push.PushOption) error {
	p.logger.Info("Pushing ios notification", zap.String("token", token), zap.Any("message", message))

	if err := p.Validate(message); err != nil {
		return err
	}

	notification := p.buildNotification(token, message)

	if p.client == nil {
		return fmt.Errorf("ios provisioner not initialized")
	}

	// 发送通知
	res, err := p.client.PushWithContext(ctx, notification)
	if err != nil {
		return fmt.Errorf("failed to push notification: %w", err)
	}

	// 检查响应
	if !res.Sent() {
		return fmt.Errorf("notification rejected by APNs: %v", res.Reason)
	}

	p.logger.Debug("Successfully sent iOS push notification",
		zap.String("token", token),
		zap.String("messageID", res.ApnsID),
		zap.Int("statusCode", res.StatusCode),
	)

	return nil
}

func (p *Provisioner) buildNotification(token string, message push.Message) *apns2.Notification {
	// 创建 APNs 负载
	payload := payload.NewPayload()

	// 如果消息长度> 0且标题为空，则添加警报对象
	if len(message.GetContent()) > 0 && message.GetTitle() == "" {
		payload.Alert(message.GetContent())
	}

	// 设置提醒内容
	payload.AlertTitle(message.GetTitle())
	payload.AlertBody(message.GetContent())
	payload.Category(message.GetCategory())

	// 设置声音
	if sound := message.GetSound(); sound != "" {
		payload.Sound(sound)
	}

	// 设置分类
	if category := message.GetCategory(); category != "" {
		payload.Category(category)
	}

	// 设置自定义数据
	if data := message.GetData(); data != nil {
		for k, v := range data {
			payload.Custom(k, v)
		}
	}

	// 设置优先级
	priority := apns2.PriorityHigh
	switch message.GetPriority() {
	case push.Low:
		priority = apns2.PriorityLow
	case push.Normal:
		priority = apns2.PriorityHigh
	case push.High:
		priority = apns2.PriorityHigh
	}

	// 创建通知
	notification := &apns2.Notification{
		DeviceToken: token,
		Topic:       p.cfg.BundleID,
		Payload:     payload,
		Priority:    priority,
	}

	// 设置消息过期时间
	if ttl := message.GetTTL(); ttl > 0 {
		notification.Expiration = time.Unix(int64(message.GetTTL()), 0)
	}

	return notification
}
