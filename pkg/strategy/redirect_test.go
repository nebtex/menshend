package strategy

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "net/url"
    "github.com/nebtex/menshend/pkg/resolvers"
    "net/http/httptest"
    "net/http"
    
    vault "github.com/hashicorp/vault/api"
)

type testBackend struct {
    url     string
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
        
        tb := &resolvers.YAMLResolve{}
        tb.Content = `baseUrl: http://google.db:27072
headersMap:
  h1: t1
  h2: t2`
        
        r := Redirect{}
        httpReq, err := http.NewRequest("GET", "http://google.menshend.local/search?q=google&x=mars", nil)
        So(err, ShouldBeNil)
        httpWriter := httptest.NewRecorder()
        r.Execute(tb, &vault.Secret{}).ServeHTTP(httpWriter, httpReq)
        header := httpWriter.Header()
        So(header.Get("h1"), ShouldEqual, "t1")
        So(header.Get("Location"), ShouldEqual, "http://google.db:27072/search?q=google&x=mars")
        
    })
    
}
