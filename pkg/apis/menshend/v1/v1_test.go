package v1

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "net/http/httptest"
    "net/http"
    "os"
    vault "github.com/hashicorp/vault/api"
    mutils "github.com/nebtex/menshend/pkg/utils"
    "encoding/json"
    "github.com/nebtex/menshend/pkg/config"
    "github.com/nebtex/menshend/pkg/strategy"
    "bytes"
    "github.com/nebtex/menshend/pkg/resolvers"
)

func TestCSRFHandler(t *testing.T) {
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))
    config.Config.Uris.BaseUrl = "http://example.com"
    config.Config.Uris.MenshendSubdomain = "lab."
    Convey("should always set the csrf cookie and header on get requests", t, func(c C) {
        Convey("Browser request", func(c C) {
            handler := APIHandler()
            httpReq, err := http.NewRequest("GET", "/v1/account", nil)
            ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: "myroot"}
            cm := &http.Cookie{Path: "/", Name: "md-role", Value: "admin"}
            httpReq.Header.Set("Content-Type", "application/json")
            
            httpReq.AddCookie(ct)
            httpReq.AddCookie(cm)
            So(err, ShouldBeNil)
            httpWriter := httptest.NewRecorder()
            handler.ServeHTTP(httpWriter, httpReq)
            So(len(httpWriter.Result().Cookies()), ShouldEqual, 1)
            So(httpWriter.Result().Header.Get("X-Next-CSRF-Token"), ShouldNotBeEmpty)
        })
        Convey("Non Browser request", func(c C) {
            handler := APIHandler()
            httpReq, err := http.NewRequest("GET", "/v1/account", nil)
            httpReq.Header.Set("X-Vault-Token", "myroot")
            So(err, ShouldBeNil)
            httpWriter := httptest.NewRecorder()
            handler.ServeHTTP(httpWriter, httpReq)
            So(len(httpWriter.Result().Cookies()), ShouldEqual, 1)
            So(httpWriter.Result().Header.Get("X-Next-CSRF-Token"), ShouldNotBeEmpty)
            
        })
    })
    
    Convey("should protect mutation only on browsers", t, func(c C) {
        Convey("Browser request", func(c C) {
            handler := APIHandler()
            httpReq, err := http.NewRequest("POST", "/v1/AdminServiceResource", nil)
            ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: "myroot"}
            cm := &http.Cookie{Path: "/", Name: "md-role", Value: "admin"}
            httpReq.AddCookie(ct)
            httpReq.AddCookie(cm)
            So(err, ShouldBeNil)
            httpWriter := httptest.NewRecorder()
            handler.ServeHTTP(httpWriter, httpReq)
            So(httpWriter.Result().StatusCode, ShouldEqual, 403)
            
        })
        Convey("Non Browser request ", func(c C) {
            handler := APIHandler()
            postBody, err := json.Marshal(&AdminServiceResource{
                Meta:&ServiceMetadata{Name:"gitlab"},
                Resolver:&ServiceResolver{Yaml: &resolvers.YAMLResolver{}},
                Strategy:&ServiceStrategy{Proxy: &strategy.Proxy{}},
            })
            So(err, ShouldBeNil)
            httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", bytes.NewReader(postBody))
            httpReq.Header.Set("X-Vault-Token", "myroot")
            httpReq.Header.Set("Content-Type", "application/json")
            
            So(err, ShouldBeNil)
            httpWriter := httptest.NewRecorder()
            handler.ServeHTTP(httpWriter, httpReq)
            So(httpWriter.Result().StatusCode, ShouldEqual, 200)
            
        })
    })
    
}

