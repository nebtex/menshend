package v1

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "net/http/httptest"
    "net/http"
    "encoding/json"
    "github.com/emicklei/go-restful"
    "io/ioutil"
    testutils "github.com/nebtex/menshend/pkg/utils/test"
    mutils "github.com/nebtex/menshend/pkg/utils"
    
    "fmt"
    vault "github.com/hashicorp/vault/api"
    "os"
)

func TestAccountStatus(t *testing.T) {
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))
    Convey("Should return user status", t, func() {
        testutils.CleanVault()
        wsContainer := APIHandler()
        httpReq, err := http.NewRequest("GET", "/v1/account", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        So(err, ShouldBeNil)
    
        vClient, err := vault.NewClient(vault.DefaultConfig())
        So(err, ShouldBeNil)
        vClient.SetToken("myroot")
        err = vClient.Sys().PutPolicy("admin-test-cesh", `
        path "secret/menshend/roles/devops/*" { policy = "write" }
        path "secret/menshend/Admin" { policy = "write" },
        path "/auth/token/lookup-self"  { policy = "read" } `)
        So(err, ShouldBeNil)
        secret, err := vClient.Auth().Token().Create(&vault.TokenCreateRequest{Policies:[]string{"admin-test-cesh"}, TTL:"1h"})
        So(err, ShouldBeNil)
        
    
        httpReq.Header.Add("X-Vault-Token", secret.Auth.ClientToken)
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
        So(status.CanImpersonate, ShouldEqual, false)
        
    })
    Convey("Test when user is not logged", t, func() {
        testutils.CleanVault()
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
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))
    Convey("Should logout", t, func() {
        testutils.CleanVault()
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
