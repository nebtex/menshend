package strategy

import (
    "github.com/nebtex/menshend/pkg/pfclient/chisel/server"
    "github.com/nebtex/menshend/pkg/backend"
    "net/http"
    . "github.com/nebtex/menshend/pkg/utils"
    "github.com/ansel1/merry"
    "net/url"
    "fmt"
    "strings"
)

//BadRequest ...
var NotFound = merry.New("Resource not found").WithHTTPCode(404)
//PermissionError this mean that the acl token has not access to x key on consul
var PermissionError = merry.New("Permission Error").WithHTTPCode(403)
var BadRequest = merry.New("Bad request").WithHTTPCode(400)
var BadGateway = merry.New("Bad Gateway").WithHTTPCode(http.StatusBadGateway)


type PortForward struct {
}

//PortForward ..
func (r *PortForward) Execute(b backend.Backend) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var err error
        URL, err := url.Parse(b.BaseUrl())
        HttpCheckPanic(err, InternalError)
        host := strings.Split(URL.Host, ":")[0]
        remote := host + ":" + r.Header.Get("X-Menshend-Port-Forward")
        fmt.Println(remote)
    
        chiselServer, err := chserver.NewServer(&chserver.Config{
            Remote:remote,
        })
        HttpCheckPanic(err, InternalError)
        chiselServer.HandleHTTP(w, r)
    }
}


