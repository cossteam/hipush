package decorator

import (
	"context"

	"github.com/cossteam/hipush/pkg/log"
	"github.com/cossteam/hipush/pkg/push"
	"go.uber.org/zap"
)

// DryRunPushHandler 装饰器，模拟推送操作，不会真正执行
type DryRunPushHandler struct {
	pushHandler push.ProvisionHandler
	dryRun      bool
	logger      *log.Logger
}

func NewDryRunPushHandler(logger *log.Logger, pushHandler push.ProvisionHandler, dryRun bool) *DryRunPushHandler {
	return &DryRunPushHandler{
		logger:      logger,
		pushHandler: pushHandler,
		dryRun:      dryRun,
	}
}

// Push 执行推送操作，在 DryRun 模式下不会执行实际操作
func (d *DryRunPushHandler) Push(ctx context.Context, token string, message push.Message, opts ...push.PushOption) error {
	if d.dryRun {
		if validator, ok := d.pushHandler.(push.MessageValidator); ok {
			if err := validator.Validate(message); err != nil {
				d.logger.Warn("Message validation failed", zap.String("token", token), zap.Error(err))
				return err
			}
		} else {
			d.logger.Warn("Message does not implement MessageValidator", zap.String("token", token))
		}

		if builder, ok := d.pushHandler.(push.MessageBuilder); ok {
			constructedMessage := builder.Build(token, message)
			d.logger.Debug("Constructed message for dry run", zap.String("token", token), zap.Any("constructedMessage", constructedMessage))
		} else {
			d.logger.Warn("Message does not implement MessageBuilder", zap.String("token", token))
		}

		d.logger.Debug("Dry run: Skipping push operation",
			zap.String("token", token),
			zap.Any("message", message),
		)

		return nil
	}

	return d.pushHandler.Push(ctx, token, message, opts...)
}
