package strategy

import (
    "testing"
    "net/http"
    "github.com/nebtex/menshend/pkg/resolvers"
    mutils "github.com/nebtex/menshend/pkg/utils"
    . "github.com/smartystreets/goconvey/convey"
    "github.com/nebtex/menshend/pkg/pfclient"
    "time"
    vault "github.com/hashicorp/vault/api"
    "context"
    "github.com/parnurzeal/gorequest"
    "fmt"
    "github.com/ansel1/merry"
    "net/http/httptest"
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
        tb.Content = `baseUrl: tcp://localhost:8200`
        
        r := PortForward{}
        http.HandleFunc("/", noBrowserHandler(r.Execute(tb, &vault.Secret{})).ServeHTTP)
        go http.ListenAndServe(":9090", nil)
        
        go pfclient.PFConnect(true, 0, "http://vault.localhost:9090", "myroot", "demo", "25300")
        go pfclient.PFConnect(true, 0, "http://vault.localhost:9090", "myroot", "demo", "25400")
        go pfclient.PFConnect(true, 0, "http://vault.localhost:9090", "myroot", "demo", "25500")
        go pfclient.PFConnect(true, 0, "http://vault.localhost:9090", "myroot", "demo", "25600")
        go pfclient.PFConnect(true, 0, "http://vault.localhost:9090", "myroot", "demo", "25700")
        go pfclient.PFConnect(true, 0, "http://vault.localhost:9090", "myroot", "demo", "25800")
        time.Sleep(5 * time.Second)
        nr, err := http.NewRequest("GET", "http://vault.localhost:9090", nil)
        
        So(err, ShouldBeNil)
        client := &http.Client{}
        resp, err := client.Do(nr)
        So(err, ShouldBeNil)
        So(resp.StatusCode, ShouldEqual, 200)
        
        for _, port := range []string{"25300", "25400", "25500", "25600", "25700", "25800"} {
            _, body, err := gorequest.New().Get("http://localhost:" + port + "/v1/sys/seal-status").
                Set("X-Vault-Token", "myroot").End()
            So(err, ShouldBeNil)
            fmt.Println(body)
        }
        
    })
    
}


func TestPortForward_BadGateway(t *testing.T) {
    
    Convey("Should panic if the backend is not online", t, func(c C) {
        defer func(c C) {
            r := recover()
            if (r == nil) {
                t.Error("did not panicked")
                t.Fail()
            }
            switch x := r.(type) {
            case error:
                c.So(merry.Is(x, mutils.BadGateway), ShouldBeTrue)
            default:
                t.Errorf("%v", x)
                t.Fail()
            }
        }(c)
        
        tb := &resolvers.YAMLResolver{}
        tb.Content = `baseUrl: tcp://abc.local:15000`
        
        r := PortForward{}
        nr, err := http.NewRequest("GET", "http://vault.localhost:9095", nil)
        So(err, ShouldBeNil)
        response:= httptest.NewRecorder()
        noBrowserHandler(r.Execute(tb, &vault.Secret{})).ServeHTTP(response, nr)
        
    })
}
