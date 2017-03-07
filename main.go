package menshend

import (
    "github.com/gorilla/mux"
    "net/http"
    "github.com/gorilla/csrf"
    "github.com/gorilla/context"
    "fmt"
    "time"
    "net/http/httptest"
)

//CSRFHeader ...
func CSRFHeader(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-CSRF-Token", csrf.Token(r))
        next.ServeHTTP(w, r)
    })
}


func Server() {
    r := mux.NewRouter()
    fmt.Println(Config.Host)
    
    //menshend app
    menshendRouter := r.Host(Config.HostWithoutPort()).Subrouter()
    
    //login api
    menshendRouter.Path("/v1/api/auth/login/token").HandlerFunc(TokenLoginHandler).Methods("POST")
    menshendRouter.Path("/v1/api/auth/login/userpass/{username}").HandlerFunc(UserPasswordHandler).Methods("POST")
    menshendRouter.Path("/v1/api/auth/login/github").HandlerFunc(UserPasswordHandler).Methods("POST")
    menshendRouter.Path("/v1/api/auth/menshend/token").HandlerFunc(UserPasswordHandler).Methods("DELETE")
    
    //admin api
    // this endpoints provide the CRUD operation to manage the services and roles
    menshendRouter.Path("/v1/api/admin/role/{role}").Handler(NeedLogin(NeedAdmin(http.HandlerFunc(DeleteServiceHandler)))).Methods("DELETE")
    menshendRouter.Path("/v1/api/admin/role/{role}/service/{subDomain}").Handler(NeedLogin(NeedAdmin(http.HandlerFunc(CreateEditServiceHandler)))).Methods("PUT")
    menshendRouter.Path("/v1/api/admin/role/{role}/service/{subDomain}").Handler(NeedLogin(NeedAdmin(http.HandlerFunc(DeleteServiceHandler)))).Methods("DELETE")
    menshendRouter.Path("/v1/api/admin/role/{role}/service/{subDomain}").Handler(NeedLogin(NeedAdmin(http.HandlerFunc(GetServiceHandler)))).Methods("GET")
    
    //client api
    //provide basic features that allow to any user list the services in the ui
    menshendRouter.Path("/v1/api/client/services").Handler(NeedLogin(http.HandlerFunc(ServiceListHandler)))
    menshendRouter.Path("/v1/api/client/status").Handler(http.HandlerFunc(LoginStatusHandler))
    
    //secrets api
    menshendRouter.Path("/v1/api/secrets/{app}/read/{secretPath}").Handler(http.HandlerFunc(LoginStatusHandler)).Methods("GET")
    
    //impersonate api
    //impersonate other users
    menshendRouter.Path("/v1/api/impersonate").Handler(NeedLogin(http.HandlerFunc(ImpersonateHandler))).Methods("POST")
    
    //space api
    //get info about the space
    menshendRouter.Path("/v1/api/space").Handler(NeedLogin(http.HandlerFunc(ImpersonateHandler)))
    
    
    //ui
    menshendRouter.Path("/ui/login/*").Handler(NeedLogin(http.HandlerFunc(ImpersonateHandler)))
    menshendRouter.Path("/ui/v1/*").Handler(NeedLogin(http.HandlerFunc(ImpersonateHandler)))
    
    //flash message  api
    menshendRouter.Path("/ui/messages/flash").HandlerFunc(UserPasswordHandler)
    //static
    menshendRouter.Path("/ui/static").Handler(NeedLogin(http.HandlerFunc(ImpersonateHandler)))
    
    //menshendServerHandlers(r.Host("menshend." + BaseDomain).Subrouter())
    //ProxyHandlers(r.Host("{subdomain:.+}." + BaseDomain).Subrouter())
    
    secure := true
    if Config.Scheme == "http" {
        secure = false
    }
    CSRF := csrf.Protect([]byte(Config.HashKey), csrf.Secure(secure), csrf.Domain(Config.HostWithoutPort()), csrf.Path("/"))
    http.ListenAndServe(fmt.Sprintf(":%d", Config.ListenPort), context.ClearHandler(PanicHandler(CSRF(CSRFHeader(r)))))
    
}

func main() {
    Server()
}
//setToken ..
// expiresIn in milliseconds
func setToken(u *User, expiresIn int64, w http.ResponseWriter) {
    expireAt := MakeTimestampMillisecond()
    if expiresIn == 0 {
        expireAt += Config.DefaultTTL
    } else {
        expireAt += expiresIn
    }
    u.SetExpiresAt(expireAt)
    ct := &http.Cookie{Path: "/", Name: "X-Menshend-Token", Value: u.GenerateJWT(),
        Expires: time.Unix(u.ExpiresAt / 1000, 0),
        HttpOnly:true }
    
    ct.Domain = "." + Config.HostWithoutPort()
    
    if Config.Scheme == "https" {
        ct.Secure = true
    }
    http.SetCookie(w, ct)
}
//TestSetToken
func TestSetToken(t *testing.T) {
    
    Convey("TestSetToken",
        t, func(c C) {
            Convey("Cookie should has the same expiration date that the token",
                func() {
                    Config.Scheme = "http"
                    u, err := NewUser("test-acl")
                    So(err, ShouldBeNil)
                    u.GitHubLogin("criloz", "delos", "umbrella")
                    w := &httptest.ResponseRecorder{}
                    expTime := MakeTimestampMillisecond() + 3600 * 1000
                    setToken(u, 3600 * 1000, w)
                    r := w.Result()
                    c := r.Cookies()[0]
                    So(c.Value, ShouldEqual, u.GenerateJWT())
                    So(c.HttpOnly, ShouldEqual, true)
                    So(c.Expires.Unix(), ShouldEqual, expTime/1000)
                    So(c.Secure, ShouldEqual, false)
                })
            Convey("Cookie should be secure if kuber is behind an https proxy",
                func() {
                    Config.Scheme = "https"
                    u, err := NewUser("test-acl")
                    u.SetExpiresAt(GetNow() + 3600 * 1000)
                    So(err, ShouldBeNil)
                    u.GitHubLogin("criloz", "delos", "umbrella")
                    w := &httptest.ResponseRecorder{}
                    expTime := MakeTimestampMillisecond() + 3600 * 1000
                    setToken(u, 3600 * 1000, w)
                    r := w.Result()
                    c := r.Cookies()[0]
                    So(c.Value, ShouldEqual, u.GenerateJWT())
                    So(c.HttpOnly, ShouldEqual, true)
                    So(c.Expires.Unix(), ShouldEqual, expTime/1000)
                    So(c.Secure, ShouldEqual, true)
                })
        })
}
