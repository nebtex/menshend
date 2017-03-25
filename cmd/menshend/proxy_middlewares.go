package main

import (
    "net/http"
    "context"
    mconfig "github.com/nebtex/menshend/pkg/config"
    mutils "github.com/nebtex/menshend/pkg/utils"
    mfilters "github.com/nebtex/menshend/pkg/filters"
    vault "github.com/hashicorp/vault/api"
    "fmt"
    "github.com/nebtex/menshend/pkg/apis/menshend/v1"
    "github.com/mitchellh/mapstructure"
    "github.com/ansel1/merry"
    "github.com/Sirupsen/logrus"
    "strings"
)

func getSubDomain(s string) string {
    return strings.TrimSuffix(s, mconfig.Config.Host())
    
}

func GetSubDomainHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        subdomain := getSubDomain(r.Host)
        ctx := context.WithValue(r.Context(), mutils.Subdomain, subdomain)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

//TODO: add impersonate handler

//


//PanicHandler
func PanicHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        IsBrowserRequest := r.Context().Value(mutils.IsBrowserRequest).(bool)
        defer func() {
            rec := recover()
            if (rec == nil) {
                return
            }
            logrus.Errorln(rec)
            errorMessage := "Internal server error"
            errorCode := http.StatusInternalServerError
            switch x := rec.(type) {
            case merry.Error:
                logrus.Errorln(merry.Details(x))
                errorMessage = merry.UserMessage(x)
                errorCode = merry.HTTPCode(x)
            }
            if (!IsBrowserRequest) {
                http.Error(w, errorMessage, errorCode)
            } else {
                // Get a session.
                session, err := mconfig.FlashStore.Get(r, "flashes")
                if err != nil {
                    http.Error(w, err.Error(), http.StatusInternalServerError)
                    return
                }
                // Set a new flash.
                session.AddFlash(errorMessage)
                session.Save(r, w)
                subdomain := getSubDomain(r.Host)
                http.Redirect(w, r, mconfig.Config.Scheme() + "://" + mconfig.Config.Uris.Api + mconfig.Config.Host() + "/login?subdomain=" + subdomain, 302)
            }
        }()
        next.ServeHTTP(w, r)
    })
}


func GetTokenFromRequest(r *http.Request) string {
    bearerToken, _ := mfilters.ParseBearerAuth(r.Header.Get("Authorization"))
    vaultToken := r.Header.Get("X-Vault-Token")
    r.Header.Del("X-Vault-Token")
    
    if bearerToken != "" {
        if vaultToken == "" {
            vaultToken = bearerToken
            r.Header.Del("Authorization")
        }
    }
    return vaultToken
}


//NeedLogin auth middleware, for router that need the jwt token
func NeedLogin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        vaultToken := GetTokenFromRequest(r)
        ctx := context.WithValue(r.Context(), mutils.VaultToken, vaultToken)
        vc, err := vault.NewClient(vault.DefaultConfig())
        mutils.HttpCheckPanic(err, mutils.InternalError)
        vc.SetToken(vaultToken)
        tokenInfo, err := vc.Auth().Token().LookupSelf()
        mutils.HttpCheckPanic(err, mutils.NotAuthorized)
        ctx = context.WithValue(ctx, mutils.TokenInfo, tokenInfo)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

//RoleHandler pick a role
func RoleHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        mdRole := r.Header.Get("md-role")
        mdRoleQ := r.FormValue("md-role")
        if r.PostForm.Get("md-role") != "" {
            mdRoleQ = r.PostForm.Get("md-role")
        }
        if mdRoleQ != "" {
            mdRole = mdRoleQ
        }
        IsBrowserRequest := r.Context().Value(mutils.IsBrowserRequest).(bool)
        if !IsBrowserRequest {
            if mdRole != "" {
                ctx := context.WithValue(r.Context(), mutils.Role, mdRole)
                next.ServeHTTP(w, r.WithContext(ctx))
                return
            }
            
            ctx := context.WithValue(r.Context(), mutils.Role, mconfig.Config.DefaultRole)
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
            ctx := context.WithValue(r.Context(), mutils.Role, ck.Value)
            next.ServeHTTP(w, r.WithContext(ctx))
            return
        }
        ctx := context.WithValue(r.Context(), mutils.Role, mconfig.Config.DefaultRole)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

//GetServiceHandler read the service from vault
func GetServiceHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        role := r.Context().Value(mutils.Role).(string)
        subdomain := r.Context().Value(mutils.Subdomain).(string)
        vaultToken := r.Context().Value(mutils.VaultToken).(string)
        vc, err := vault.NewClient(vault.DefaultConfig())
        mutils.HttpCheckPanic(err, mutils.InternalError)
        vc.SetToken(vaultToken)
        serviceId := fmt.Sprintf("roles/%s/%s", role, subdomain)
        secret, err := vc.Logical().Read(fmt.Sprintf("%s/%s", mconfig.Config.VaultPath, serviceId))
        mutils.HttpCheckPanic(err, mutils.InternalError)
        v1.CheckSecretFailIfIsNull(secret)
        as := &v1.AdminServiceResource{}
        mapstructure.Decode(secret.Data, as)
        if !as.Active() {
            panic(mutils.NotAuthorized.WithUserMessage("service " + as.Meta.ID + " is deactivated"))
        }
        ctx := context.WithValue(r.Context(), mutils.Service, as)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
/*
//ImpersonateWithinRoleHandler
func ImpersonateWithinRoleHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        service := r.Context().Value(v1.Service).(*v1.AdminServiceResource)
        tokenInfo := r.Context().Value(v1.TokenInfo).(*vault.Secret)
        
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
}*/
/*
func ProxyHandlers() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var CSRF func(http.Handler) http.Handler
        IsBrowserRequest := r.Context().Value(vI.sBrowserRequest).(bool)
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
*/
