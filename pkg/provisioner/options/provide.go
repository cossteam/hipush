package options

import (
	"github.com/cossteam/hipush/pkg/provisioner"
	"github.com/cossteam/hipush/pkg/provisioner/providers/android"
	"github.com/cossteam/hipush/pkg/provisioner/providers/honor"
	"github.com/cossteam/hipush/pkg/provisioner/providers/huawei"
	"github.com/cossteam/hipush/pkg/provisioner/providers/ios"
	"github.com/cossteam/hipush/pkg/provisioner/providers/meizu"
	"github.com/cossteam/hipush/pkg/provisioner/providers/oppo"
	"github.com/cossteam/hipush/pkg/provisioner/providers/vivo"
	"github.com/cossteam/hipush/pkg/provisioner/providers/xiaomi"
)

// RegisterAllProvisioner registers all provisioners.
// The order of registration is irrelevant.
func RegisterAllProvisioner(plugins *provisioner.Provisioners) {
	ios.Register(plugins)
	xiaomi.Register(plugins)
	huawei.Register(plugins)
	vivo.Register(plugins)
	oppo.Register(plugins)
	meizu.Register(plugins)
	android.Register(plugins)
	honor.Register(plugins)
}
