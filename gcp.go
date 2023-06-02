package interpret

import (
	"fmt"
	"os"

	"github.com/dop251/goja"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"golang.org/x/oauth2/google"
)

func init() {
	modules.Register("k6/x/interpret", New())
}

type (
	// RootModule is the global module instance that will create module
	// instances for each VU.
	RootModule struct{}

	// ModuleInstance represents an instance of the JS module.
	ModuleInstance struct {
		// vu provides methods for accessing internal k6 objects for a VU
		vu modules.VU
	}
)

var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &ModuleInstance{}
)

func New() *RootModule {
	return &RootModule{}
}

func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &ModuleInstance{
		vu: vu,
	}
}

type Gcp struct {
	vu      modules.VU // provides methods for accessing internal k6 objects
	jsonKey []byte
}

// KubeConfig represents the initialization settings for the kubernetes api client.
type ServiceAccountKey struct {
	ConfigPath string
}

func (r *Gcp) GetOauth2Token(sa string) (string, error) {
	ts, err := google.JWTAccessTokenSourceFromJSON(r.jsonKey, "audience")
	if err != nil {
		return "", fmt.Errorf("Failed to get Oauth2 Token from JSON %v <%w>", string(r.jsonKey), err)
	}

	token, err := ts.Token()
	if err != nil {
		return "", fmt.Errorf("Failed to get Token: %v", err)
	}

	return token.AccessToken, nil
}

func (mi *ModuleInstance) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"Gcp": mi.newGcp,
		},
	}
}

func (mi *ModuleInstance) newGcp(c goja.ConstructorCall) *goja.Object {
	rt := mi.vu.Runtime()
	obj := &Gcp{}

	var options ServiceAccountKey
	err := rt.ExportTo(c.Argument(0), &options)
	if err != nil {
		common.Throw(rt,
			fmt.Errorf("Kubernetes constructor expects KubeConfig as it's argument: %w", err))
	}

	const envVar = "GOOGLE_SERVICE_ACCOUNT_KEY"

	if filename := os.Getenv(envVar); filename != "" {
		b, err := os.ReadFile(filename)
		if err != nil {
			common.Throw(rt, fmt.Errorf("google: error getting credentials using %v environment variable: %v", envVar, err))
			obj.jsonKey = b

			return rt.ToValue(obj).ToObject(rt)
		}
	}
	common.Throw(rt, fmt.Errorf("environment variable %v not found", envVar))

	return nil
}
