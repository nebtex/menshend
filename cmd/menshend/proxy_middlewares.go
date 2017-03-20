package main

import (
    "net/http"
    "context"
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
    "github.com/gorilla/csrf"
    "github.com/rs/cors"
)

func GetSubDomainHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        subdomain := getSubDomain(r.Host)
        ctx := context.WithValue(r.Context(), "subdomain", subdomain)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
//TODO: add impersonate handler
//TODO: portfoward from broser should return error
//TODO: add role handler - {map menshend-role, vault-role}.
//TODO: panic handler when ui is not active should print error
//TODO: read browser header
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
            ck, err := r.Cookie("X-Vault-Token")
            if err == nil {
                ibr = true
                r.Header.Add("X-Vault-Token", ck.Value)
                // remove menshend cookie
                cks := r.Cookies()
                r.Header.Del("Cookie")
                for _, c := range cks {
                    if c.Name == "X-Vault-Token" {
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

//put in the pat service that the user is triying to access
//PanicHandler
func PanicHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        IsBrowserRequest := r.Context().Value("IsBrowserRequest").(bool)
        var errorMessage string
        var errorCode int
        
        defer func() {
            rec := recover()
            if (rec == nil) {
                return
            }
            switch x := rec.(type) {
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
            
            if (!IsBrowserRequest) || (!Config.EnableUI) {
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
                http.Redirect(w, r, Config.Scheme() + "://" + Config.MenshendSubdomain + Config.Host() + "/ui/login", 302)
            }
        }()
        
        next.ServeHTTP(w, r)
    })
}

func getUserFromRequest(r *http.Request) *User {
    jwtCookie := r.Header.Get("X-Vault-Token")
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
        r.Header.Del("X-Vault-Token")
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func RoleHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        mdRole := r.Header.Get("md-role")
        mdRoleQ := r.FormValue("md-role")
        if mdRoleQ != "" {
            mdRole = mdRoleQ
        }
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
            panic(NotAuthorized.WithUserMessage("service " + as.ID + " is deactivated"))
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
        next.ServeHTTP(w, r)
    })
}

func NextCSRFHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ct := csrf.Token(r)
        r.Header.Set("X-Next-CSRF-Token", ct)
        w.Header().Set("X-Next-CSRF-Token", ct)
        next.ServeHTTP(w, r)
    })
}

func ProxyHandlers() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var CSRF func(http.Handler) http.Handler
        IsBrowserRequest := r.Context().Value("IsBrowserRequest").(bool)
        user := r.Context().Value("User").(*User)
        service := r.Context().Value("service").(*v1.AdminServiceResource)
        cr := &resolvers.CacheResolver{}
        backend := cr.Resolve(service, user)
        //        var co cors.Options
        var handler http.Handler
        switch  service.Strategy {
        case "redirect":
            handler = (&strategy.Redirect{}).Execute(backend)
            handler.ServeHTTP(w, r)
            return
        
        case "proxy":
            handler = (&strategy.Proxy{}).Execute(backend)
            if IsBrowserRequest {
                if service.CSRF {
                    if Config.Scheme() == "http" {
                        CSRF = csrf.Protect([]byte(Config.BlockKey), csrf.Secure(false), csrf.Domain(service.SubDomain + Config.HostWithoutPort()))
                    }
                    CSRF = csrf.Protect([]byte(Config.BlockKey), csrf.Domain(service.SubDomain + Config.HostWithoutPort()))
                    handler = CSRF(NextCSRFHandler(handler))
                }
                if service.EnableCustomCors {
                    
                    co := cors.Options{
                        AllowedOrigins:service.Cors.AllowedOrigins,
                        AllowedMethods:service.Cors.AllowedMethods,
                        AllowedHeaders:service.Cors.AllowedHeaders,
                        ExposedHeaders:service.Cors.ExposedHeaders,
                        AllowCredentials:service.Cors.AllowCredentials,
                        MaxAge:service.Cors.MaxAge,
                        OptionsPassthrough:service.Cors.OptionsPassthrough,
                        Debug:service.Cors.Debug,
                    }
                    crs := cors.New(co)
                    handler = crs.Handler(handler)
                }
            }
            handler.ServeHTTP(w, r)
            return
        case "port-forward":
            handler = (&strategy.PortForward{}).Execute(backend)
            handler.ServeHTTP(w, r)
            return
        default:
            panic(InternalError.WithUserMessage("strategy for service " + service.ID + " was not recognized"))
        }
        
    })
}
