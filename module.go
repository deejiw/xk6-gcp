package gcp

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"cloud.google.com/go/pubsub"
	"github.com/dop251/goja"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"google.golang.org/api/sheets/v4"
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
		// vu      modules.VU
		keyByte   []byte
		scope     []string
		projectId string

		// Client
		sheet  *sheets.Service
		pubsub *pubsub.Client
	}

	GcpConfig struct {
		Key       ServiceAccountKey
		Scope     []string
		ProjectId string
	}

	Option func(*Gcp) error

	ServiceAccountKey struct {
		AuthProviderX509CertUrl string `json:"auth_provider_x509_cert_url"`
		AuthURL                 string `json:"auth_uri"`
		ClientEmail             string `json:"client_email"`
		ClientID                string `json:"client_id"`
		ClientSecret            string `json:"client_secret"`
		ClientX509CertUrl       string `json:"client_x509_cert_url"`
		PrivateKey              string `json:"private_key"`
		PrivateKeyID            string `json:"private_key_id"`
		ProjectID               string `json:"project_id"`
		TokenURL                string `json:"token_uri"`
		Type                    string `json:"type"`
		UniverseDomain          string `json:"universe_domain"`
	}
)

var (
	_                          modules.Module   = &RootModule{}
	_                          modules.Instance = &ModuleInstance{}
	gcpConstructorDefaultScope                  = []string{"https://www.googleapis.com/auth/cloud-platform"}
)

func New() *RootModule {
	return &RootModule{}
}

func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &ModuleInstance{
		vu: vu,
	}
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
	const envKey = "GOOGLE_SERVICE_ACCOUNT_KEY"
	var options GcpConfig

	err := rt.ExportTo(c.Argument(0), &options)
	if err != nil {
		common.Throw(rt,
			fmt.Errorf("gcp constructor fails to read options: %w", err))
	}

	g, err := newGcpConstructor(
		withGcpConstructorKey(options.Key, envKey),
		withGcpConstructorScope(options.Scope),
		withGcpConstructorProjectId(options.ProjectId),
	)

	if err != nil {
		common.Throw(rt, fmt.Errorf("cannot initialize gcp constructor <%w>", err))
	}

	return rt.ToValue(g).ToObject(rt)
}

func convertToByte(key interface{}) ([]byte, error) {
	b, err := json.Marshal(key)

	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal key <%w>", err)
	}

	return b, nil
}

// The function creates a new instance of the Gcp struct with specified options.
func newGcpConstructor(opts ...Option) (Gcp, error) {
	g := Gcp{
		scope: gcpConstructorDefaultScope,
	}

	for _, opt := range opts {
		if err := opt(&g); err != nil {
			return Gcp{}, fmt.Errorf("gcp constructor fails to read options %w", err)
		}
	}

	return g, nil
}

func withGcpConstructorKey(_ ServiceAccountKey, _ string) func(*Gcp) error {
	return nil
}

func withGcpConstructorScope(scope []string) func(*Gcp) error {
	return func(g *Gcp) error {
		if len(scope) != 0 {
			g.scope = scope
		}

		return nil
	}
}

func withGcpConstructorProjectId(projectId string) func(*Gcp) error {
	return func(g *Gcp) error {
		if projectId != "" {
			g.projectId = projectId
		} else {
			s := &ServiceAccountKey{}
			err := json.Unmarshal(g.keyByte, s)
			if err != nil {
				log.Fatalf("unable to unmarshal byte <%v>", err)
			}
			g.projectId = s.ProjectID
		}

		return nil
	}
}

func isStructEmpty(object interface{}) bool {
	// check normal definitions of empty
	if object == nil {
		return true
	} else if object == "" {
		return true
	} else if object == false {
		return true
	}

	// see if it's a struct
	if reflect.ValueOf(object).Kind() == reflect.Struct {
		// and create an empty copy of the struct object to compare against
		empty := reflect.New(reflect.TypeOf(object)).Elem().Interface()
		if reflect.DeepEqual(object, empty) {
			return true
		}
	}
	return false
}
