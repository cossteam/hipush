//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/cossteam/hipush/internal/handler"
	"github.com/cossteam/hipush/internal/server"
	"github.com/cossteam/hipush/pkg/config"
	"github.com/cossteam/hipush/pkg/log"
	"github.com/cossteam/hipush/pkg/provisioner"
	"github.com/cossteam/hipush/pkg/provisioner/initializer"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var ProvisionerSet = wire.NewSet(provisioner.NewProvisioners, initializer.NewPluginInitializer)

var ServerSet = wire.NewSet(server.NewServerHTTP)

var HandlerSet = wire.NewSet(
	handler.NewPushHandler,
	handler.NewHandler,
)

func NewWire(*config.Config, *log.Logger) (*gin.Engine, func(), error) {
	panic(wire.Build(
		ProvisionerSet,
		ServerSet,
		HandlerSet,
	))
}
