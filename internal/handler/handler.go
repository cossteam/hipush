package handler

import (
	"net/http"

	v1 "github.com/cossteam/hipush/api/http/v1"
	"github.com/cossteam/hipush/pkg/log"
	"github.com/cossteam/hipush/pkg/version"
	"github.com/gin-gonic/gin"
)

var _ v1.ServerInterface = &Handler{}

type Handler struct {
	logger *log.Logger
	*PushHandler
}

func NewHandler(logger *log.Logger, pushHandler *PushHandler) *Handler {
	return &Handler{
		logger:      logger,
		PushHandler: pushHandler,
	}
}

// HelloHipush implements v1.ServerInterface.
func (h *Handler) HelloHipush(c *gin.Context) {
	c.String(http.StatusOK, "Hello Hipush \n%s", version.FullVersion())
}
