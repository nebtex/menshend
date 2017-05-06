package strategy

import (
    "github.com/nebtex/menshend/pkg/pfclient/chisel/server"
    "net/http"
    mutils "github.com/nebtex/menshend/pkg/utils"
    "net/url"
    "github.com/nebtex/menshend/pkg/resolvers"
    vault "github.com/hashicorp/vault/api"
    "net"
)

type PortForward struct {
}

//PortForward ..
func (r *PortForward) Execute(rs resolvers.Resolver, tokenInfo *vault.Secret) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Context().Value(mutils.IsBrowserRequest).(bool) {
            panic(mutils.BadRequest.WithUserMessage("This endpoint can not be used from the browser, download the menshend client"))
        }
        b := rs.Resolve(tokenInfo)
        if !b.Passed() {
            panic(mutils.NotAuthorized.WithUserMessage(b.Error().Error()))
        }
        var err error
        URL, err := url.Parse(b.BaseUrl())
        mutils.HttpCheckPanic(err, mutils.InternalError)
        remote := URL.Host
        
        //test the remote
        conn, err := net.Dial(URL.Scheme, URL.Host)
        if err != nil {
            panic(mutils.BadGateway.Append(err.Error()).WithUserMessage("The backend is not online"))
        } else {
            conn.Close()
        }
        
        chiselServer, err := chserver.NewServer(&chserver.Config{
            Remote:remote,
        })
        mutils.HttpCheckPanic(err, mutils.InternalError)
        chiselServer.HandleHTTP(w, r)
    })
}


