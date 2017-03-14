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
    "github.com/ansel1/merry"
    "github.com/Sirupsen/logrus"
)

func GetSubDomainHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        subdomain := getSubDomain(r.Host)
        ctx := context.WithValue(r.Context(), "subdomain", subdomain)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

//TokenRealmSecurity don't allow api token to be used in the browser as cookies or headers
func TokenRealmSecurityHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        IsBrowserRequest := r.Context().Value("IsBrowserRequest").(bool)
        user := r.Context().Value("User").(*User)
        if (IsBrowserRequest) {
            if (user.Menshend.Realm != BrowserRealm) {
                panic(NotAuthorized)
            }
        }
        next.ServeHTTP(w, r)
    })
}

func DetectBrowser(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ibr := false
        if len(r.Cookies()) > 0 {
            ck, err := r.Cookie("X-Menshend-Token")
            if err == nil {
                ibr = true
                r.Header.Add("X-Menshend-Token", ck.Value)
                // remove menshend cookie
                cks := r.Cookies()
                r.Header.Del("Cookie")
                for _, c := range cks {
                    if c.Name == "X-Menshend-Token" {
                        continue
                    }
                    r.AddCookie(c)
                }
            }
        }
        ctx := context.WithValue(r.Context(), "IsBrowserRequest", ibr)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func PanicHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        IsBrowserRequest := r.Context().Value("IsBrowserRequest").(bool)
        var errorMessage string
        var errorCode int
        
        defer func() {
            r := recover()
            if (r == nil) {
                return
            }
            switch x := r.(type) {
            case merry.Error:
                logrus.Errorln(merry.Details(x))
                errorMessage = x.Error()
                errorCode = merry.HTTPCode(x)
            case error:
                logrus.Errorln(x)
                errorMessage = "Internal server error"
                errorCode = http.StatusInternalServerError
            default:
                errorMessage = "Uknown error"
                errorCode = http.StatusInternalServerError
                
            }
            
            if !IsBrowserRequest {
                http.Error(w, errorMessage, errorCode)
                
            } else {
                // Get a session.
                session, err := FlashStore.Get(r, "flashes")
                if err != nil {
                    http.Error(w, err.Error(), http.StatusInternalServerError)
                    return
                }
                // Set a new flash.
                session.AddFlash(errorMessage)
                session.Save(r, w)
                http.Redirect(w, r, Config.MenshendSubdomain + Config.Host + "/ui", 302)
            }
        }()
        
        next.ServeHTTP(w, r)
    })
}

func getUserFromRequest(r *http.Request) *User {
    jwtCookie := r.Header.Get("X-Menshend-Token")
    user, err := FromJWT(jwtCookie)
    HttpCheckPanic(err, NotAuthorized)
    return user
}

//NeedLogin auth middleware, for router that need the jwt token
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
        mdRole := r.FormValue("md-role")
        IsBrowserRequest := r.Context().Value("IsBrowserRequest").(bool)
        
        if !IsBrowserRequest {
            
            if mdRole != "" {
                ctx := context.WithValue(r.Context(), "role", mdRole)
                next.ServeHTTP(w, r.WithContext(ctx))
                return
            }
            
            ctx := context.WithValue(r.Context(), "role", Config.DefaultRole)
            next.ServeHTTP(w, r.WithContext(ctx))
            return
        }
        
        if mdRole != "" {
            ct := &http.Cookie{
                Path: "/",
                Name: "md-role",
                Value: mdRole,
                HttpOnly:true,
                Domain: r.Host }
            
            http.SetCookie(w, ct)
            q := r.URL.Query()
            q.Del("md-role")
            r.URL.RawQuery = q.Encode()
            http.Redirect(w, r, r.URL.String(), 302)
            return
        }
        ck, err := r.Cookie("md-role")
        if err == nil {
            ctx := context.WithValue(r.Context(), "role", ck.Value)
            next.ServeHTTP(w, r.WithContext(ctx))
            return
        }
        ctx := context.WithValue(r.Context(), "role", Config.DefaultRole)
        next.ServeHTTP(w, r.WithContext(ctx))
        
    })
}

func GetServiceHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        role := r.Context().Value("role").(string)
        subdomain := r.Context().Value("subdomain").(string)
        user := r.Context().Value("User").(*User)
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
        service := r.Context().Value("service").(*v1.AdminServiceResource)
        user := r.Context().Value("User").(*User)
        
        if service.ImpersonateWithinRole {
            if r.URL.Query().Get("md-user") != "" {
                user.Menshend.Username = r.URL.Query().Get("md-user")
            }
            
            if len(r.URL.Query()["md-groups"]) > 0 {
                user.Menshend.Groups = r.URL.Query()["md-groups"]
            }
        }
    })
}

func ProxyHandlers(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := r.Context().Value("User").(*User)
        service := r.Context().Value("service").(*v1.AdminServiceResource)
        cr := &resolvers.CacheResolver{}
        backend := cr.Resolve(service, user)
        
        var handler http.Handler
        switch  service.Strategy {
        case "redirect":
            handler = CSRF(&strategy.Redirect{}.Execute(backend))
        case "proxy":
            handler = &strategy.Proxy{}.Execute(backend)
        case "port-forward":
            handler = &strategy.PortForward{}.Execute(backend)
        default:
            panic(InternalError.Append("strategy for service " + service.ID + " was not recognized"))
        }
        handler.ServeHTTP(w, r)
        
    })
}
