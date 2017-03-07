package strategy

import (
    "github.com/nebtex/menshend/pkg/backend"
    "net/http"
    "github.com/vulcand/oxy/forward"
    "github.com/nebtex/menshend/pkg/utils"
    "net/url"
    . "github.com/nebtex/menshend/pkg/utils"
)

var Fwd *forward.Forwarder

type errorHandler struct {
    
}

func (*errorHandler)ServeHTTP(w http.ResponseWriter, req *http.Request, err error) {
    panic(BadGateway.Append("backend is not responding"))
}

func init() {
    var err error
    Fwd, err = forward.New(forward.ErrorHandler(&errorHandler{}))
    menshend.CheckPanic(err)
}

type Proxy struct {
}

//ProxyHandler forward request to the backend services
func (r *Proxy) Execute(b backend.Backend) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        for key, value := range b.Headers() {
            r.Header.Set(key, value)
        }
        bUrl, err := url.Parse(b.BaseUrl())
        HttpCheckPanic(err, InternalError)
        r.URL.Host = bUrl.Host
        r.URL.User = bUrl.User
        r.URL.Scheme = bUrl.Scheme
        Fwd.ServeHTTP(w, r)
    }
}
