package strategy

import (
    "testing"
    "net/http"
    "net/http/httptest"
    . "github.com/smartystreets/goconvey/convey"
    "io/ioutil"
    "encoding/json"
    "github.com/nebtex/menshend/pkg/resolvers"
    . "github.com/nebtex/menshend/pkg/utils"
    vault "github.com/hashicorp/vault/api"
    "github.com/ansel1/merry"
)

func TestProxy_Execute(t *testing.T) {
    Convey("Should proxy http/https", t, func() {
        tb := &resolvers.YAMLResolver{}
        tb.Content = `baseUrl: http://localhost:8200
headersMap:
  X-Vault-Token: myroot
  h2: t2`
        r := &Proxy{}
        httpReq, err := http.NewRequest("GET", "http://vault.menshend.local/v1/sys/seal-status", nil)
        So(err, ShouldBeNil)
        httpWriter := httptest.NewRecorder()
        noBrowserHandler(r.Execute(tb, &vault.Secret{})).ServeHTTP(httpWriter, httpReq)
        So(httpReq.Header.Get("X-Vault-Token"), ShouldEqual, "myroot")
        d, err := ioutil.ReadAll(httpWriter.Body)
        So(err, ShouldBeNil)
        result := map[string]interface{}{}
        err = json.Unmarshal(d, &result)
        So(err, ShouldBeNil)
        So(result["sealed"], ShouldBeFalse)
    })
    
    Convey("Should panic when backend is not online", t, func(c C) {
        defer func() {
            r := recover()
            if (r == nil) {
                t.Error("did not panicked")
                t.Fail()
            }
            switch x := r.(type) {
            case error:
                c.So(merry.Is(x, BadGateway), ShouldBeTrue)
            default:
                t.Errorf("%v", x)
                t.Fail()
            }
        }()
        tb := &resolvers.YAMLResolver{}
        tb.Content = `baseUrl: http://example.local:444
headersMap:
  X-Vault-Token: myroot
  h2: t2`
        
        r := &Proxy{}
        httpReq, err := http.NewRequest("GET", "http://vault.menshend.local/v1/sys/seal-status", nil)
        So(err, ShouldBeNil)
        httpWriter := httptest.NewRecorder()
        noBrowserHandler(r.Execute(tb, &vault.Secret{})).ServeHTTP(httpWriter, httpReq)
    })
}
