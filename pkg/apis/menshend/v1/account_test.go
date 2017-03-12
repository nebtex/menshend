package v1

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "net/http/httptest"
    "net/http"
    "encoding/json"
    
    "github.com/emicklei/go-restful"
    "io/ioutil"
    . "github.com/nebtex/menshend/pkg/utils"
    . "github.com/nebtex/menshend/pkg/utils/test"
    . "github.com/nebtex/menshend/pkg/users"
    . "github.com/nebtex/menshend/pkg/config"
    . "github.com/nebtex/menshend/pkg/apis/menshend"
    "fmt"
    vault "github.com/hashicorp/vault/api"
    "bytes"
    "github.com/ansel1/merry"
)

func TestAccountStatus(t *testing.T) {
    VaultConfig.Address = "http://127.0.0.1:8200"
    Convey("Should return user status", t, func() {
        
        CleanVault()
        wsContainer := restful.NewContainer()
        ar := AuthResource{}
        ar.Register(wsContainer)
        httpReq, err := http.NewRequest("GET", "/v1/account", nil)
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
    VaultConfig.Address = "http://127.0.0.1:8200"
    Convey("Should logout", t, func() {
        CleanVault()
        //create token
        vc, err := vault.NewClient(VaultConfig)
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
        user, err := NewUser(secret.Auth.ClientToken)
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        httpReq.Header.Add("X-Menshend-Token", user.GenerateJWT())
        httpWriter := httptest.NewRecorder()
        wsContainer.ServeHTTP(httpWriter, httpReq)
        _, err = vc.Auth().Token().Lookup(secret.Auth.ClientToken)
        So(err, ShouldNotBeNil)
    })
    
}

func TestTokenLogin(t *testing.T) {
    VaultConfig.Address = "http://127.0.0.1:8200"
    Convey("Should store the token in the jwt token with the apropiate expiration time", t, func() {
        CleanVault()
        
        wsContainer := restful.NewContainer()
        ar := AuthResource{}
        ar.Register(wsContainer)
        
        body := bytes.NewReader([]byte(`{"authProvider": "token", "data": {"token": "myroot"} }`))
        httpReq, err := http.NewRequest("PUT", "/v1/account", body)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        httpWriter := httptest.NewRecorder()
        wsContainer.ServeHTTP(httpWriter, httpReq)
        So(httpWriter.HeaderMap["X-Menshend-Token"], ShouldNotBeNil)
    })
    
    Convey("Should fail when triying to login with a non " +
        "existen token", t, func(c C) {
        
        defer func() {
            r := recover()
            if (r == nil) {
                t.Error("did not panicked")
                t.Fail()
            }
            switch x := r.(type) {
            case error:
                c.So(merry.Is(x, NotAuthorized), ShouldBeTrue)
            default:
                t.Errorf("%v", x)
                t.Fail()
            }
        }()
        CleanVault()
        
        wsContainer := restful.NewContainer()
        ar := AuthResource{}
        ar.Register(wsContainer)
        
        body := bytes.NewReader([]byte(`{"authProvider": "token", "data": {"token": "nonToken"} }`))
        httpReq, err := http.NewRequest("PUT", "/v1/account", body)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        httpWriter := httptest.NewRecorder()
        wsContainer.ServeHTTP(httpWriter, httpReq)
    })
    
    Convey("Should fail when a invalid json is sent", t, func(c C) {
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
        ar := AuthResource{}
        ar.Register(wsContainer)
        
        body := bytes.NewReader([]byte(`{"authProvider": "token"}`))
        httpReq, err := http.NewRequest("PUT", "/v1/account", body)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        httpWriter := httptest.NewRecorder()
        wsContainer.ServeHTTP(httpWriter, httpReq)
    })
    
}

func Test_UserPasswordHandler(t *testing.T) {
    VaultConfig.Address = "http://127.0.0.1:8200"
    Convey("Should fail when it try to login using " +
        "bad user or password", t, func(c C) {
        defer func() {
            r := recover()
            if (r == nil) {
                t.Error("did not panicked")
                t.Fail()
            }
            switch x := r.(type) {
            case error:
                c.So(merry.Is(x, NotAuthorized), ShouldBeTrue)
            default:
                t.Errorf("%v", x)
                t.Fail()
            }
        }()
        CleanVault()
        
        wsContainer := restful.NewContainer()
        ar := AuthResource{}
        ar.Register(wsContainer)
        
        body := bytes.NewReader([]byte(`{"authProvider": "userpass", "data": {"user": "baduser", "password": "badpassword"} }`))
        httpReq, err := http.NewRequest("PUT", "/v1/account", body)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        httpWriter := httptest.NewRecorder()
        wsContainer.ServeHTTP(httpWriter, httpReq)
    })
    
	 Convey("Should login  and save the vault token in the jwt token with the apropiate" +
		 " expiration time", t, func(c C) {
         CleanVault()
        
         wsContainer := restful.NewContainer()
         ar := AuthResource{}
         ar.Register(wsContainer)
        
         body := bytes.NewReader([]byte(`{"authProvider": "userpass", "data": {"user": "menshend", "password": "test"} }`))
         httpReq, err := http.NewRequest("PUT", "/v1/account", body)
         So(err, ShouldBeNil)
         httpReq.Header.Set("Content-Type", "application/json")
         httpWriter := httptest.NewRecorder()
         wsContainer.ServeHTTP(httpWriter, httpReq)
         So(httpWriter.HeaderMap["X-Menshend-Token"], ShouldNotBeNil)
	 })
    
}

