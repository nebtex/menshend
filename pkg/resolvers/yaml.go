package resolvers

import (
    . "github.com/nebtex/menshend/pkg/backend"
    "gopkg.in/yaml.v2"
    . "github.com/nebtex/menshend/pkg/apis/menshend/v1"
    . "github.com/nebtex/menshend/pkg/utils"
    . "github.com/nebtex/menshend/pkg/apis/menshend"
    "net/url"
    vault "github.com/hashicorp/vault/api"
)

type Resolver interface {
    Resolve(*AdminServiceResource, *vault.Secret) Backend
}

type YAMLResolve struct {
    
}
type backendValues struct {
    BaseUrl    string `yaml:"baseUrl"`
    HeadersMap map[string]string `yaml:"headersMap"`
}

type backendImplementation struct {
    ys *backendValues
}

func (ym *backendImplementation)BaseUrl() string {
    return ym.ys.BaseUrl
}

func (ym *backendImplementation)Headers() map[string]string {
    return ym.ys.HeadersMap
}

func (yr *YAMLResolve)Resolve(c *AdminServiceResource, u *vault.Secret) (*backendImplementation) {
    ys := &backendValues{}
    err := yaml.Unmarshal([]byte(c.ProxyCode), ys)
    HttpCheckPanic(err, InternalError)
    _, parseErr := url.ParseRequestURI(ys.BaseUrl)
    HttpCheckPanic(parseErr, InternalError.WithValue("user", u).WithValue("service", c))
    return &backendImplementation{ys:ys}
}
