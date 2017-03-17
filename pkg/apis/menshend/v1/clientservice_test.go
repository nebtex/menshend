package v1

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    . "github.com/nebtex/menshend/pkg/utils/test"
    . "github.com/nebtex/menshend/pkg/config"
    "net/http"
    "net/http/httptest"
    "github.com/emicklei/go-restful"
    "io/ioutil"
    "encoding/json"
)

func Test_ListClientService(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    Convey("get by role", t, func(c C) {
        CleanVault()
        PopulateVault()
        httpReq, err := http.NewRequest("GET", "/v1/clientServices?subdomain=gitlab.&role=ml-team", nil)
        httpReq.Header.Set("Content-Type", "application/json")
        So(err, ShouldBeNil)
        httpReq.Header.Add("X-Vault-Token", "myroot")
        httpWriter := httptest.NewRecorder()
        wsContainer := restful.NewContainer()
        s := ClientServiceResource{}
        s.Register(wsContainer)
        wsContainer.ServeHTTP(httpWriter, httpReq)
        jsres, err := ioutil.ReadAll(httpWriter.Body)
        So(err, ShouldBeNil)
        ret := &[]ClientServiceResource{}
        err = json.Unmarshal(jsres, ret)
        So(err, ShouldBeNil)
        So(len(*ret), ShouldEqual, 1)
        
    })
    
    Convey("get by role", t, func(c C) {
        CleanVault()
        PopulateVault()
        httpReq, err := http.NewRequest("GET", "/v1/clientServices?role=ml-team", nil)
        httpReq.Header.Set("Content-Type", "application/json")
        So(err, ShouldBeNil)
        httpReq.Header.Add("X-Vault-Token", "myroot")
        httpWriter := httptest.NewRecorder()
        wsContainer := restful.NewContainer()
        s := ClientServiceResource{}
        s.Register(wsContainer)
        wsContainer.ServeHTTP(httpWriter, httpReq)
        jsres, err := ioutil.ReadAll(httpWriter.Body)
        So(err, ShouldBeNil)
        ret := &[]ClientServiceResource{}
        err = json.Unmarshal(jsres, ret)
        So(err, ShouldBeNil)
        So(len(*ret), ShouldEqual, 5)
        
    })
    
    
    Convey("get by subdomain", t, func(c C) {
        CleanVault()
        PopulateVault()
        httpReq, err := http.NewRequest("GET", "/v1/clientServices?subdomain=redis.", nil)
        httpReq.Header.Set("Content-Type", "application/json")
        So(err, ShouldBeNil)
        httpReq.Header.Add("X-Vault-Token", "myroot")
        httpWriter := httptest.NewRecorder()
        wsContainer := restful.NewContainer()
        s := ClientServiceResource{}
        s.Register(wsContainer)
        wsContainer.ServeHTTP(httpWriter, httpReq)
        jsres, err := ioutil.ReadAll(httpWriter.Body)
        So(err, ShouldBeNil)
        ret := &[]ClientServiceResource{}
        err = json.Unmarshal(jsres, ret)
        So(err, ShouldBeNil)
        So(len(*ret), ShouldEqual, 2)
        
    })
    
    Convey("get  all the services", t, func(c C) {
        CleanVault()
        PopulateVault()
        httpReq, err := http.NewRequest("GET", "/v1/clientServices", nil)
        httpReq.Header.Set("Content-Type", "application/json")
        So(err, ShouldBeNil)
        httpReq.Header.Add("X-Vault-Token", "myroot")
        httpWriter := httptest.NewRecorder()
        wsContainer := restful.NewContainer()
        s := ClientServiceResource{}
        s.Register(wsContainer)
        wsContainer.ServeHTTP(httpWriter, httpReq)
        jsres, err := ioutil.ReadAll(httpWriter.Body)
        So(err, ShouldBeNil)
        ret := &[]ClientServiceResource{}
        err = json.Unmarshal(jsres, ret)
        So(err, ShouldBeNil)
        So(len(*ret), ShouldEqual, 8)
        
    })
    
    
}

