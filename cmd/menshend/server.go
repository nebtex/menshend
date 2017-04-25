package main

import (
    "fmt"
    "net/http"
    
    "github.com/Sirupsen/logrus"
    "github.com/nebtex/menshend/pkg/apis/menshend/v1"
    "github.com/rakyll/statik/fs"
    mconfig "github.com/nebtex/menshend/pkg/config"
    
    "github.com/gorilla/mux"
    . "github.com/nebtex/menshend/statik"
    "strings"
)

func proxyServer() http.Handler {
    return PanicHandler(GetSubDomainHandler(v1.BrowserDetectorHandler(
        NeedLogin(RoleHandler(GetServiceHandler(ProxyHandler()))))))
}

func react() http.Handler {
    r := mux.NewRouter()
    statikFS, _ := fs.New()
    r.PathPrefix("/static").Handler(http.FileServer(statikFS))
    r.PathPrefix("/").Handler(Index())
    return r
}

func menshendServer(address, port string) error {
    CSRF := getUiCSRF()
    
    http.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
        if strings.HasSuffix(request.Host, mconfig.Config.Host()) {
            subdomain := getSubDomain(request.Host)
            if subdomain == mconfig.Config.Uris.MenshendSubdomain {
                subrouter := mux.NewRouter()
                subrouter.PathPrefix("/ui").Handler(uiHandler())
                subrouter.PathPrefix("/v1").Handler(v1.APIHandler())
                subrouter.PathPrefix("/").Handler(CSRF(react()))
                subrouter.ServeHTTP(response, request)
                return
            }
            proxyServer().ServeHTTP(response, request)
            return
        }
        
        health := mux.NewRouter()
        health.PathPrefix("/status").HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
            response.WriteHeader(200)
            response.Write([]byte("OK"))
        })
        health.ServeHTTP(response, request)
        
    })
    
    logrus.Infof("Server listing on %s:%s", address, port)
    return http.ListenAndServe(fmt.Sprintf("%s:%s", address, port), nil)
    
}


