package initializer

import (
	"github.com/cossteam/hipush/pkg/provisioner"
	"github.com/cossteam/hipush/pkg/provisioner/options"
)

type Initializer interface {
	Initialize() error
}

type ProvisionerInitializer struct {
	provisioners *provisioner.Provisioners
}

func NewPluginInitializer(provisioners *provisioner.Provisioners) ProvisionerInitializer {
	return ProvisionerInitializer{
		provisioners: provisioners,
	}
}

func (p ProvisionerInitializer) Initialize() error {
	options.NewProvisionerOptions(p.provisioners)
	return p.provisioners.InitProvisioner()
}
