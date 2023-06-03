package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dop251/goja"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
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

	Gcp struct {
		vu        modules.VU
		jwtConfig *jwt.Config
	}

	GcpConfig struct {
		Scope string
	}

	ServiceAccountKey struct {
		AuthProviderX509CertUrl string `json:"auth_provider_x509_cert_url"`
		AuthURL                 string `json:"auth_uri"`
		ClientEmail             string `json:"client_email" validate:"required"`
		ClientID                string `json:"client_id" validate:"required"`
		ClientSecret            string `json:"client_secret" validate:"required"`
		ClientX509CertUrl       string `json:"client_x509_cert_url" validate:"required"`
		PrivateKey              string `json:"private_key" validate:"required"`
		PrivateKeyID            string `json:"private_key_id" validate:"required"`
		ProjectID               string `json:"project_id" validate:"required"`
		TokenURL                string `json:"token_uri" validate:"required"`
		Type                    string `json:"type" validate:"required"`
		UniverseDomain          string `json:"universe_domain"`
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

// This function is a method of the `Gcp` struct and it returns an OAuth2 token obtained using the JWT
// configuration stored in the `jwtConfig` field of the `Gcp` struct. It first creates a new background
// context using `context.Background()`, then obtains a token source from the JWT configuration using
// `r.jwtConfig.TokenSource(ctx)`, and finally gets the token using `Token()` method of the token
// source. If any error occurs during this process, it returns an error with a formatted message.
func (r *Gcp) GetOAUth2Token() (*oauth2.Token, error) {
	ctx := context.Background()

	token, err := r.jwtConfig.TokenSource(ctx).Token()
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
	const envVar = "GOOGLE_SERVICE_ACCOUNT_KEY"

	if keyString := os.Getenv(envVar); keyString != "" {
		key := &ServiceAccountKey{}

		err := json.Unmarshal([]byte(keyString), &key)
		if err != nil {
			common.Throw(rt, fmt.Errorf("Cannot unmarshal environment variable %v <%w>", envVar, err))
		}

		var options GcpConfig
		err = rt.ExportTo(c.Argument(0), &options)
		if err != nil {
			common.Throw(rt,
				fmt.Errorf("Gcp constructor expects Scope as it's argument: %w", err))
		}

		b, _ := json.Marshal(key)

		jwtConfig, err := google.JWTConfigFromJSON(b, options.Scope)
		if err != nil {
			common.Throw(rt, fmt.Errorf("Failed to obtain JWT Config from Service Account Key <%w>", err))
		}

		obj := &Gcp{
			// vu:        mi.vu,
			jwtConfig: jwtConfig,
		}

		return rt.ToValue(obj).ToObject(rt)

	}

	common.Throw(rt, fmt.Errorf("environment variable %v not found", envVar))

	return nil
}
