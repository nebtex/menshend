package strategy

import (
    "testing"
    "net/http"
    "github.com/nebtex/menshend/pkg/resolvers"
    mutils "github.com/nebtex/menshend/pkg/utils"
    . "github.com/smartystreets/goconvey/convey"
    pfclient "github.com/nebtex/menshend/pkg/pfclient"
    "os"
    "github.com/parnurzeal/gorequest"
    "fmt"
    "time"
    vault "github.com/hashicorp/vault/api"
    "context"
)

func noBrowserHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := context.WithValue(r.Context(), mutils.IsBrowserRequest, false)
        ctx = context.WithValue(ctx, mutils.Subdomain, "consul.")
       next.ServeHTTP(w, r.WithContext(ctx))
    })
}
func browserHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := context.WithValue(r.Context(), mutils.IsBrowserRequest, true)
        ctx = context.WithValue(ctx, mutils.Subdomain, "consul.")
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func TestPortForward_Execute(t *testing.T) {
    
    Convey("Should forward port", t, func(c C) {
        
        tb := &resolvers.YAMLResolver{}
        tb.Content = `baseUrl: http://localhost:8200
headersMap:
  h1: t1
  h2: t2`
        
        os.Setenv("VAULT_TOKEN", "myroot")
        r := PortForward{}
        http.HandleFunc("/", noBrowserHandler(r.Execute(tb, &vault.Secret{})).ServeHTTP)
        go http.ListenAndServe(":9090", nil)
        go pfclient.PFConnect(true, 0, "http://vault.localhost:9090", "25300:8200")
        go pfclient.PFConnect(true, 0, "http://vault.localhost:9090", "25400:8200")
        go pfclient.PFConnect(true, 0, "http://vault.localhost:9090", "25500:8200")
        go pfclient.PFConnect(true, 0, "http://vault.localhost:9090", "25600:8200")
        go pfclient.PFConnect(true, 0, "http://vault.localhost:9090", "25700:8200")
        go pfclient.PFConnect(true, 0, "http://vault.localhost:9090", "25800:8200")
        
        time.Sleep(5 * time.Second)
        for _, port := range []string{"25300", "25400", "25500", "25600", "25700", "25800"} {
            _, body, err := gorequest.New().Get("http://localhost:" + port + "/v1/sys/seal-status").
                Set("X-Vault-Token", "myroot").End()
            So(err, ShouldBeNil)
            fmt.Println(body)
        }
        
    })
    
}

