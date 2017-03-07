package strategy

import (
    "testing"
    "net/http"
    . "github.com/smartystreets/goconvey/convey"
    . "github.com/nebtex/menshend/pkg/pfclient"
    . "github.com/nebtex/menshend/pkg/users"
    "os"
    . "github.com/nebtex/menshend/pkg/utils"
    "github.com/parnurzeal/gorequest"
    "fmt"
    "time"
)

func TestPortForward_Execute(t *testing.T) {
    
    Convey("Should forward port", t, func(c C) {
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.SetExpiresAt(GetNow() + 3600)
        os.Setenv("MENSHEND_TOKEN", user.GenerateJWT())
        
        tb := &testBackend{url:"http://localhost:8200"}
        r := PortForward{}
        http.HandleFunc("/", r.Execute(tb))
        go http.ListenAndServe(":9090", nil)
        go PFConnect(true, 0, "http://vault.localhost:9090", "25300:8200")
        go PFConnect(true, 0, "http://vault.localhost:9090", "25400:8200")
        go PFConnect(true, 0, "http://vault.localhost:9090", "25500:8200")
        go PFConnect(true, 0, "http://vault.localhost:9090", "25600:8200")
        go PFConnect(true, 0, "http://vault.localhost:9090", "25700:8200")
        go PFConnect(true, 0, "http://vault.localhost:9090", "25800:8200")
        
        time.Sleep(5 * time.Second)
        for _, port := range []string{"25300", "25400", "25500", "25600", "25700", "25800"} {
            _, body, err := gorequest.New().Get("http://localhost:" + port + "/v1/sys/seal-status").
                Set("X-Vault-Token", "myroot").End()
            So(err, ShouldBeNil)
            fmt.Println(body)
        }
        
    })
    
}
