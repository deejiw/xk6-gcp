package gcp

import (
	"context"
	"fmt"
	"os"

	"github.com/dop251/goja"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func init() {
	modules.Register("k6/x/gcp", New())
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

func (r *Gcp) GetOAUth2Token(scope string) (*oauth2.Token, error) {
	ctx := context.Background()
	ts, err := google.JWTConfigFromJSON(r.jsonKey, scope)
	if err != nil {
		return nil, fmt.Errorf("Failed to obtain JWT Config from Service Account Key <%w>", err)
	}

	token, err := ts.TokenSource(ctx).Token()
	if err != nil {
		return nil, fmt.Errorf("Failed to obtain Access Token <%w>", err)
	}

	return token, nil
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

	const envVar = "GOOGLE_SERVICE_ACCOUNT_KEY"

	if jsonKey := os.Getenv(envVar); jsonKey != "" {
		obj.jsonKey = []byte(jsonKey)
		return rt.ToValue(obj).ToObject(rt)
	}

	common.Throw(rt, fmt.Errorf("environment variable %v not found", envVar))

	return nil
}
