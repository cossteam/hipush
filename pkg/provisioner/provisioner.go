package provisioner

import (
	"fmt"
	"sync"

	"github.com/cossteam/hipush/pkg/config"
	"github.com/cossteam/hipush/pkg/log"
	"github.com/cossteam/hipush/pkg/push"
	"go.uber.org/zap"
)

// InitializationValidator holds ValidateInitialization functions, which are responsible for validation of initialized
// shared resources and should be implemented on admission plugins
type InitializationValidator interface {
	ValidateInitialization() error
}

// Initializer is used for initialization of shareable resources between provisioner plugins.
// After initialization the resources have to be set separately
type Initializer interface {
	Initialize(provisioner push.Provisioner)
}

type Factory func(*log.Logger, config.ProvisionerConfig) (push.Provisioner, error)

type Provisioners struct {
	lock     sync.Mutex
	registry map[string]Factory
	// 存储已初始化的 provisioner 实例
	// platform:appName
	provisioners map[string]push.Provisioner

	logger *log.Logger
	config *config.Config
}

func NewProvisioners(logger *log.Logger, config *config.Config) *Provisioners {
	return &Provisioners{
		registry:     make(map[string]Factory),
		provisioners: make(map[string]push.Provisioner),
		logger:       logger,
		config:       config,
	}
}

// InitProvisioner creates an instance of the named interface.
func (ps *Provisioners) InitProvisioner() error {
	for name := range ps.registry {
		for _, provisionerConfig := range ps.config.Apps {
			if name == provisionerConfig.Platform {
				if !provisionerConfig.Enabled {
					ps.logger.Warn(fmt.Sprintf("Provisioner %q is disabled", name))
					continue
				}
				provisioner, found, err := ps.getProvisioner(name, ps.logger, provisionerConfig)
				if err != nil {
					ps.logger.Error(fmt.Sprintf("couldn't init provisioner %q", name), zap.Error(err))
					continue
				}
				if !found {
					return fmt.Errorf("unknown provisioner: %s", name)
				}
				key := makeKey(provisioner.GetType(), provisioner.GetAppName())
				ps.logger.Info("Provisioner initialized success", zap.String("key", key))
				ps.provisioners[key] = provisioner
			}
		}
	}
	return nil
}

// GetProvisioner 获取已初始化的服务提供商实例
func (ps *Provisioners) GetProvisioner(provType string, appName string) (push.Provisioner, error) {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	key := makeKey(provType, appName)
	p, ok := ps.provisioners[key]
	if !ok {
		return nil, fmt.Errorf("provisioner not found for type %s and app %s", provType, appName)
	}

	return p, nil
}

// getProvisioner creates an instance of the named provisioner.  It returns `false` if
// the name is not known. The error is returned only when the named provider was
// known but failed to initialize.  The config parameter specifies the io.Reader
// handler of the configuration file for the cloud provider, or nil for no configuration.
func (ps *Provisioners) getProvisioner(name string, logger *log.Logger, config2 config.ProvisionerConfig) (push.Provisioner, bool, error) {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	f, found := ps.registry[name]
	if !found {
		return nil, false, nil
	}

	ret, err := f(logger, config2)
	return ret, true, err
}

// Register 注册推送服务提供商
func (ps *Provisioners) Register(name string, factory Factory) {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if factory == nil {
		panic("provisioner: Register init is nil")
	}

	_, found := ps.registry[name]
	if found {
		panic(fmt.Sprintf("Provisioner %q was registered twice", name))
	}

	ps.logger.Info("Register provisioner", zap.String("provisioner", name))
	ps.registry[name] = factory
}

// UnregisterProvisioner 注销服务提供商
func (ps *Provisioners) UnregisterProvisioner(provType string, appName string) {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	key := makeKey(provType, appName)
	delete(ps.registry, key)
	delete(ps.provisioners, key)
}

// makeKey 生成服务提供商的唯一标识
func makeKey(platform, appName string) string {
	return fmt.Sprintf("%s:%s", platform, appName)
}
