package vivo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	vivopush "github.com/cossim/vivo-push"
	"go.uber.org/zap"

	"github.com/cossteam/hipush/pkg/config"
	"github.com/cossteam/hipush/pkg/log"
	"github.com/cossteam/hipush/pkg/provisioner"
	"github.com/cossteam/hipush/pkg/push"
)

const (
	provisionerName = "vivo"
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
			logger.Logger.Error("vivo provisioner config is invalid", zap.Any("config", provisionerConfig))
			return nil, fmt.Errorf("vivo provisioner config is invalid")
		}
		client, err := vivopush.NewClient(cfg.AppId, cfg.AppKey, cfg.AppSecret)
		if err != nil {
			logger.Logger.Error("create vivo client error", zap.Error(err))
			return nil, fmt.Errorf("create vivo client error: %v", err)
		}

		return &Provisioner{
			cfg:    cfg,
			logger: logger,
			client: client,
		}, nil
	})
}

// Config vivo 推送配置
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
	client *vivopush.VivoPush
}

func (p *Provisioner) Validate(message push.Message) error {
	if message.GetTitle() == "" {
		return errors.New("title cannot be empty")
	}

	if message.GetContent() == "" {
		return errors.New("content cannot be empty")
	}

	notifyType := message.GetNotifyType()
	// 检查 NotifyType 是否为有效值
	if notifyType != 0 && notifyType < 1 || notifyType > 4 {
		return errors.New("invalid notify type")
	}

	return nil
}

func (p *Provisioner) Build(token string, message push.Message) interface{} {
	return p.buildNotification(token, message)
}

func (p *Provisioner) Push(ctx context.Context, token string, message push.Message, opts ...push.PushOption) error {
	p.logger.Info("Pushing huawei notification", zap.String("token", token), zap.Any("message", message))

	if err := p.Validate(message); err != nil {
		return err
	}

	notification := p.buildNotification(token, message)

	if p.client == nil {
		return fmt.Errorf("vivo provisioner not initialized")
	}

	_, err := p.client.Send(notification, token)
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

func (p *Provisioner) buildNotification(token string, message push.Message) *vivopush.Message {
	clickAction := message.GetClickAction()
	// Action 点击动作
	// 1 打开 APP 首页
	// 2 打开链接
	// 3 自定义
	// 4 打开 app 内指定页面
	action := 1
	switch clickAction.Action {
	case push.OpenApp:
	case push.OpenActivity:
		action = 4
	case push.OpenURL:
		action = 2
	}

	notifyType := message.GetNotifyType()
	if notifyType == 0 {
		notifyType = 2
	}

	//var pushMode int
	//if so.Development {
	//	pushMode = 1
	//}

	var ttl int64
	if message.GetTTL() == 0 {
		ttl = 60
	}

	data := make(map[string]string)
	for key, value := range message.GetData() {
		if strValue, ok := value.(string); ok {
			data[key] = strValue
		} else {
			data[key] = fmt.Sprintf("%v", value)
		}
	}

	notify := &vivopush.Message{
		RegId:           strings.Join([]string{token}, ","),
		NotifyType:      notifyType,
		Title:           message.GetTitle(),
		Content:         message.GetContent(),
		TimeToLive:      ttl,
		SkipType:        action,
		SkipContent:     clickAction.URL,
		NetworkType:     -1,
		ClientCustomMap: data,
		// Extra:           req.Data.ExtraMap(),
		// RequestId:      req.RequestId,
		// NotifyID:       req.NotifyID,
		Category: message.GetCategory(),
		// PushMode:       pushMode, // 默认为正式推送
		ForegroundShow: message.GetForeground(),
	}
	return notify
}
