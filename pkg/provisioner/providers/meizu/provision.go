package meizu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	mzp "github.com/cossim/go-meizu-push-sdk"
	"go.uber.org/zap"

	"github.com/cossteam/hipush/pkg/config"
	"github.com/cossteam/hipush/pkg/log"
	"github.com/cossteam/hipush/pkg/provisioner"
	"github.com/cossteam/hipush/pkg/push"
)

const (
	provisionerName = "meizu"
)

func Register(plugins *provisioner.Provisioners) {
	plugins.Register(provisionerName, func(logger *log.Logger, provisionerConfig config.ProvisionerConfig) (push.Provisioner, error) {
		cfg := &Config{
			AppName: provisionerConfig.AppName,
			AppId:   provisionerConfig.AppConfig["appId"],
			AppKey:  provisionerConfig.AppConfig["appKey"],
			Package: provisionerConfig.AppConfig["package"],
		}

		if cfg.AppId == "" || cfg.AppKey == "" {
			logger.Logger.Error("meizu provisioner config is invalid", zap.Any("config", provisionerConfig))
			return nil, fmt.Errorf("meizu provisioner config is invalid")
		}

		pushFunc := func(token, message string) mzp.PushResponse {
			return mzp.PushNotificationMessageByPushId(cfg.AppId, token, message, cfg.AppKey)
		}

		return &Provisioner{
			cfg:    cfg,
			logger: logger,
			push:   pushFunc,
		}, nil
	})
}

// Config oppo 推送配置
type Config struct {
	AppName string
	AppId   string
	AppKey  string
	Package string
}

var (
	_ push.Provisioner      = &Provisioner{}
	_ push.MessageValidator = &Provisioner{}
	_ push.MessageBuilder   = &Provisioner{}
)

type Provisioner struct {
	cfg    *Config
	logger *log.Logger
	push   func(token, message string) mzp.PushResponse
}

func (p *Provisioner) Validate(message push.Message) error {
	return nil
}

func (p *Provisioner) Build(token string, message push.Message) interface{} {
	notification, err := p.buildNotification(message)
	if err != nil {
		return nil
	}
	return notification
}

func (p *Provisioner) Push(ctx context.Context, token string, message push.Message, opts ...push.PushOption) error {
	p.logger.Info("Pushing meizu notification", zap.String("token", token), zap.Any("message", message))

	if err := p.Validate(message); err != nil {
		return err
	}

	notification, err := p.buildNotification(message)
	if err != nil {
		return err
	}

	p.logger.Info("Pushing meizu notification", zap.String("token", token), zap.Any("notification", notification))

	resp := p.push(token, notification)
	// if resp.GetCode() != http.StatusOK {
	var errorDetails struct {
		Code     string `json:"code"`
		Message  string `json:"message"`
		Redirect string `json:"redirect"`
	}
	if err := json.Unmarshal([]byte(resp.GetMessage()), &errorDetails); err != nil {
		return errors.New(resp.GetMessage())
	}
	if errorDetails.Code != "200" {
		return errors.New("code: " + errorDetails.Code + ", message: " + errorDetails.Message)
	}
	//}

	p.logger.Debug("Push meizu notification success", zap.Int("code", resp.GetCode()), zap.String("message", resp.GetMessage()))
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

func (p *Provisioner) buildNotification(message push.Message) (string, error) {
	msg := mzp.BuildNotificationMessage()
	msg.NoticeBarInfo.Title = message.GetTitle()
	msg.NoticeBarInfo.Content = message.GetContent()

	clickAction := message.GetClickAction()
	// Action 点击动作
	// 0 打开应用
	// 1 打开应用页面
	// 2 打开URI页面
	action := 0
	switch clickAction.Action {
	case push.OpenApp:
	case push.OpenActivity:
		action = 1
	case push.OpenURL:
		action = 2
	}

	msg.ClickTypeInfo = mzp.ClickTypeInfo{
		ClickType:  action,
		Url:        clickAction.URL,
		Parameters: clickAction.Parameters,
		Activity:   clickAction.Activity,
	}

	//offLine := 0
	//if req.OffLine {
	//	offLine = 1
	//}
	msg.PushTimeInfo = mzp.PushTimeInfo{
		// OffLine:   offLine,
		ValidTime: int(message.GetTTL()),
	}

	notify, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(notify), nil
}
