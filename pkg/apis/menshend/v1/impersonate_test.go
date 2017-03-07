package v1

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "net/http/httptest"
    "github.com/emicklei/go-restful"
    "net/http"
    . "github.com/nebtex/menshend/pkg/utils/test"
    . "github.com/nebtex/menshend/pkg/config"
    . "github.com/nebtex/menshend/pkg/users"
    . "github.com/nebtex/menshend/pkg/utils"
    
    "encoding/json"
    "bytes"
    "io/ioutil"
)

func Test_PutImpersonateHandler(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    Convey("should impersonate another user", t, func(c C) {
        CleanVault()
        PopulateVault()
        ip:=&ImpersonateResource{}
        ip.User = StringPtr("anotherUser")
        ip.AuthProvider = AuthProviderPtr(UsernamePasswordProvider)
        postBody, err := json.Marshal(ip)
        So(err, ShouldBeNil)
        httpReq, err := http.NewRequest("PUT", "/v1/impersonate", bytes.NewReader(postBody))
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.UsernamePasswordLogin("root")
        user.SetExpiresAt(GetNow() + 3600)
        httpReq.Header.Add("X-Menshend-Token", user.GenerateJWT())
        httpWriter := httptest.NewRecorder()
        wsContainer := restful.NewContainer()
        s := ImpersonateResource{}
        s.Register(wsContainer)
        wsContainer.ServeHTTP(httpWriter, httpReq)
        token := httpWriter.Header().Get("X-Menshend-Token")
        neUser, err:= FromJWT(token)
        So(err, ShouldBeNil)
        So(neUser.Menshend.Username, ShouldEqual, "anotherUser")
    
    })
    
}
func Test_ReadImpersonateHandler(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    Convey("should be inactive when there is not an impersonification goings on", t, func(c C) {
        CleanVault()
        PopulateVault()
        ip:=&ImpersonateResource{}
        ip.User = StringPtr("anotherUser")
        ip.AuthProvider = AuthProviderPtr(UsernamePasswordProvider)
        httpReq, err := http.NewRequest("GET", "/v1/impersonate", nil)
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.UsernamePasswordLogin("root")
        user.SetExpiresAt(GetNow() + 3600)
        httpReq.Header.Add("X-Menshend-Token", user.GenerateJWT())
        httpWriter := httptest.NewRecorder()
        wsContainer := restful.NewContainer()
        s := ImpersonateResource{}
        s.Register(wsContainer)
        wsContainer.ServeHTTP(httpWriter, httpReq)
        jsres, err := ioutil.ReadAll(httpWriter.Body)
        So(err, ShouldBeNil)
        ret := &ImpersonateResource{}
        err = json.Unmarshal(jsres, ret)
        So(err, ShouldBeNil)
        So(ret.Active, ShouldBeFalse)
        
    })
    
    Convey("should be active when there is an impersonification goings on", t, func(c C) {
        CleanVault()
        PopulateVault()
        ip:=&ImpersonateResource{}
        ip.User = StringPtr("anotherUser")
        ip.AuthProvider = AuthProviderPtr(UsernamePasswordProvider)
        httpReq, err := http.NewRequest("GET", "/v1/impersonate", nil)
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.UsernamePasswordLogin("root")
        user.SetExpiresAt(GetNow() + 3600)
        user.Impersonate(AuthProviderPtr(UsernamePasswordProvider), StringPtr("other"))
        httpReq.Header.Add("X-Menshend-Token", user.GenerateJWT())
        httpWriter := httptest.NewRecorder()
        wsContainer := restful.NewContainer()
        s := ImpersonateResource{}
        s.Register(wsContainer)
        wsContainer.ServeHTTP(httpWriter, httpReq)
        jsres, err := ioutil.ReadAll(httpWriter.Body)
        So(err, ShouldBeNil)
        ret := &ImpersonateResource{}
        err = json.Unmarshal(jsres, ret)
        So(err, ShouldBeNil)
        So(ret.Active, ShouldBeTrue)
        
    })
    
}
func Test_StopImpersonateHandler(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    Convey("should stop the impersonification", t, func(c C) {
        CleanVault()
        PopulateVault()
        ip:=&ImpersonateResource{}
        ip.User = StringPtr("anotherUser")
        ip.AuthProvider = AuthProviderPtr(UsernamePasswordProvider)
        httpReq, err := http.NewRequest("DELETE", "/v1/impersonate", nil)
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.UsernamePasswordLogin("root")
        user.SetExpiresAt(GetNow() + 3600)
        user.Impersonate(AuthProviderPtr(UsernamePasswordProvider), StringPtr("other"))
        httpReq.Header.Add("X-Menshend-Token", user.GenerateJWT())
        httpWriter := httptest.NewRecorder()
        wsContainer := restful.NewContainer()
        s := ImpersonateResource{}
        s.Register(wsContainer)
        wsContainer.ServeHTTP(httpWriter, httpReq)
        token := httpWriter.Header().Get("X-Menshend-Token")
        neUser, err:= FromJWT(token)
        So(err, ShouldBeNil)
        So(neUser.Menshend.Username, ShouldEqual, "root")
        
    })
    
}
