package v1

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    testutils "github.com/nebtex/menshend/pkg/utils/test"
    "net/http"
    "net/http/httptest"
    "github.com/emicklei/go-restful"
    "io/ioutil"
    "encoding/json"
    "os"
    vault "github.com/hashicorp/vault/api"
    mutils "github.com/nebtex/menshend/pkg/utils"
)

func listServices(c C, path string) []ClientServiceResource {
    testutils.CleanVault()
    testutils.PopulateVault()
    httpReq, err := http.NewRequest("GET", path, nil)
    httpReq.Header.Set("Content-Type", "application/json")
    c.So(err, ShouldBeNil)
    httpReq.Header.Add("X-Vault-Token", "myroot")
    httpWriter := httptest.NewRecorder()
    wsContainer := restful.NewContainer()
    s := ClientServiceResource{}
    s.Register(wsContainer)
    wsContainer.ServeHTTP(httpWriter, httpReq)
    jsres, err := ioutil.ReadAll(httpWriter.Body)
    c.So(err, ShouldBeNil)
    ret := []ClientServiceResource{}
    err = json.Unmarshal(jsres, &ret)
    c.So(err, ShouldBeNil)
    return ret
}

func Test_ListClientService(t *testing.T) {
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))
    Convey("get by role", t, func(c C) {
        ret:=listServices(c, "/v1/clientServices?subdomain=gitlab.&role=ml-team")
        So(len(ret), ShouldEqual, 1)
    })
    Convey("get by role", t, func(c C) {
        ret:=listServices(c, "/v1/clientServices?role=ml-team")
        So(len(ret), ShouldEqual, 5)
    })
    Convey("get by subdomain", t, func(c C) {
        ret:=listServices(c, "/v1/clientServices?subdomain=redis.")
        So(len(ret), ShouldEqual, 2)
    })
    Convey("get  all the services", t, func(c C) {
        ret:=listServices(c, "/v1/clientServices")
        So(len(ret), ShouldEqual, 8)
    })
}

