package server

import (
	"fmt"
	"os"

	"github.com/cossteam/hipush/internal/handler"
	"github.com/cossteam/hipush/internal/middleware"
	"github.com/cossteam/hipush/pkg/log"
	"github.com/cossteam/hipush/pkg/provisioner/initializer"
	"github.com/gin-gonic/gin"
	oapimiddleware "github.com/oapi-codegen/gin-middleware"
	"go.uber.org/zap"

	v1 "github.com/cossteam/hipush/api/http/v1"
)

func NewServerHTTP(
	logger *log.Logger,
	handler *handler.Handler,
	provisionInitializer initializer.ProvisionerInitializer,
) *gin.Engine {
	if err := provisionInitializer.Initialize(); err != nil {
		logger.Fatal("init provisioner failed", zap.Error(err))
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(
		middleware.CORSMiddleware(),
	)
	swagger, err := v1.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	validatorOptions := &oapimiddleware.Options{
		ErrorHandler: middleware.HandleOpenAPIError,
	}
	// validatorOptions.Options.AuthenticationFunc = func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
	// 	return middleware.HandleOpenApiAuthentication(ctx, h.authService, input)
	// }

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	r.Use(oapimiddleware.OapiRequestValidatorWithOptions(swagger, validatorOptions))
	v1.RegisterHandlers(r, handler)

	return r
}
