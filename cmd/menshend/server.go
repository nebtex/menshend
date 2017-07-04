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
    "regexp"
    "github.com/vulcand/oxy/forward"
    "github.com/vulcand/oxy/utils"
)

func proxyServer() http.Handler {
    return v1.BrowserDetectorHandler(PanicHandler(GetSubDomainHandler(
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
        if proto := request.Header.Get(forward.XForwardedProto); proto != "" {
            if proto != mconfig.Config.Scheme() {
                newUrl := utils.CopyURL(request.URL)
                newUrl.Scheme = mconfig.Config.Scheme()
                newUrl.Host = request.Host
                logrus.Info("redirecting to: ", newUrl.String())
                http.Redirect(response, request, newUrl.String(), http.StatusTemporaryRedirect)
                return
            }
        }

        var re = regexp.MustCompile(`(.+\.)?` + mconfig.Config.HostWithoutPort() + `(:[0-9]+)?`)
        var str = request.Host
        all := re.FindAllStringSubmatch(str, -1)

        if len(all) > 0 {
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

        subrouter2 := mux.NewRouter()
        subrouter2.PathPrefix("/status").HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
            response.WriteHeader(200)
            response.Write([]byte("OK"))
        })
        subrouter2.PathPrefix("/ui").Handler(uiHandler())
        subrouter2.PathPrefix("/v1").Handler(v1.APIHandler())
        subrouter2.PathPrefix("/").Handler(CSRF(react()))
        subrouter2.ServeHTTP(response, request)
    })

    logrus.Infof("Server listing on %s:%s", address, port)
    return http.ListenAndServe(fmt.Sprintf("%s:%s", address, port), nil)

}
