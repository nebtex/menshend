package strategy

import (
    "testing"
    "net/http"
    "net/http/httptest"
    . "github.com/smartystreets/goconvey/convey"
    "net/url"
)

type testBackend struct {
    url    string
    headers map[string]string
    user    string
    pass    string
}

func (t *testBackend)Login() *url.Userinfo {
    if t.user == "" {
        return nil
    }
    return url.UserPassword(t.user, t.pass)
}

func (t *testBackend)BaseUrl() string {
    return t.url
}

func (t *testBackend)Headers() map[string]string {
    if t.headers == nil {
        t.headers = map[string]string{}
    }
    return t.headers
}

func TestRedirect_Execute(t *testing.T) {
    Convey("Should redirect to backend", t, func() {
        tb := &testBackend{url:"https://google.com:3000",
            headers:map[string]string{"test":"true"},
            user:"criloz", pass:"criloz"}
        r := Redirect{}
        httpReq, err := http.NewRequest("GET", "http://google.menshend.local/search?q=google&x=mars", nil)
        So(err, ShouldBeNil)
        httpWriter := httptest.NewRecorder()
        r.Execute(tb)(httpWriter, httpReq)
        header := httpWriter.Header()
        So(header["Test"][0], ShouldEqual, "true")
        So(header["Location"][0], ShouldEqual, "https://google.com:3000/search?q=google&x=mars")
        
    })
    
}
