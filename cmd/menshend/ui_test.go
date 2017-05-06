package main

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "time"
    mtestutils "github.com/nebtex/menshend/pkg/utils/test"
    mutils "github.com/nebtex/menshend/pkg/utils"
    mconfig "github.com/nebtex/menshend/pkg/config"
    "os"
    vault "github.com/hashicorp/vault/api"
    "net/http/httptest"
    "github.com/ansel1/merry"
    "net/url"
    "github.com/markbates/goth"
    "github.com/markbates/goth/providers/github"

)

func TestSetToken(t *testing.T) {
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))
    
    Convey("should panic if token has no the lookup-self permission", t, func(c C) {
        defer func() {
            r := recover()
            if (r == nil) {
                t.Error("did not panicked")
                t.Fail()
            }
            switch x := r.(type) {
            case error:
                c.So(merry.Is(x, mutils.PermissionError), ShouldBeTrue)
            default:
                t.Errorf("%v", x)
                t.Fail()
            }
        }()
        httpWriter := httptest.NewRecorder()
        mtestutils.CleanVault()
        vClient, err := vault.NewClient(vault.DefaultConfig())
        So(err, ShouldBeNil)
        vClient.SetToken("myroot")
        err = vClient.Sys().PutPolicy("without-lookup-permission", `
        path "auth/token/lookup-self" {
            capabilities = ["deny"]
        }`)
        So(err, ShouldBeNil)
        secret, err := vClient.Auth().Token().Create(&vault.TokenCreateRequest{Policies:[]string{"without-lookup-permission"}})
        So(err, ShouldBeNil)
        setToken(secret.Auth.ClientToken, tokenProvider, httpWriter)
        
    })
    
    Convey("if vault token ttl  is different from 0, the cookie should use it as expiration time", t, func() {
        httpWriter := httptest.NewRecorder()
        vClient, err := vault.NewClient(vault.DefaultConfig())
        So(err, ShouldBeNil)
        vClient.SetToken("myroot")
        secret, err := vClient.Auth().Token().Create(&vault.TokenCreateRequest{TTL:"1h"})
        So(err, ShouldBeNil)
        setToken(secret.Auth.ClientToken, tokenProvider, httpWriter)
        So(httpWriter.Result().Cookies()[0].Expires.Unix(), ShouldAlmostEqual, time.Now().Unix() + 3600, 10)
        
    })
    
    Convey("when menshend base url use https, token should have the property secure marked as true", t, func() {
        mconfig.Config.Uris.BaseUrl = "https://nebtex.com"
        httpWriter := httptest.NewRecorder()
        setToken("myroot", tokenProvider, httpWriter)
        So(httpWriter.Result().Cookies()[0].Secure, ShouldBeTrue)
        
    })
    
    Convey("token domain should use the menshend baseurl", t, func() {
        mconfig.Config.Uris.BaseUrl = "https://nebtex.com:65656"
        httpWriter := httptest.NewRecorder()
        setToken("myroot", tokenProvider, httpWriter)
        So(httpWriter.Result().Cookies()[0].Domain, ShouldEqual, "nebtex.com")
        
    })
}

func TestProviderRedirect(t *testing.T) {
    mconfig.Config.Github.ClientID = "test"
    mconfig.Config.Github.ClientSecret = "test-test"
    mconfig.LoadConfig()
    Convey("if not subdamin was provided should not set the redirect_uri property", t, func(c C) {
        rt:=  loginRouters()
        httpWriter := httptest.NewRecorder()
        request := httptest.NewRequest("GET", "/ui/auth/github", nil)
        rt.ServeHTTP(httpWriter, request)
        URL, err := httpWriter.Result().Location()
        So(err, ShouldBeNil)
        So(URL.Query().Get("redirect_uri"), ShouldEqual, mconfig.Config.GetCallbackUrl("github", ""))
        
    })
    Convey("if  subdamin was provided the redirect uri should use a subdomain as subpath", t, func(c C) {
        rt:=  loginRouters()
        httpWriter := httptest.NewRecorder()
        request := httptest.NewRequest("GET", "/ui/auth/github/gitlab.", nil)
        rt.ServeHTTP(httpWriter, request)
        URL, err := httpWriter.Result().Location()
        So(err, ShouldBeNil)
        So(URL.Query().Get("redirect_uri"), ShouldEqual, mconfig.Config.GetCallbackUrl("github", "gitlab."))
        
    })
    
}


func TestTokenLogin(t *testing.T) {
    mconfig.LoadConfig()
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))
    goth.UseProviders(
        github.New("test", "test-test", "read:org"),
    )
    Convey("test login method, without subdomain", t, func(c C) {
        rt:=  loginRouters()
        httpWriter := httptest.NewRecorder()
        request := httptest.NewRequest("POST", "/ui/login", nil)
        form := url.Values{}
        form.Add("method", "token")
        form.Add("token", "myroot")
        request.PostForm = form
        request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
        rt.ServeHTTP(httpWriter, request)
        URL, err := httpWriter.Result().Location()
        So(err, ShouldBeNil)
        So(URL.String(), ShouldEqual, mconfig.Config.GetServicePath())
        So(httpWriter.Result().Cookies()[0].Value, ShouldEqual, "myroot")
    })
    
    Convey("test login method, with subdomain", t, func(c C) {
        rt:=  loginRouters()
        httpWriter := httptest.NewRecorder()
        request := httptest.NewRequest("POST", "/ui/login/mongo.", nil)
        form := url.Values{}
        form.Add("method", "token")
        form.Add("token", "myroot")
        request.PostForm = form
        request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
        rt.ServeHTTP(httpWriter, request)
        URL, err := httpWriter.Result().Location()
        So(err, ShouldBeNil)
        So(URL.String(), ShouldEqual, mconfig.Config.GetSubdomainFullUrl("mongo."))
        So(httpWriter.Result().Cookies()[0].Value, ShouldEqual, "myroot")
    })
}



func TestTokenUsername(t *testing.T) {
    mconfig.LoadConfig()
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))
    goth.UseProviders(
        github.New("test", "test-test", "read:org"),
    )
    Convey("test user/pass  method", t, func(c C) {
        rt:=  loginRouters()
        httpWriter := httptest.NewRecorder()
        request := httptest.NewRequest("POST", "/ui/login", nil)
        form := url.Values{}
        form.Add("method", "userpass")
        form.Add("user", "menshend")
        form.Add("password", "test")
        request.PostForm = form
        request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
        rt.ServeHTTP(httpWriter, request)
        URL, err := httpWriter.Result().Location()
        So(err, ShouldBeNil)
        So(URL.String(), ShouldEqual, mconfig.Config.GetServicePath())
        So(httpWriter.Result().Cookies()[0].Value, ShouldNotBeEmpty)
    })
    
}

func TestSameOriginHandler(t *testing.T) {
    mconfig.Config.Github.ClientID = ""
    mconfig.Config.Github.ClientSecret = ""
    mconfig.LoadConfig()
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))
    goth.UseProviders(
        github.New("test", "test-test", "read:org"),
    )
    Convey("should ignore same origin policy with method=get", t, func(c C) {
        rt:=  uiHandler()
        httpWriter := httptest.NewRecorder()
        request := httptest.NewRequest("GET", "/ui/auth/github/callback", nil)
        request.Header.Add("Origin", "http://evil.com")
        rt.ServeHTTP(httpWriter, request)
        URL, err := httpWriter.Result().Location()
        So(err, ShouldBeNil)
        So(URL.String(), ShouldEqual, mconfig.Config.GetLoginPath()+"?loginError=Github")
    })
    
}
