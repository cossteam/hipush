package options

import (
	"github.com/cossteam/hipush/pkg/provisioner"
)

type ProvisionerOptions struct{}

func NewProvisionerOptions(provisioners *provisioner.Provisioners) *ProvisionerOptions {
	// register all provisioners
	RegisterAllProvisioner(provisioners)

	return &ProvisionerOptions{}
}
