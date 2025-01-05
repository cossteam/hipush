package handler

import (
	"net/http"
	"time"

	"github.com/cossteam/hipush/pkg/adapter"
	"github.com/cossteam/hipush/pkg/decorator"
	"github.com/cossteam/hipush/pkg/push"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/cossteam/hipush/api/http/v1"
	"github.com/cossteam/hipush/pkg/log"
	"github.com/cossteam/hipush/pkg/provisioner"
)

func NewPushHandler(logger *log.Logger, provisioners *provisioner.Provisioners) *PushHandler {
	return &PushHandler{
		logger:       logger,
		provisioners: provisioners,
	}
}

type PushHandler struct {
	logger       *log.Logger
	provisioners *provisioner.Provisioners
}

func (h *PushHandler) PushMessage(c *gin.Context, platform string, appName string, token string, params v1.PushMessageParams) {
	req := &v1.MessageRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		h.logger.Error("failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, v1.Response{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	prov, err := h.provisioners.GetProvisioner(platform, appName)
	if err != nil {
		h.logger.Error("failed to get provisioner", zap.Error(err))
		c.JSON(http.StatusBadRequest, v1.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	msg := adapter.NewMessageAdapter(req)

	pushHandler := h.createPushHandler(prov, params)

	if err := pushHandler.Push(c, token, msg); err != nil {
		h.logger.Error("failed to push message",
			zap.String("platform", platform),
			zap.String("appName", appName),
			zap.String("token", token),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, v1.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, v1.Response{
		Code:    http.StatusOK,
		Message: "success",
	})
}

// createPushHandler initializes pushHandler with decorators based on the provided parameters.
func (h *PushHandler) createPushHandler(prov push.ProvisionHandler, params v1.PushMessageParams) push.ProvisionHandler {
	pushHandler := prov

	// Apply DryRun decorator if needed
	if params.DryRun != nil && *params.DryRun {
		pushHandler = decorator.NewDryRunPushHandler(h.logger, pushHandler, true)
	}

	// Apply Retry decorator if needed
	if params.Retry != nil && *params.Retry > 0 {
		retryInterval := 1 * time.Second // default retry interval
		if params.RetryInterval != nil {
			retryInterval = time.Duration(*params.RetryInterval) * time.Second
		}
		pushHandler = decorator.NewRetryPushHandler(h.logger, pushHandler, *params.Retry, retryInterval)
	}

	return pushHandler
}
