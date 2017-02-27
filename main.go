package kuper

import (
    "github.com/gorilla/mux"
    "net/http"
    "github.com/gorilla/csrf"
    "github.com/gorilla/context"
    
    "fmt"
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
    //kuper app
    kuperRouter := r.Host(Config.Host).Subrouter()
    
    //login api
    kuperRouter.Path("/login/oauth/{provider}").HandlerFunc(OauthLoginHandler)
    kuperRouter.Path("/login/oauth/{provider}/callback").HandlerFunc(OauthLoginCallback)
    kuperRouter.Path("/login/token").HandlerFunc(TokenLoginHandler).Methods("POST")
    kuperRouter.Path("/login/userpass").HandlerFunc(UserPasswordHandler).Methods("POST")
    
    //admin api
    // this endpoints provide the CRUD operation to manage the services
    kuperRouter.Path("/v1/api/admin/service/save").Handler(NeedLogin(NeedAdmin(http.HandlerFunc(CreateEditServiceHandler)))).Methods("POST")
    kuperRouter.Path("/v1/api/admin/service/delete").Handler(NeedLogin(NeedAdmin(http.HandlerFunc(DeleteServiceHandler)))).Methods("POST")
    kuperRouter.Path("/v1/api/admin/service/{subDomain}").Handler(NeedLogin(NeedAdmin(http.HandlerFunc(DeleteServiceHandler))))
    
    //client api
    //provide basic features that allow to any user list the services in the ui
    kuperRouter.Path("/v1/api/client/service/list").Handler(NeedLogin(http.HandlerFunc(ServiceListHandler)))
    kuperRouter.Path("/v1/api/client/status").Handler(http.HandlerFunc(LoginStatusHandler))
    
    //flash message  api
    kuperRouter.Path("/messages/flash").HandlerFunc(UserPasswordHandler)
    
    //impersonate api
    //impersonate other users
    kuperRouter.Path("/v1/api/impersonate").Handler(NeedLogin(http.HandlerFunc(ImpersonateHandler))).Methods("POST")
    
    //space api
    //get info about the space
    kuperRouter.Path("/v1/api/space").Handler(NeedLogin(http.HandlerFunc(ImpersonateHandler)))
    
    //static
    kuperRouter.Path("/static").Handler(NeedLogin(http.HandlerFunc(ImpersonateHandler)))
    
    //KuperServerHandlers(r.Host("kuper." + BaseDomain).Subrouter())
    //ProxyHandlers(r.Host("{subdomain:.+}." + BaseDomain).Subrouter())
    
    secure := true
    if Config.Scheme == "http" {
        secure = false
    }
    CSRF := csrf.Protect([]byte(Config.Salt), csrf.Secure(secure), csrf.Domain(Config.Host), csrf.Path("/"))
    http.ListenAndServe(fmt.Sprintf(":%d", Config.ListenPort), context.ClearHandler(PanicHandler(CSRF(CSRFHeader(r)))))
    
}

func main() {
    Server()
}
