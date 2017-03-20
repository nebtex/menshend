package v1

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "net/http/httptest"
    "net/http"
    "encoding/json"
    "github.com/emicklei/go-restful"
    "io/ioutil"
    . "github.com/nebtex/menshend/pkg/utils/test"
    "fmt"
    vault "github.com/hashicorp/vault/api"
    "os"
)

func TestAccountStatus(t *testing.T) {
    os.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
    Convey("Should return user status", t, func() {
        CleanVault()
        wsContainer := ApiHandler()
        httpReq, err := http.NewRequest("GET", "/v1/account", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        So(err, ShouldBeNil)
        httpReq.Header.Add("X-Vault-Token", "myroot")
        httpWriter := httptest.NewRecorder()
        wsContainer.ServeHTTP(httpWriter, httpReq)
        jsres, err := ioutil.ReadAll(httpWriter.Body)
        So(err, ShouldBeNil)
        status := &LoginStatus{}
        fmt.Println(string(jsres))
        err = json.Unmarshal(jsres, status)
        So(err, ShouldBeNil)
        So(status.IsLogged, ShouldEqual, true)
        So(status.IsAdmin, ShouldEqual, true)
        So(status.CanImpersonate, ShouldEqual, true)
        
    })
    Convey("Test when user is not logged", t, func() {
		  CleanVault()
		  wsContainer := restful.NewContainer()
		  ar := AuthResource{}
		  ar.Register(wsContainer)
		  httpReq, err := http.NewRequest("GET", "/v1/account", nil)
		  So(err, ShouldBeNil)
		  httpReq.Header.Set("Content-Type", "application/json")
		  httpWriter := httptest.NewRecorder()
		  wsContainer.ServeHTTP(httpWriter, httpReq)
		  jsres, err := ioutil.ReadAll(httpWriter.Body)
		  So(err, ShouldBeNil)
		  status := &LoginStatus{}
		  fmt.Println(string(jsres))
		  err = json.Unmarshal(jsres, status)
		  So(err, ShouldBeNil)
		  So(status.IsLogged, ShouldEqual, false)
		  So(status.IsAdmin, ShouldEqual, false)
		  So(status.CanImpersonate, ShouldEqual, false)
    })
}

func TestLogout(t *testing.T) {
    os.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
    Convey("Should logout", t, func() {
        CleanVault()
        //create token
        vc, err := vault.NewClient(vault.DefaultConfig())
        So(err, ShouldBeNil)
        vc.SetToken("myroot")
        secret, err := vc.Auth().Token().Create(nil)
        So(err, ShouldBeNil)
        
        wsContainer := restful.NewContainer()
        ar := AuthResource{}
        ar.Register(wsContainer)
        httpReq, err := http.NewRequest("DELETE", "/v1/account", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        httpReq.Header.Add("X-Vault-Token", secret.Auth.ClientToken)
        httpWriter := httptest.NewRecorder()
        wsContainer.ServeHTTP(httpWriter, httpReq)
        _, err = vc.Auth().Token().Lookup(secret.Auth.ClientToken)
        So(err, ShouldNotBeNil)
    })
}
