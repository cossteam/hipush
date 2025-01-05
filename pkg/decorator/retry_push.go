package decorator

import (
	"context"
	"fmt"
	"time"

	"github.com/cossteam/hipush/pkg/log"
	"github.com/cossteam/hipush/pkg/push"
	"go.uber.org/zap"
)

// RetryPushHandler 装饰器，添加重试逻辑
type RetryPushHandler struct {
	logger        *log.Logger
	pushHandler   push.ProvisionHandler
	maxAttempts   int
	retryInterval time.Duration
}

// NewRetryPushHandler 创建一个新的 RetryPushHandler
func NewRetryPushHandler(logger *log.Logger, pushHandler push.ProvisionHandler, maxAttempts int, retryInterval time.Duration) *RetryPushHandler {
	return &RetryPushHandler{
		logger:        logger,
		pushHandler:   pushHandler,
		maxAttempts:   maxAttempts,
		retryInterval: retryInterval,
	}
}

// Push 执行推送操作，并支持重试逻辑
func (r *RetryPushHandler) Push(ctx context.Context, token string, message push.Message, opts ...push.PushOption) error {
	var err error
	for i := 0; i < r.maxAttempts; i++ {
		if err = r.pushHandler.Push(ctx, token, message, opts...); err == nil {
			return nil
		}
		r.logger.Warn("Attempt failed", zap.Int("attempt", i+1), zap.String("token", token), zap.Error(err))
		time.Sleep(r.retryInterval)
	}
	return fmt.Errorf("push failed after %d attempts: %w", r.maxAttempts, err)
}
