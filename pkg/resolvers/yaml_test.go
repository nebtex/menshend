package resolvers

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    
    . "github.com/nebtex/menshend/pkg/users"
    "github.com/nebtex/menshend/pkg/apis/menshend/v1"
)

func TestYAMLResolve_Resolve(t *testing.T) {
    Convey("should return backend", t, func() {
        c := &v1.AdminServiceResource{}
        c.ProxyLanguage = "yaml"
        c.ProxyCode = `baseUrl: http://google.db:27072
headersMap:
  h1: t1
  h2: t2
        `
        u, err := NewUser("xxx")
        So(err, ShouldBeNil)
        yr := CacheResolver{}
        bi := yr.Resolve(c, u)
        So(bi.BaseUrl(), ShouldEqual, "http://google.db:27072")
        So(bi.Headers()["h1"], ShouldEqual, "t1")
        So(bi.Headers()["h2"], ShouldEqual, "t2")
    })
    
}
