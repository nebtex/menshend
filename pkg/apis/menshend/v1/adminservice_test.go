package v1

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "net/http/httptest"
    "net/http"
    . "github.com/nebtex/menshend/pkg/config"
    . "github.com/nebtex/menshend/pkg/utils"
    . "github.com/nebtex/menshend/pkg/utils/test"
    . "github.com/nebtex/menshend/pkg/users"
    . "github.com/nebtex/menshend/pkg/apis/menshend"
    
    "github.com/emicklei/go-restful"
    "github.com/ansel1/merry"
    
    vault "github.com/hashicorp/vault/api"
    "bytes"
    "encoding/json"
    "io/ioutil"
    "fmt"
)

func TestCreateEditServiceHandler(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    Convey("Should create or modify a service", t, func() {
        Convey("Should save the service and return it as response",
            func(c C) {
                
                CleanVault()
                wsContainer := restful.NewContainer()
                u := AdminServiceResource{}
                u.Register(wsContainer)
                postBody, err := json.Marshal(&AdminServiceResource{ProxyLanguage:"lua", Strategy:"redirect"})
                So(err, ShouldBeNil)
                httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", bytes.NewReader(postBody))
                So(err, ShouldBeNil)
                
                httpReq.Header.Set("Content-Type", "application/json")
                user, err := NewUser("myroot")
                So(err, ShouldBeNil)
                user.TokenLogin()
                user.SetExpiresAt(GetNow() + 3600)
                httpReq.Header.Add("X-Menshend-Token", user.GenerateJWT())
                httpWriter := httptest.NewRecorder()
                wsContainer.ServeHTTP(httpWriter, httpReq)
                jsres, err := ioutil.ReadAll(httpWriter.Body)
                So(err, ShouldBeNil)
                rService := &AdminServiceResource{}
                err = json.Unmarshal(jsres, rService)
                So(err, ShouldBeNil)
                So(rService.RoleID, ShouldEqual, "ml-team")
                So(rService.SubDomain, ShouldEqual, "gitlab.")
            })
    })
}

func Test_LoadLongDescriptionFromUrl(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    
    Convey("if long description url is defined should use it for populate " +
        "long description", t, func(c C) {
        
        CleanVault()
        wsContainer := restful.NewContainer()
        u := AdminServiceResource{}
        u.Register(wsContainer)
        postBody, err := json.Marshal(&AdminServiceResource{ProxyLanguage:"lua", Strategy:"redirect", LongDescriptionUrl:"https://raw.githubusercontent.com/gitlabhq/gitlabhq/master/README.md"})
        So(err, ShouldBeNil)
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", bytes.NewReader(postBody))
        So(err, ShouldBeNil)
        
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        httpReq.Header.Add("X-Menshend-Token", user.GenerateJWT())
        httpWriter := httptest.NewRecorder()
        wsContainer.ServeHTTP(httpWriter, httpReq)
        jsres, err := ioutil.ReadAll(httpWriter.Body)
        So(err, ShouldBeNil)
        rService := &AdminServiceResource{}
        err = json.Unmarshal(jsres, rService)
        So(err, ShouldBeNil)
        So(rService.RoleID, ShouldEqual, "ml-team")
        So(rService.SubDomain, ShouldEqual, "gitlab.")
        So(rService.LongDescription, ShouldNotBeEmpty)
    })
    
    Convey("should return error with a invalid url", t, func(c C) {
        defer func() {
            r := recover()
            if (r == nil) {
                t.Error("did not panicked")
                t.Fail()
            }
            switch x := r.(type) {
            case error:
                c.So(merry.Is(x, BadRequest), ShouldBeTrue)
            default:
                t.Errorf("%v", x)
                t.Fail()
            }
        }()
        
        CleanVault()
        wsContainer := restful.NewContainer()
        u := AdminServiceResource{}
        u.Register(wsContainer)
        postBody, err := json.Marshal(&AdminServiceResource{LongDescriptionUrl:"httpsinvali¡d/", Strategy:"redirect", ProxyLanguage:"lua" })
        So(err, ShouldBeNil)
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", bytes.NewReader(postBody))
        So(err, ShouldBeNil)
        
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        httpReq.Header.Add("X-Menshend-Token", user.GenerateJWT())
        httpWriter := httptest.NewRecorder()
        wsContainer.ServeHTTP(httpWriter, httpReq)
        
    })
    
    Convey("should return error if it cant load the url", t, func(c C) {
        defer func() {
            r := recover()
            if (r == nil) {
                t.Error("did not panicked")
                t.Fail()
            }
            switch x := r.(type) {
            case error:
                c.So(merry.Is(x, BadRequest), ShouldBeTrue)
            default:
                t.Errorf("%v", x)
                t.Fail()
            }
        }()
        
        CleanVault()
        wsContainer := restful.NewContainer()
        u := AdminServiceResource{}
        u.Register(wsContainer)
        postBody, err := json.Marshal(&AdminServiceResource{LongDescriptionUrl:"http://example.loca:54545/Readme.md", Strategy:"redirect", ProxyLanguage:"lua"})
        So(err, ShouldBeNil)
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", bytes.NewReader(postBody))
        So(err, ShouldBeNil)
        
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        httpReq.Header.Add("X-Menshend-Token", user.GenerateJWT())
        httpWriter := httptest.NewRecorder()
        wsContainer.ServeHTTP(httpWriter, httpReq)
        
    })
}

func TestDeleteServiceHandler_Permissions(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    
    Convey("Should return permission error if the user can't delete a " +
        "service in the role", t, func(c C) {
        defer func() {
            r := recover()
            if (r == nil) {
                t.Error("did not panicked")
                t.Fail()
            }
            switch x := r.(type) {
            case error:
                c.So(merry.Is(x, PermissionError), ShouldBeTrue)
            default:
                t.Errorf("%v", x)
                t.Fail()
            }
        }()
        CleanVault()
        vClient, err := vault.NewClient(VaultConfig)
        So(err, ShouldBeNil)
        vClient.SetToken("myroot")
        err = vClient.Sys().PutPolicy("admin-test-cesh", `
        path "secret/menshend/roles/devops/*" { policy = "write" }
        path "secret/menshend/Admin" { policy = "write" }`)
        So(err, ShouldBeNil)
        secret, err := vClient.Auth().Token().Create(&vault.TokenCreateRequest{
            Policies:[]string{"admin-test-cesh"}})
        So(err, ShouldBeNil)
        
        user, err := NewUser(secret.Auth.ClientToken)
        So(err, ShouldBeNil)
        user.SetExpiresAt(GetNow() + 3600 * 1000)
        user.GitHubLogin("criloz", "admin", "delos", "umbrella")
        
        httpReq, err := http.NewRequest("DELETE", "/v1/adminServices/roles/ml-team/gitlab.", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        httpReq.Header.Add("X-Menshend-Token", user.GenerateJWT())
        httpWriter := httptest.NewRecorder()
        
        wsContainer := restful.NewContainer()
        u := AdminServiceResource{}
        u.Register(wsContainer)
        wsContainer.ServeHTTP(httpWriter, httpReq)
        
    })
}

func TestGetServiceHandler(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
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
                        c.So(merry.Is(x, NotFound), ShouldBeTrue)
                    default:
                        t.Errorf("%v", x)
                        t.Fail()
                    }
                }()
                
                CleanVault()
                
                wsContainer := restful.NewContainer()
                u := AdminServiceResource{}
                u.Register(wsContainer)
                httpReq, _ := http.NewRequest("GET", "/v1/adminServices/roles/admin/ldap.", nil)
                httpReq.Header.Set("Content-Type", "application/json")
                user, err := NewUser("myroot")
                So(err, ShouldBeNil)
                user.TokenLogin()
                user.SetExpiresAt(GetNow() + 3600)
                httpReq.Header.Add("X-Menshend-Token", user.GenerateJWT())
                recorder := new(httptest.ResponseRecorder)
                wsContainer.ServeHTTP(recorder, httpReq)
            })
        
        Convey("Should return  the service",
            func(c C) {
                CleanVault()
                PopulateVault()
                wsContainer := restful.NewContainer()
                u := AdminServiceResource{}
                u.Register(wsContainer)
                httpReq, _ := http.NewRequest("GET", "/v1/adminServices/roles/ml-team/gitlab.", nil)
                httpReq.Header.Set("Content-Type", "application/json")
                user, err := NewUser("myroot")
                So(err, ShouldBeNil)
                user.TokenLogin()
                user.SetExpiresAt(GetNow() + 3600)
                httpReq.Header.Add("X-Menshend-Token", user.GenerateJWT())
                httpWriter := httptest.NewRecorder()
                wsContainer.ServeHTTP(httpWriter, httpReq)
                So(httpWriter.Body, ShouldNotBeNil)
            })
    })
}

func TestListServiceHandler(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    Convey("Should get all the services with the same subdomian across roles", t, func() {
        CleanVault()
        PopulateVault()
        wsContainer := restful.NewContainer()
        u := AdminServiceResource{}
        u.Register(wsContainer)
        httpReq, _ := http.NewRequest("GET", "/v1/adminServices?subdomain=redis.", nil)
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        httpReq.Header.Add("X-Menshend-Token", user.GenerateJWT())
        httpWriter := httptest.NewRecorder()
        wsContainer.ServeHTTP(httpWriter, httpReq)
        fmt.Println(httpWriter.Body)
        So(httpWriter.Body, ShouldNotBeNil)
        
    })
}
