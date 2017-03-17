package menshend

import (
    vault "github.com/hashicorp/vault/api"
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "github.com/ansel1/merry"
    "net/http/httptest"
    "github.com/emicklei/go-restful"
    "net/http"
    "github.com/Sirupsen/logrus"
    . "github.com/nebtex/menshend/pkg/utils/test"
    . "github.com/nebtex/menshend/pkg/config"
    . "github.com/nebtex/menshend/pkg/utils"

)

func Test_IsAdmin(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    Convey("This function should indicate if the user is admin or" +
        " not", t, func() {
        Convey("Should return false if th user is not an admin", func() {
            CleanVault()
            So(IsAdmin("test_token"), ShouldBeFalse)
            vaultClient, vaultErr := vault.NewClient(VaultConfig)
            So(vaultErr, ShouldBeNil)
            vaultClient.SetToken("myroot")
            
            vaultErr = vaultClient.Sys().PutPolicy("check-capabilities", `
        path "/sys/capabilities-self" { policy = "read" }
            `)
            So(vaultErr, ShouldBeNil)
            secret, vaultErr := vaultClient.Auth().Token().
                Create(&vault.TokenCreateRequest{
                Policies:[]string{"check-capabilities"}})
            So(vaultErr, ShouldBeNil)
            So(IsAdmin(secret.Auth.ClientToken), ShouldBeFalse)
        })
        Convey("Should return true if th user is  an admin", func() {
            CleanVault()
            So(IsAdmin("myroot"), ShouldBeTrue)
        })
    })
}

func Test_CanImpersonateHandler(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    
    Convey("This endpoint should indicate if the user can impersonate or" +
        " not", t, func() {
        Convey("Should return false if th user can't impersonate", func() {
            CleanVault()
            So(CanImpersonate("test_token"), ShouldBeFalse)
            
            vaultClient, vaultErr := vault.NewClient(VaultConfig)
            So(vaultErr, ShouldBeNil)
            vaultClient.SetToken("myroot")
            
            vaultErr = vaultClient.Sys().PutPolicy("check-capabilities", `
        path "/sys/capabilities-self" { policy = "read" }
            `)
            So(vaultErr, ShouldBeNil)
            secret, vaultErr := vaultClient.Auth().Token().
                Create(&vault.TokenCreateRequest{
                Policies:[]string{"check-capabilities"}})
            So(vaultErr, ShouldBeNil)
            So(CanImpersonate(secret.Auth.ClientToken), ShouldBeFalse)
            
        })
        
        Convey("Should return true if th user can impersonate", func() {
            CleanVault()
            So(CanImpersonate("myroot"), ShouldBeTrue)
        })
    })
    
}

func TestLoginFilter(t *testing.T) {
    Convey("Should panic when there is not a jwt token", t,
        func(c C) {
            defer func() {
                r := recover()
                switch x := r.(type) {
                case error:
                    c.So(merry.Is(x, NotAuthorized), ShouldBeTrue)
                default:
                    t.Errorf("%v", x)
                    t.Fail()
                }
            }()
            httpReq, _ := http.NewRequest("GET", "/", nil)
            req := restful.NewRequest(httpReq)
            
            recorder := new(httptest.ResponseRecorder)
            resp := restful.NewResponse(recorder)
            target := func(freq *restful.Request, fresp *restful.Response) {}
            tf := &restful.FilterChain{Target:target}
            panicked := false
            LoginFilter(req, resp, tf)
            So(panicked, ShouldBeTrue)
            
        })
    
    Convey("Should panic if there is an invalid token", t,
        func(c C) {
            defer func() {
                r := recover()
                logrus.Error(r)
                
                switch x := r.(type) {
                case error:
                    c.So(merry.Is(x, NotAuthorized), ShouldBeTrue)
                case nil:
                    t.Errorf("%v", "does not panic")
                    t.Fail()
                default:
                    t.Errorf("%v", x)
                    t.Fail()
                }
            }()
            httpReq, _ := http.NewRequest("GET", "/", nil)
            req := restful.NewRequest(httpReq)
            req.Request.Header.Add("X-Vault-Token", "test-token")
            
            recorder := new(httptest.ResponseRecorder)
            resp := restful.NewResponse(recorder)
            target := func(freq *restful.Request, fresp *restful.Response) {}
            tf := &restful.FilterChain{Target:target}
            LoginFilter(req, resp, tf)
            panicked := false
            So(panicked, ShouldBeTrue)
            
        })
    
    Convey("Should make the user available in the context if the token" +
        " is valid", t, func(c C) {
        httpReq, _ := http.NewRequest("GET", "/", nil)
        req := restful.NewRequest(httpReq)
        req.Request.Header.Add("X-Vault-Token", "test-token")
        
        recorder := new(httptest.ResponseRecorder)
        resp := restful.NewResponse(recorder)
        target := func(freq *restful.Request, fresp *restful.Response) {
            user := GetTokenFromRequest(freq)
            So(user, ShouldNotBeNil)
        }
        tf := &restful.FilterChain{Target:target}
        LoginFilter(req, resp, tf)
        
    })
}

func TestAdminFilter(t *testing.T) {
    Convey("Should panic when the user is not an admin", t,
        func(c C) {
            defer func() {
                r := recover()
                switch x := r.(type) {
                case error:
                    c.So(merry.Is(x, NotAuthorized), ShouldBeTrue)
                default:
                    t.Errorf("%v", x)
                    t.Fail()
                }
            }()
            
            httpReq, _ := http.NewRequest("GET", "/", nil)
            req := restful.NewRequest(httpReq)
            req.Request.Header.Add("X-Vault-Token", "test-token")
            recorder := new(httptest.ResponseRecorder)
            resp := restful.NewResponse(recorder)
            target := func(freq *restful.Request, fresp *restful.Response) {
            }
            tf := &restful.FilterChain{Target:target}
            AdminFilter(req, resp, tf)
            panicked := false
            So(panicked, ShouldBeTrue)
        })
    
}

func TestCanImpersonateFilter(t *testing.T) {
    Convey("Should panic when the user is can not impersonate", t,
        func(c C) {
            defer func() {
                r := recover()
                switch x := r.(type) {
                case error:
                    c.So(merry.Is(x, NotAuthorized), ShouldBeTrue)
                default:
                    t.Errorf("%v", x)
                    t.Fail()
                }
            }()
            
            httpReq, _ := http.NewRequest("GET", "/", nil)
            req := restful.NewRequest(httpReq)
            req.Request.Header.Add("X-Vault-Token", "test-toke")
            
            recorder := new(httptest.ResponseRecorder)
            resp := restful.NewResponse(recorder)
            target := func(freq *restful.Request, fresp *restful.Response) {
            }
            tf := &restful.FilterChain{Target:target}
            ImpersonateFilter(req, resp, tf)
            panicked := false
            So(panicked, ShouldBeTrue)
        })
    
}