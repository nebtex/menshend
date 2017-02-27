package kuper

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "io/ioutil"
    "net/http/httptest"
    "net/http"
    "bytes"
    "encoding/json"
    
)

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
        jsonResponse, err := ioutil.ReadAll(response.Body)
        ar := &Response{}
        err = json.Unmarshal(jsonResponse, ar)
        So(err, ShouldBeNil)
        So(ar.Success, ShouldBeTrue)
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
        jsonResponse, err := ioutil.ReadAll(response.Body)
        ar := &Response{}
        err = json.Unmarshal(jsonResponse, ar)
        So(err, ShouldBeNil)
        So(ar.Success, ShouldBeFalse)
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
        jsonResponse, err := ioutil.ReadAll(response.Body)
        ar := &Response{}
        err = json.Unmarshal(jsonResponse, ar)
        So(err, ShouldBeNil)
        So(ar.Success, ShouldBeFalse)
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
        jsonResponse, err := ioutil.ReadAll(response.Body)
        ar := &Response{}
        err = json.Unmarshal(jsonResponse, ar)
        So(err, ShouldBeNil)
        So(ar.Success, ShouldBeFalse)
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
        jsonResponse, err := ioutil.ReadAll(response.Body)
        ar := &Response{}
        err = json.Unmarshal(jsonResponse, ar)
        So(err, ShouldBeNil)
        So(ar.Success, ShouldBeFalse)
    })
    Convey("Should login store the vault token in the jwt token with the apropiate" +
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
        tp := UPLogin{User:"kuper",
            Password:"test"}
        postBody, err := json.Marshal(tp)
        So(err, ShouldBeNil)
        req, err := http.NewRequest("POST", u.String(),
            bytes.NewReader(postBody))
        So(err, ShouldBeNil)
        client := &http.Client{}
        response, err := client.Do(req)
        So(err, ShouldBeNil)
        jsonResponse, err := ioutil.ReadAll(response.Body)
        ar := &Response{}
        err = json.Unmarshal(jsonResponse, ar)
        So(err, ShouldBeNil)
        So(ar.Success, ShouldBeTrue)
    })
    
    
    
    
}

