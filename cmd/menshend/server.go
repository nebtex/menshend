package main

import (
    "net/http"
    "github.com/gorilla/mux"
    "github.com/nebtex/menshend/pkg/apis/menshend/v1"
)
//TODO: prefiling check in server mode check if vault token is present

func mainHandler(response http.ResponseWriter, request *http.Request) {
    //detect  menshend host
    //use proxy server
}

func proxyServer(handler http.Handler) http.Handler {
    return PanicHandler(GetSubDomainHandler(v1.BrowserDetectorHandler(
        NeedLogin(RoleHandler(GetServiceHandler(ProxyHandler()))))))
}


func uilogin(){
    
}

func menshendServer() http.Handler {
    // /ui
    // /uilogin
    // /uilogout
    // /v1 - api
    r := mux.NewRouter()
    //r.Host("{subdomain:[a-z\\-]+}." + Config.HostWithoutPort()).Handler(handler)
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


