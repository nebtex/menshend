package v1

import (
    "net/http"
    "github.com/emicklei/go-restful"
    "github.com/ansel1/merry"
    "github.com/Sirupsen/logrus"
    "fmt"
    "context"
    "github.com/gorilla/csrf"
    "github.com/nebtex/menshend/pkg/config"
)


//APIHandler menshend api endpoint handler
func APIHandler() http.Handler {
    wsContainer := restful.NewContainer()
    account := &AuthResource{}
    account.Register(wsContainer)
    admin := &AdminServiceResource{}
    admin.Register(wsContainer)
    client := &ClientServiceResource{}
    client.Register(wsContainer)
    secret := SecretResource{}
    secret.Register(wsContainer)
    space := SpaceResource{}
    space.Register(wsContainer)
    flash := FlashResource{}
    flash.Register(wsContainer)
    wsContainer.RecoverHandler(APIPanicHandler)
    wsContainer.DoNotRecover(false)
    return BrowserDetectorHandler(APICSRFHandler(wsContainer))
}


//APIPanicHandler handle any panic in the api endpoint
func APIPanicHandler(rec interface{}, w http.ResponseWriter) {
    
    logrus.Errorln(rec)
    errorMessage := "Internal server error"
    errorCode := http.StatusInternalServerError
    
    switch x := rec.(type) {
    case merry.Error:
        errorMessage = merry.UserMessage(x)
        errorCode = merry.HTTPCode(x)
    }
    
    w.WriteHeader(errorCode)
    
    _, err := w.Write([]byte(fmt.Sprintf(`{"message": "%s"}`, errorMessage)))
    if (err != nil) { logrus.Error(err)}
    w.Header().Set("Content-Type", "application/json")
}

//RequestType ..
type RequestType int

//IsBrowserRequest ...
const IsBrowserRequest RequestType = 0

//BrowserDetectorHandler If the vault token is read from the cookie it will assume that is a browser
//vault token from the cookie will always be selected if both header and cookie are present
func BrowserDetectorHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ibr := false
        if len(r.Cookies()) > 0 {
            ck, err := r.Cookie("X-Vault-Token")
            if err == nil {
                ibr = true
                r.Header.Add("X-Vault-Token", ck.Value)
                // remove Vault cookie
                cks := r.Cookies()
                r.Header.Del("Cookie")
                for _, c := range cks {
                    if c.Name == "X-Vault-Token" {
                        continue
                    }
                    r.AddCookie(c)
                }
            }
        }
        ctx := context.WithValue(r.Context(), IsBrowserRequest, ibr)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

//NextCSRFHandler set the next csrf token, js application
// should read this token and use it in the next request
func NextCSRFHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ct := csrf.Token(r)
        w.Header().Set("X-Next-CSRF-Token", ct)
        next.ServeHTTP(w, r)
    })
}

//APICSRFHandler add csrf protection only for browsers (see BrowserDetectorHandler)
func APICSRFHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var CSRF func(http.Handler) http.Handler
        var handler http.Handler
        handler = next
        isBrowserRequest := r.Context().Value(IsBrowserRequest).(bool)
        if r.Method == "GET" || isBrowserRequest {
            CSRF = csrf.Protect([]byte(config.Config.BlockKey), csrf.Domain(config.Config.Uris.Api + config.Config.HostWithoutPort()))
            if config.Config.Scheme() == "http" {
                CSRF = csrf.Protect([]byte(config.Config.BlockKey), csrf.Secure(false), csrf.Domain(config.Config.Uris.Api + config.Config.HostWithoutPort()))
            }
            handler = CSRF(NextCSRFHandler(handler))
        }
        handler.ServeHTTP(w, r)
    })
}

