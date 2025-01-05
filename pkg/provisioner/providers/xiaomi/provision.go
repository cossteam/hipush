package xiaomi

import (
	"context"
	"fmt"
	"strings"

	xp "github.com/cossim/xiaomi-push"
	"go.uber.org/zap"

	"github.com/cossteam/hipush/pkg/config"
	"github.com/cossteam/hipush/pkg/log"
	"github.com/cossteam/hipush/pkg/provisioner"
	"github.com/cossteam/hipush/pkg/push"
)

const provisionerName = "xiaomi"

// Register registers a provisioner
func Register(plugins *provisioner.Provisioners) {
	plugins.Register(provisionerName, func(logger *log.Logger, provisionerConfig config.ProvisionerConfig) (push.Provisioner, error) {
		v := &Config{
			AppName:   provisionerConfig.AppName,
			AppID:     provisionerConfig.AppConfig["appId"],
			AppSecret: provisionerConfig.AppConfig["appSecret"],
			Package:   strings.Split(provisionerConfig.AppConfig["package"], ","),
		}

		if v.AppID == "" || v.AppSecret == "" {
			logger.Logger.Error("xiaomi provisioner config is invalid", zap.Any("config", provisionerConfig))
			return nil, fmt.Errorf("xiaomi provisioner config is invalid")
		}
		client := xp.NewClient(v.AppSecret, v.Package)

		return &Provisioner{
			cfg:    v,
			logger: logger,
			client: client,
		}, nil
	})
}

type Config struct {
	AppName   string   `yaml:"appName"`
	AppID     string   `yaml:"appId"`
	AppSecret string   `yaml:"appSecret"`
	Package   []string `yaml:"package"`
}

var (
	_ push.Provisioner      = &Provisioner{}
	_ push.MessageValidator = &Provisioner{}
	_ push.MessageBuilder   = &Provisioner{}
)

type Provisioner struct {
	cfg    *Config
	logger *log.Logger
	client *xp.MiPush
}

func (p *Provisioner) Validate(message push.Message) error {
	return nil
}

func (p *Provisioner) Build(token string, message push.Message) interface{} {
	return p.buildNotification(message)
}

func (p *Provisioner) Push(ctx context.Context, token string, message push.Message, opts ...push.PushOption) error {
	p.logger.Info("Pushing Xiaomi notification", zap.String("token", token), zap.Any("message", message))

	if err := p.Validate(message); err != nil {
		return err
	}

	notification := p.buildNotification(message)

	if p.client == nil {
		return fmt.Errorf("xiaomi provisioner not initialized")
	}

	_, err := p.client.Send(ctx, notification, token)
	if err != nil {
		return fmt.Errorf("failed to push xiaomi notification: %w", err)
	}

	return nil
}

func (p *Provisioner) GetType() string {
	return provisionerName
}

func (p *Provisioner) GetAppId() string {
	return p.cfg.AppID
}

func (p *Provisioner) GetAppName() string {
	return p.cfg.AppName
}

func (p *Provisioner) buildNotification(message push.Message) *xp.Message {
	msg := xp.NewAndroidMessage(message.GetTitle(), message.GetContent())
	msg.SetNotifyType(int32(message.GetNotifyType()))

	if message.GetTTL() != 0 {
		msg.SetTimeToLive(int64(message.GetTTL()))
	}

	if message.GetForeground() {
		msg.Extra["notify_foreground"] = "1"
	} else {
		msg.Extra["notify_foreground"] = "0"
	}

	return msg
}
