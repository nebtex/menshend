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
    
    "fmt"
    vault "github.com/hashicorp/vault/api"
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
        _,err = vc.Auth().Token().Lookup(secret.Auth.ClientToken)
        So(err, ShouldNotBeNil)
    })
    
}

/*
func Test_TokenLoginHandler(t *testing.T) {
    VaultConfig.Address = "http://127.0.0.1:8200"
    
    Convey("Should store the token in the jwt token with the apropiate" +
        " expiration time", t, func() {
        var u bytes.Buffer
        
        testHandler := func() http.HandlerFunc {
            return http.HandlerFunc(TokenLoginHandler)
        }
        
        ts := httptest.NewServer(testHandler())
        defer ts.Close()
        u.WriteString(string(ts.URL))
        u.WriteString("/api/v1/login/token")
        tp := TokenLogin{Token:"myroot"}
        postBody, err := json.Marshal(tp)
        So(err, ShouldBeNil)
        req, err := http.NewRequest("POST", u.String(),
            bytes.NewReader(postBody))
        So(err, ShouldBeNil)
        client := &http.Client{}
        response, err := client.Do(req)
        So(err, ShouldBeNil)
        url, err:= response.Location()
        So(err, ShouldBeNil)
        So(url.String(), ShouldEqual, Config.GetServicePath())
    })
    
    Convey("Should fail when triying to login with a non " +
        "existen token", t, func() {
        var u bytes.Buffer
        
        testHandler := func() http.HandlerFunc {
            return http.HandlerFunc(TokenLoginHandler)
        }
        
        ts := httptest.NewServer(testHandler())
        defer ts.Close()
        u.WriteString(string(ts.URL))
        u.WriteString("/api/v1/login/token")
        tp := TokenLogin{Token:"404token"}
        postBody, err := json.Marshal(tp)
        So(err, ShouldBeNil)
        req, err := http.NewRequest("POST", u.String(),
            bytes.NewReader(postBody))
        So(err, ShouldBeNil)
        client := &http.Client{}
        response, err := client.Do(req)
        So(err, ShouldBeNil)
        url, err:= response.Location()
        So(err, ShouldBeNil)
        So(url.RawQuery, ShouldEqual, "token_error=true")
        
    })
    
    Convey("Should fail when a invalid json is sent", t, func() {
        var u bytes.Buffer
        
        testHandler := func() http.HandlerFunc {
            return http.HandlerFunc(TokenLoginHandler)
        }
        
        ts := httptest.NewServer(testHandler())
        defer ts.Close()
        u.WriteString(string(ts.URL))
        u.WriteString("/api/v1/login/token")
        req, err := http.NewRequest("POST", u.String(),
            nil)
        So(err, ShouldBeNil)
        client := &http.Client{}
        response, err := client.Do(req)
        So(err, ShouldBeNil)
        url, err:= response.Location()
        So(err, ShouldBeNil)
        So(url.RawQuery, ShouldEqual, "token_error=true")
    })
    
}


func Test_UserPasswordHandler(t *testing.T) {
    VaultConfig.Address = "http://127.0.0.1:8200"
    
    Convey("Should fail when a invalid json is sent", t, func() {
        var u bytes.Buffer
        
        testHandler := func() http.HandlerFunc {
            return http.HandlerFunc(UserPasswordHandler)
        }
        
        ts := httptest.NewServer(testHandler())
        defer ts.Close()
        u.WriteString(string(ts.URL))
        u.WriteString("/api/v1/login/userpass")
        req, err := http.NewRequest("POST", u.String(),
            nil)
        So(err, ShouldBeNil)
        client := &http.Client{}
        response, err := client.Do(req)
        So(err, ShouldBeNil)
        url, err:= response.Location()
        So(err, ShouldBeNil)
        So(url.RawQuery, ShouldEqual, "user_pass_error=true")
    })
    
    Convey("Should fail when it try to login using " +
        "bad user or password", t, func() {
        var u bytes.Buffer
        
        testHandler := func() http.HandlerFunc {
            return http.HandlerFunc(UserPasswordHandler)
        }
        
        ts := httptest.NewServer(testHandler())
        defer ts.Close()
        u.WriteString(string(ts.URL))
        u.WriteString("/api/v1/login/token")
        tp := UPLogin{User:"404token",
            Password:"404token"}
        postBody, err := json.Marshal(tp)
        So(err, ShouldBeNil)
        req, err := http.NewRequest("POST", u.String(),
            bytes.NewReader(postBody))
        So(err, ShouldBeNil)
        client := &http.Client{}
        response, err := client.Do(req)
        So(err, ShouldBeNil)
        url, err:= response.Location()
        So(err, ShouldBeNil)
        So(url.RawQuery, ShouldEqual, "user_pass_error=true")
    })
    Convey("Should login  and store the vault token in the jwt token with the apropiate" +
        " expiration time", t, func() {
        var u bytes.Buffer
        
        testHandler := func() http.HandlerFunc {
            return http.HandlerFunc(UserPasswordHandler)
        }
        
        ts := httptest.NewServer(testHandler())
        defer ts.Close()
        defer ts.Close()
        u.WriteString(string(ts.URL))
        u.WriteString("/api/v1/login/token")
        tp := UPLogin{User:"menshend",
            Password:"test"}
        postBody, err := json.Marshal(tp)
        So(err, ShouldBeNil)
        req, err := http.NewRequest("POST", u.String(),
            bytes.NewReader(postBody))
        So(err, ShouldBeNil)
        client := &http.Client{}
        response, err := client.Do(req)
        So(err, ShouldBeNil)
        url, err:= response.Location()
        So(err, ShouldBeNil)
        So(url.String(), ShouldEqual, Config.GetServicePath())
    })
    
}*/

