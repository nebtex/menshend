package main

import (
    . "github.com/nebtex/menshend/pkg/config"
    "net/http"
    "strings"
    "github.com/gorilla/mux"
)

func getSubDomain(s string) string {
    return strings.TrimSuffix(s, Config.Host())
    
}


func mainHandler(response http.ResponseWriter, request *http.Request) {
   
    
}

func proxyServer() http.Handler {
    r := mux.NewRouter()
    handler := TokenRealmSecurityHandler(RoleHandler(GetServiceHandler(ImpersonateWithinRoleHandler(ProxyHandlers()))))
    handler = GetSubDomainHandler(DetectBrowser(PanicHandler(NeedLogin(handler))))
    r.Host("{subdomain:[a-z\\-]+}." + Config.HostWithoutPort()).Handler(handler)
    return r
}

func uiServer() http.Handler {
    r := mux.NewRouter()
    handler := TokenRealmSecurityHandler(RoleHandler(GetServiceHandler(ImpersonateWithinRoleHandler(ProxyHandlers()))))
    handler = GetSubDomainHandler(DetectBrowser(PanicHandler(NeedLogin(handler))))
    r.Host("{subdomain:[a-z\\-]+}." + Config.HostWithoutPort()).Handler(handler)
    return r
}

/*
func setToken(u *User, expiresIn int64, response *restful.Response, hasCSRF bool) {
    expireAt := MakeTimestampMillisecond()
    if expiresIn == 0 {
        expireAt += Config.DefaultTTL
    } else {
        expireAt += expiresIn
    }
    u.SetExpiresAt(expireAt)
    
    if !hasCSRF {
        response.AddHeader("X-Menshend-Token", u.GenerateJWT())
        
    } else {
        ct := &http.Cookie{Path: "/", Name: "X-Menshend-Token", Value: u.GenerateJWT(),
            Expires: time.Unix(u.ExpiresAt / 1000, 0),
            HttpOnly:true }
        
        ct.Domain = "." + Config.HostWithoutPort()
        
        if Config.Scheme == "https" {
            ct.Secure = true
        }
        http.SetCookie(response.ResponseWriter, ct)
        
    }
}
func setExpirationTime(u *User, expiresIn int64) *User {
    expireAt := MakeTimestampMillisecond()
    if expiresIn == 0 {
        expireAt += Config.DefaultTTL
    } else {
        expireAt += expiresIn
    }
    u.SetExpiresAt(expireAt)
    return u
    
}
*/


//TODO: prefiling check in server mode check if vault token is present
