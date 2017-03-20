package resolvers

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    vault "github.com/hashicorp/vault/api"
)

func TestYAMLResolve_Resolve(t *testing.T) {
    Convey("should return backend", t, func() {
        Yaml := &YAMLResolver{}
        Yaml.Content = `baseUrl: http://google.db:27072
headersMap:
  h1: t1
  h2: t2
        `
        bi := Yaml.Resolve(&vault.Secret{})
        So(bi.BaseUrl(), ShouldEqual, "http://google.db:27072")
        So(bi.Headers()["h1"], ShouldEqual, "t1")
        So(bi.Headers()["h2"], ShouldEqual, "t2")
        So(bi.Error().Error(), ShouldEqual, "")
        So(bi.Passed(), ShouldEqual, true)
    
    })
    
}
