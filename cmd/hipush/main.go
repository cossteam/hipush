package main

import (
	"github.com/cossteam/hipush/cmd/hipush/wire"
	"github.com/cossteam/hipush/pkg/config"
	"github.com/cossteam/hipush/pkg/http"
	"github.com/cossteam/hipush/pkg/log"
	"go.uber.org/zap"
)

func main() {
	// conf := config.NewConfig()
	conf := config.Load()
	logger := log.NewLog(conf)

	logger.Info("Hipush start", zap.String("host", conf.HTTP.Addr))

	app, cleanup, err := wire.NewWire(conf, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	http.Run(app, conf.HTTP.Addr)
}
