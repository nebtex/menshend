package menshend

import (
    . "github.com/smartystreets/goconvey/convey"
    "github.com/levigross/grequests"
    "testing"
    "time"
    "fmt"
)

func TestServer(t *testing.T) {
    //init http server in the background
    VaultConfig.Address = "http://127.0.0.1:8200"
    Config.ListenPort = 18080
    Config.Host = "localhost"
    Config.Scheme = "http"
    go main()
    time.Sleep(1 * time.Second)
    Convey("Test routers ", t, func() {
        cleanVault()
    
        session := grequests.NewSession(&grequests.RequestOptions{Host:Config.HostWithoutPort()})
        response, err := session.Get("http://localhost:18080/", nil)
        So(err, ShouldBeNil)
        csrfToken := response.Header.Get("X-CSRF-Token")
    
        options := &grequests.RequestOptions{Host:Config.HostWithoutPort()}
        options.Headers = map[string]string{}
        options.Headers["X-CSRF-Token"] = csrfToken
        options.JSON = TokenLogin{Token:"myroot"}
        response, err = session.Post("http://localhost:18080/login/token", options)
        So(err, ShouldBeNil)
        csrfToken = response.Header.Get("X-CSRF-Token")
        fmt.Println(csrfToken)
        //save services consul, vault, redis
        //session.Post()
    })
}
GetNow()
