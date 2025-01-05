package oppo

import (
	"context"
	"encoding/json"
	"fmt"

	oppopush "github.com/316014408/oppo-push"
	"go.uber.org/zap"

	"github.com/cossteam/hipush/pkg/config"
	"github.com/cossteam/hipush/pkg/log"
	"github.com/cossteam/hipush/pkg/provisioner"
	"github.com/cossteam/hipush/pkg/push"
)

const (
	provisionerName = "oppo"
)

func Register(plugins *provisioner.Provisioners) {
	plugins.Register(provisionerName, func(logger *log.Logger, provisionerConfig config.ProvisionerConfig) (push.Provisioner, error) {
		cfg := &Config{
			AppName:   provisionerConfig.AppName,
			AppId:     provisionerConfig.AppConfig["appId"],
			AppSecret: provisionerConfig.AppConfig["appSecret"],
			AppKey:    provisionerConfig.AppConfig["appKey"],
		}

		if cfg.AppId == "" || cfg.AppSecret == "" || cfg.AppKey == "" {
			logger.Logger.Error("oppo provisioner config is invalid", zap.Any("config", provisionerConfig))
			return nil, fmt.Errorf("oppo provisioner config is invalid")
		}
		client := oppopush.NewClient(cfg.AppKey, cfg.AppSecret)

		return &Provisioner{
			cfg:    cfg,
			logger: logger,
			client: client,
		}, nil
	})
}

// Config oppo 推送配置
type Config struct {
	AppName   string
	AppId     string
	AppSecret string
	AppKey    string
}

var (
	_ push.Provisioner      = &Provisioner{}
	_ push.MessageValidator = &Provisioner{}
	_ push.MessageBuilder   = &Provisioner{}
)

type Provisioner struct {
	cfg    *Config
	logger *log.Logger
	client *oppopush.OppoPush
}

func (p *Provisioner) Validate(message push.Message) error {
	return nil
}

func (p *Provisioner) Build(token string, message push.Message) interface{} {
	notification, err := p.buildNotification(token, message)
	if err != nil {
		return nil
	}
	return notification
}

func (p *Provisioner) Push(ctx context.Context, token string, message push.Message, opts ...push.PushOption) error {
	p.logger.Info("Pushing huawei notification", zap.String("token", token), zap.Any("message", message))

	if err := p.Validate(message); err != nil {
		return err
	}

	notification, err := p.buildNotification(token, message)
	if err != nil {
		return err
	}

	if p.client == nil {
		return fmt.Errorf("vivo provisioner not initialized")
	}

	_, err = p.client.Unicast(notification)
	if err != nil {
		return fmt.Errorf("failed to push vivo notification: %w", err)
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

func (p *Provisioner) buildNotification(token string, message push.Message) (*oppopush.Message, error) {
	notify := oppopush.NewMessage(message.GetTitle(), message.GetContent()).
		// SetSubTitle(req.Subtitle).
		SetTargetType(2)
	notify.SetTargetValue(token)

	clickAction := message.GetClickAction()
	// Action 点击动作
	// Action 点击跳转类型 1：打开APP首页 2：打开链接 3：自定义 4:打开app内指定页面 5:跳转Intentscheme URL   默认值为 0
	// 0 启动应用
	// 1 打开APP首页
	// 2 打开网页
	// 3 自定义
	// 4 打开应用内页（activity 全路径类名）
	// 5 Intentscheme URL
	action := 0
	switch clickAction.Action {
	case push.OpenApp:
	case push.OpenActivity:
		action = 4
		notify.SetClickActionActivity(clickAction.Activity)
	case push.OpenURL:
		action = 2
		notify.SetClickActionUrl(clickAction.URL)
	}

	notify.SetClickActionType(action)
	paramet, err := json.Marshal(clickAction.Parameters)
	if err != nil {
		return nil, err
	}
	notify.SetActionParameters(string(paramet))

	return notify, nil
}
