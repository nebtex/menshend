package v1

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "net/http/httptest"
    "net/http"
    mutils "github.com/nebtex/menshend/pkg/utils"
    testutils "github.com/nebtex/menshend/pkg/utils/test"
    "github.com/emicklei/go-restful"
    "github.com/ansel1/merry"
    vault "github.com/hashicorp/vault/api"
    "bytes"
    "encoding/json"
    "io/ioutil"
    "fmt"
    "os"
    "github.com/nebtex/menshend/pkg/resolvers"
    "github.com/nebtex/menshend/pkg/strategy"
)

func TestCreateEditServiceHandler(t *testing.T) {
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))
    Convey("Should create or modify a service", t, func() {
        Convey("Should save the service and return it as response",
            func(c C) {
                testutils.CleanVault()
                
                wsContainer := restful.NewContainer()
                u := AdminServiceResource{}
                u.Register(wsContainer)
                postBody, err := json.Marshal(&AdminServiceResource{
                    Meta:&ServiceMetadata{Name:"gitlab"},
                    Resolver:&ServiceResolver{Yaml: &resolvers.YAMLResolver{Content:"xx"}},
                    Strategy:&ServiceStrategy{Proxy: &strategy.Proxy{}},
                })
                So(err, ShouldBeNil)
                httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", bytes.NewReader(postBody))
                So(err, ShouldBeNil)
                httpReq.Header.Set("Content-Type", "application/json")
                httpReq.Header.Add("X-Vault-Token", "myroot")
                httpWriter := httptest.NewRecorder()
                wsContainer.ServeHTTP(httpWriter, httpReq)
                jsres, err := ioutil.ReadAll(httpWriter.Body)
                So(err, ShouldBeNil)
                rService := &AdminServiceResource{}
                err = json.Unmarshal(jsres, rService)
                So(err, ShouldBeNil)
                So(rService.Meta.RoleID, ShouldEqual, "ml-team")
                So(rService.Meta.SubDomain, ShouldEqual, "gitlab.")
            })
    })
}

func Test_LoadLongDescriptionFromUrl(t *testing.T) {
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))
    capturePanic := func(c C) {
        r := recover()
        if (r == nil) {
            t.Error("did not panicked")
            t.Fail()
        }
        switch x := r.(type) {
        case error:
            c.So(merry.Is(x, mutils.BadRequest), ShouldBeTrue)
        default:
            t.Errorf("%v", x)
            t.Fail()
        }
    }
    Convey("if long description url is defined should use it for populate " +
        "long description", t, func(c C) {
        
        testutils.CleanVault()
        wsContainer := restful.NewContainer()
        u := AdminServiceResource{}
        u.Register(wsContainer)
        postBody, err := json.Marshal(
            &AdminServiceResource{
                Meta:&ServiceMetadata{Name:"gitlab",
                    LongDescription: &ServiceLongDescription{
                        Remote:&URLLongDescription{URL: "https://raw.githubusercontent.com/gitlabhq/gitlabhq/master/README.md"}}},
                Resolver:&ServiceResolver{Yaml: &resolvers.YAMLResolver{Content:"xxx"}},
                Strategy:&ServiceStrategy{Proxy: &strategy.Proxy{}},
            })
        So(err, ShouldBeNil)
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", bytes.NewReader(postBody))
        So(err, ShouldBeNil)
        
        httpReq.Header.Set("Content-Type", "application/json")
        So(err, ShouldBeNil)
        httpReq.Header.Add("X-Vault-Token", "myroot")
        httpWriter := httptest.NewRecorder()
        wsContainer.ServeHTTP(httpWriter, httpReq)
        jsres, err := ioutil.ReadAll(httpWriter.Body)
        So(err, ShouldBeNil)
        rService := &AdminServiceResource{}
        err = json.Unmarshal(jsres, rService)
        So(err, ShouldBeNil)
        So(rService.Meta.RoleID, ShouldEqual, "ml-team")
        So(rService.Meta.SubDomain, ShouldEqual, "gitlab.")
        So(rService.Meta.LongDescription.LongDescription(), ShouldNotBeEmpty)
    })
    
    Convey("should return error with a invalid url", t, func(c C) {
        defer capturePanic(c)
        testutils.CleanVault()
        wsContainer := restful.NewContainer()
        u := AdminServiceResource{}
        u.Register(wsContainer)
        postBody, err := json.Marshal(
            &AdminServiceResource{
                Meta:&ServiceMetadata{Name:"gitlab", LongDescription: &ServiceLongDescription{Remote:&URLLongDescription{URL: "httpsinvali¡d/"}}},
                Resolver:&ServiceResolver{Yaml: &resolvers.YAMLResolver{}},
                Strategy:&ServiceStrategy{Proxy: &strategy.Proxy{}},
            })
        So(err, ShouldBeNil)
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", bytes.NewReader(postBody))
        So(err, ShouldBeNil)
        
        httpReq.Header.Set("Content-Type", "application/json")
        So(err, ShouldBeNil)
        httpReq.Header.Add("X-Vault-Token", "myroot")
        httpWriter := httptest.NewRecorder()
        wsContainer.ServeHTTP(httpWriter, httpReq)
        
    })
    
    Convey("should return error if it cant load the url", t, func(c C) {
        defer capturePanic(c)
        testutils.CleanVault()
        wsContainer := restful.NewContainer()
        u := AdminServiceResource{}
        u.Register(wsContainer)
        postBody, err := json.Marshal(&AdminServiceResource{
            Meta:&ServiceMetadata{Name:"gitlab",
                LongDescription:&ServiceLongDescription{
                    Remote:&URLLongDescription{URL: "http://example.loca:54545/Readme.md"}}},
            Resolver:&ServiceResolver{Yaml: &resolvers.YAMLResolver{}},
            Strategy:&ServiceStrategy{Proxy: &strategy.Proxy{}},
        })
        So(err, ShouldBeNil)
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", bytes.NewReader(postBody))
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        So(err, ShouldBeNil)
        httpReq.Header.Add("X-Vault-Token", "myroot")
        httpWriter := httptest.NewRecorder()
        wsContainer.ServeHTTP(httpWriter, httpReq)
        
    })
}

func TestDeleteServiceHandler_Permissions(t *testing.T) {
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))
    
    Convey("Should return permission error if the user can't delete a " +
        "service in the role", t, func(c C) {
        testutils.CleanVault()
        vClient, err := vault.NewClient(vault.DefaultConfig())
        So(err, ShouldBeNil)
        vClient.SetToken("myroot")
        err = vClient.Sys().PutPolicy("admin-test-cesh", `
        path "secret/menshend/roles/devops/*" { policy = "write" }
        path "secret/menshend/Admin" { policy = "write" }`)
        So(err, ShouldBeNil)
        secret, err := vClient.Auth().Token().Create(&vault.TokenCreateRequest{Policies:[]string{"admin-test-cesh"}})
        So(err, ShouldBeNil)
        httpReq, err := http.NewRequest("DELETE", "/v1/adminServices/roles/ml-team/gitlab.", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        httpReq.Header.Add("X-Vault-Token", secret.Auth.ClientToken)
        httpWriter := httptest.NewRecorder()
        wsContainer := APIHandler()
        wsContainer.ServeHTTP(httpWriter, httpReq)
        So(httpWriter.Result().StatusCode, ShouldEqual, 403)
        
    })
}

func TestGetServiceHandler(t *testing.T) {
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))
    Convey("Should get a service by its path", t, func() {
        Convey("if service does not exist should return error",
            func(c C) {
                defer func() {
                    r := recover()
                    if (r == nil) {
                        t.Error("did not panicked")
                        t.Fail()
                    }
                    switch x := r.(type) {
                    case error:
                        c.So(merry.Is(x, mutils.NotFound), ShouldBeTrue)
                    default:
                        t.Errorf("%v", x)
                        t.Fail()
                    }
                }()
                
                testutils.CleanVault()
                
                wsContainer := restful.NewContainer()
                u := AdminServiceResource{}
                u.Register(wsContainer)
                httpReq, _ := http.NewRequest("GET", "/v1/adminServices/roles/admin/ldap.", nil)
                httpReq.Header.Set("Content-Type", "application/json")
                httpReq.Header.Add("X-Vault-Token", "myroot")
                recorder := new(httptest.ResponseRecorder)
                wsContainer.ServeHTTP(recorder, httpReq)
            })
        
        Convey("Should return  the service",
            func(c C) {
                testutils.CleanVault()
                testutils.PopulateVault()
                wsContainer := restful.NewContainer()
                u := AdminServiceResource{}
                u.Register(wsContainer)
                httpReq, _ := http.NewRequest("GET", "/v1/adminServices/roles/ml-team/gitlab.", nil)
                httpReq.Header.Set("Content-Type", "application/json")
                httpReq.Header.Add("X-Vault-Token", "myroot")
                httpWriter := httptest.NewRecorder()
                wsContainer.ServeHTTP(httpWriter, httpReq)
                So(httpWriter.Body, ShouldNotBeNil)
            })
    })
}

func TestListServiceHandler(t *testing.T) {
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))
    Convey("Should get all the services with the same subdomian across roles", t, func() {
        testutils.CleanVault()
        testutils.PopulateVault()
        wsContainer := restful.NewContainer()
        u := AdminServiceResource{}
        u.Register(wsContainer)
        httpReq, _ := http.NewRequest("GET", "/v1/adminServices?subdomain=redis.", nil)
        httpReq.Header.Set("Content-Type", "application/json")
        httpReq.Header.Add("X-Vault-Token", "myroot")
        httpWriter := httptest.NewRecorder()
        wsContainer.ServeHTTP(httpWriter, httpReq)
        fmt.Println(httpWriter.Body)
        So(httpWriter.Body, ShouldNotBeNil)
        
    })
}
