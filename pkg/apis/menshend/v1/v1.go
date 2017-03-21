package v1

import (
    "net/http"
    "github.com/emicklei/go-restful"
    "github.com/ansel1/merry"
    "github.com/Sirupsen/logrus"
    "fmt"
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
    return ApiPanicHandler(wsContainer)
}


//ApiPanicHandler handle any panic in the api endpoint
func ApiPanicHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var errorMessage string
        var errorCode int
        
        defer func() {
            rec := recover()
            if (rec == nil) {
                return
            }
            switch x := rec.(type) {
            case merry.Error:
                logrus.Errorln(merry.Details(x))
                errorMessage = merry.UserMessage(x)
                errorCode = merry.HTTPCode(x)
            case error:
                logrus.Errorln(x)
                errorMessage = "Internal server error"
                errorCode = http.StatusInternalServerError
            default:
                errorMessage = "Uknown error"
                errorCode = http.StatusInternalServerError
            }
            w.Write([]byte{fmt.Sprintf(`{"message": "%s"}`, errorMessage)})
            w.WriteHeader(errorCode)
        }()
        
        next.ServeHTTP(w, r)
    })
}


//BrowserDetectorHandler If the vault token is read from the cookie it will assume that is a browser
//vault token from the cookie will always be selected if both header and cookie are present



//ApiPanicHandler handle any panic in the api endpoint

func ApiCRSFHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var errorMessage string
        var errorCode int
        
        defer func() {
            rec := recover()
            if (rec == nil) {
                return
            }
            switch x := rec.(type) {
            case merry.Error:
                logrus.Errorln(merry.Details(x))
                errorMessage = merry.UserMessage(x)
                errorCode = merry.HTTPCode(x)
            case error:
                logrus.Errorln(x)
                errorMessage = "Internal server error"
                errorCode = http.StatusInternalServerError
            default:
                errorMessage = "Uknown error"
                errorCode = http.StatusInternalServerError
            }
            w.Write([]byte{fmt.Sprintf(`{"message": "%s"}`, errorMessage)})
            w.WriteHeader(errorCode)
        }()
        
        next.ServeHTTP(w, r)
    })
}

