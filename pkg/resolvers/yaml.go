package resolvers

import (
    "github.com/ghodss/yaml"
    "net/url"
    vault "github.com/hashicorp/vault/api"
    "fmt"
    . "github.com/nebtex/menshend/pkg/utils"
)
type Backend interface {
    BaseUrl() string
    Headers() map[string]string
    Passed() bool
    Error() error
}

type Resolver interface {
    Resolve(tokenData *vault.Secret) Backend
    SetBody(s string)
}

type YAMLResolver struct {
    Content string `json:"content"`
}
type backendValues struct {
    BaseUrl   string `json:"baseUrl"`
    HeaderMap map[string]string `json:"headersMap"`
    Error     string `json:"error"`
    Passed    *bool `json:"passed"`
}

type SimpleBackend struct {
    ys *backendValues
}

func (yr *YAMLResolver) SetBody(s string) {}

func (ym *SimpleBackend)BaseUrl() string {
    return ym.ys.BaseUrl
}

func (ym *SimpleBackend)Headers() map[string]string {
    return ym.ys.HeaderMap
}

func (ym *SimpleBackend)Error() error {
    return fmt.Errorf("%v", ym.ys.Error)
}

func (ym *SimpleBackend)Passed() bool {
    if ym.ys.Passed == nil {
        return true
    }
    return *ym.ys.Passed
}

func (yr *YAMLResolver)Resolve(u *vault.Secret) (Backend) {
    ys := &backendValues{}
    err := yaml.Unmarshal([]byte(yr.Content), ys)
    HttpCheckPanic(err, InternalError)
    _, parseErr := url.ParseRequestURI(ys.BaseUrl)
    HttpCheckPanic(parseErr, InternalError)
    return &SimpleBackend{ys:ys}
}
