package strategy

import (
    . "github.com/nebtex/menshend/pkg/utils"
    "net/http"
    "net/url"
    "github.com/nebtex/menshend/pkg/resolvers"
    vault "github.com/hashicorp/vault/api"
)

type Redirect struct {
    
}

func (r *Redirect) Execute(rs resolvers.Resolver, tokenInfo *vault.Secret) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        b := rs.Resolve(tokenInfo)
        if !b.Passed() {
            panic(NotAuthorized.WithUserMessage(b.Error().Error()))
        }
        for key, value := range b.Headers() {
            w.Header().Set(key, value)
        }
        newUrl, err := url.Parse(r.URL.String())
        HttpCheckPanic(err, InternalError)
        bUrl, err := url.Parse(b.BaseUrl())
        
        newUrl.Host = bUrl.Host
        newUrl.User = bUrl.User
        newUrl.Scheme = bUrl.Scheme
        http.Redirect(w, r, newUrl.String(), 301)
    })
}
