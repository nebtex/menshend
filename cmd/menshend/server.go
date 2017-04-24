package main

import (
    "fmt"
    "net/http"
    
    "github.com/Sirupsen/logrus"
    "github.com/nebtex/menshend/pkg/apis/menshend/v1"
    "github.com/rakyll/statik/fs"
    
    "github.com/gorilla/mux"
    . "github.com/nebtex/menshend/statik"
)

func mainHandler(response http.ResponseWriter, request *http.Request) {
    //detect menshend host
    //use proxy server
}

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
    // /v1 - api
    //http.Handle("/api", v1.APIHandler())
    r := mux.NewRouter()
    r.PathPrefix("/ui").Handler(uiHandler())
    http.Handle("/", r)
    logrus.Infof("Server listing on %s:%s", address, port)
    return http.ListenAndServe(fmt.Sprintf("%s:%s", address, port), nil)
    
    //r.Host("{subdomain:[a-z\\-]+}." + Config.HostWithoutPort()).Handler(handler)
}


