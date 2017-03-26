package strategy

import (
    "github.com/nebtex/menshend/pkg/resolvers"
    "net/http"
    "github.com/vulcand/oxy/forward"
    "net/url"
    mutils "github.com/nebtex/menshend/pkg/utils"
    mconfig "github.com/nebtex/menshend/pkg/config"
    vault "github.com/hashicorp/vault/api"
    "io/ioutil"
    "github.com/gorilla/csrf"
    "github.com/rs/cors"
    "github.com/Sirupsen/logrus"
)

type errorHandler struct {
    
}
// Options is a configuration container to setup the CORS middleware.
type CorsOptions struct {
    // AllowedOrigins is a list of origins a cross-domain request can be executed from.
    // If the special "*" value is present in the list, all origins will be allowed.
    // An origin may contain a wildcard (*) to replace 0 or more characters
    // (i.e.: http://*.domain.com). Usage of wildcards implies a small performance penality.
    // Only one wildcard can be used per origin.
    // Default value is ["*"]
    AllowedOrigins     []string
    // AllowedMethods is a list of me thods the client is allowed to use with
    // cross-domain requests. Default value is simple methods (GET and POST)
    AllowedMethods     []string
    // AllowedHeaders is list of non simple headers the client is allowed to use with
    // cross-domain requests.
    // If the special "*" value is present in the list, all headers will be allowed.
    // Default value is [] but "Origin" is always appended to the list.
    AllowedHeaders     []string
    // ExposedHeaders indicates which headers are safe to expose to the API of a CORS
    // API specification
    ExposedHeaders     []string
    // AllowCredentials indicates whether the request can include user credentials like
    // cookies, HTTP authentication or client side SSL certificates.
    AllowCredentials   bool
    // MaxAge indicates how long (in seconds) the results of a preflight request
    // can be cached
    MaxAge             int
    // OptionsPassthrough instructs preflight to let other potential next handlers to
    // process the OPTIONS method. Turn this on if your application handles OPTIONS.
    OptionsPassthrough bool
    // Debugging flag adds additional output to debug server side CORS issues
    Debug              bool
}

func (*errorHandler)ServeHTTP(w http.ResponseWriter, req *http.Request, err error) {
    logrus.Error(err.Error())
    panic(mutils.BadGateway.WithUserMessage("backend is not responding"))
}

type Proxy struct {
    Cors *CorsOptions `json:"cors"`
    CSRF bool `json:"csrf"`
}

func NextCSRFHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ct := csrf.Token(r)
        r.Header.Set("X-Next-CSRF-Token", ct)
        w.Header().Set("X-Next-CSRF-Token", ct)
        next.ServeHTTP(w, r)
    })
}

//ProxyHandler forward request to the backend services
func (ps *Proxy) Execute(rs resolvers.Resolver, tokenInfo *vault.Secret) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var CSRF func(http.Handler) http.Handler
        var handler http.Handler
        IsBrowserRequest := r.Context().Value(mutils.IsBrowserRequest).(bool)
        subdomain := r.Context().Value(mutils.Subdomain).(string)
        Fwd, err := forward.New(forward.ErrorHandler(&errorHandler{}))
        mutils.HttpCheckPanic(err, mutils.InternalError)
        
        if rs.NeedBody() {
            data, err := ioutil.ReadAll(r.Body)
            mutils.HttpCheckPanic(err, mutils.InternalError)
            rs.SetRequest(r.Method, string(data))
        } else {
            rs.SetRequest(r.Method, "")
        }
        
        b := rs.Resolve(tokenInfo)
        
        if !b.Passed() {
            panic(mutils.NotAuthorized.WithUserMessage(b.Error().Error()))
        }
        
        for key, value := range b.Headers() {
            r.Header.Set(key, value)
        }
        bUrl, err := url.Parse(b.BaseUrl())
        mutils.HttpCheckPanic(err, mutils.InternalError)
        r.URL.Host = bUrl.Host
        r.URL.User = bUrl.User
        r.URL.Scheme = bUrl.Scheme
        handler = Fwd
        if IsBrowserRequest {
            if ps.CSRF {
                if mconfig.Config.Scheme() == "http" {
                    CSRF = csrf.Protect([]byte(mconfig.Config.BlockKey), csrf.Secure(false), csrf.Domain(subdomain + mconfig.Config.HostWithoutPort()))
                }
                CSRF = csrf.Protect([]byte(mconfig.Config.BlockKey), csrf.Domain(subdomain + mconfig.Config.HostWithoutPort()))
                handler = CSRF(NextCSRFHandler(handler))
            }
            if ps.Cors != nil {
                co := cors.Options{
                    AllowedOrigins:ps.Cors.AllowedOrigins,
                    AllowedMethods:ps.Cors.AllowedMethods,
                    AllowedHeaders:ps.Cors.AllowedHeaders,
                    ExposedHeaders:ps.Cors.ExposedHeaders,
                    AllowCredentials:ps.Cors.AllowCredentials,
                    MaxAge:ps.Cors.MaxAge,
                    OptionsPassthrough:ps.Cors.OptionsPassthrough,
                    Debug:ps.Cors.Debug,
                }
                crs := cors.New(co)
                handler = crs.Handler(handler)
            }
        }
        handler.ServeHTTP(w, r)
    })
}
