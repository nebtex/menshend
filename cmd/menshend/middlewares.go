package main

import (
    "net/http"
    "context"
    . "github.com/nebtex/menshend/pkg/users"
    . "github.com/nebtex/menshend/pkg/utils"
    . "github.com/nebtex/menshend/pkg/config"
    . "github.com/nebtex/menshend/pkg/apis/menshend"
    vault "github.com/hashicorp/vault/api"
    "fmt"
    "github.com/nebtex/menshend/pkg/apis/menshend/v1"
    "github.com/mitchellh/mapstructure"
    "github.com/nebtex/menshend/pkg/resolvers"
    "github.com/nebtex/menshend/pkg/strategy"
)

func DetectBrowser(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ibr := false
        if len(r.Cookies()) > 0 {
            ibr = true
        }
        ctx := context.WithValue(r.Context(), "IsBrowserRequest", ibr)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func getUserFromRequest(r *http.Request) *User {
    jwtCookie := r.Header.Get("X-Menshend-Token")
    user, err := FromJWT(jwtCookie)
    HttpCheckPanic(err, NotAuthorized)
    return user
}

//NeedLogin auth middleware, for router that need the jwt token
//TODO: header deletion test
func NeedLogin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := getUserFromRequest(r)
        ctx := context.WithValue(r.Context(), "User", user)
        //remove menshend token from headers
        r.Header.Del("X-Menshend-Token")
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func RoleHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        
    })
}

func GetServiceHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        role := r.Context().Value("role").(string)
        subdomain := r.Context().Value("subdomain").(string)
        user := r.Context().Value("user").(*User)
        vc, err := vault.NewClient(VaultConfig)
        HttpCheckPanic(err, InternalError)
        vc.SetToken(user.Menshend.VaultToken)
        serviceId := fmt.Sprintf("roles/%s/%s", role, subdomain)
        secret, err := vc.Logical().Read(fmt.Sprintf("%s/%s", Config.VaultPath, serviceId))
        HttpCheckPanic(err, InternalError)
        v1.CheckSecretFailIfIsNull(secret)
        as := &v1.AdminServiceResource{}
        mapstructure.Decode(secret.Data, as)
        if !as.IsActive {
            panic(NotAuthorized.Append("service " + as.ID + " is deactivated"))
        }
        ctx := context.WithValue(r.Context(), "service", as)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func ImpersonateWithinRoleHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        
    })
}

func ProxyHandlers(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := r.Context().Value("user").(*User)
        service := r.Context().Value("user").(*v1.AdminServiceResource)
        cr := &resolvers.CacheResolver{}
        backend := cr.Resolve(service, user)
        
        if (service.Strategy == "redirect") {
            redirect := &strategy.Redirect{}
            redirect.Execute(backend)(w, r)
            return
        } else if (service.Strategy == "proxy") {
            redirect := &strategy.Proxy{}
            redirect.Execute(backend)(w, r)
            return
        } else if (service.Strategy == "port-forward") {
            redirect := &strategy.PortForward{}
            redirect.Execute(backend)(w, r)
            return
        }
        panic(InternalError.Append("startegy for service " + service.ID + " was not recognized"))
    })
}
