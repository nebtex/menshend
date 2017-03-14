package main

import (
  //  "github.com/gorilla/mux"
  //  "fmt"
    . "github.com/nebtex/menshend/pkg/config"
    "net/http"
    "strings"
)

func getSubDomain(s string) string {
    return strings.TrimSuffix(s, Config.Host)
    
}
//CorsHandler()

func proxy(response http.ResponseWriter, request *http.Request) {
    //login [delete menshend headers] // get if user come from browser or api
    //get role
    //get service
    //check if service is active
    //impersonate within role
    //proxy[http: check cors]
}

func mainHandler(response http.ResponseWriter, request *http.Request) {
    //get subdomain
    subdomain := getSubDomain(request.Host)
    if subdomain == Config.MenshendSubdomain {
        //add api router
        //set cors
        // add flash endpoin
        // proxy impersonate
        //proxy login
        
    } else {
        proxy(response, request)
    }
    
}

func serve() {
   /* r := mux.NewRouter()
    
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
    //ProxyHandlers(r.Host("{subdomain:.+}." + BaseDomain).Subrouter())*/
    
}


