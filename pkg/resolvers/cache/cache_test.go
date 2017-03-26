package resolvers

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    . "github.com/nebtex/menshend/pkg/apis/menshend/v1"
    . "github.com/nebtex/menshend/pkg/utils"
    "github.com/nebtex/menshend/pkg/resolvers"
    vault "github.com/hashicorp/vault/api"
    "github.com/nebtex/menshend/pkg/config"
)

func TestCacheResolver_Resolve(t *testing.T) {
    config.VaultConfig.Address = "http://127.0.0.1:8200"
    vc, err := vault.NewClient(config.VaultConfig)
    CheckPanic(err)
    vc.SetToken("myroot")
    Convey("should cache by user and subdomain", t, func() {
        
        s := &AdminServiceResource{}
        s.Meta = &ServiceMetadata{}
        s.Meta.SubDomain = "consul."
        s.Meta.RoleID = "ml-team"
        s.Resolver = &ServiceResolver{}
        s.Resolver.Lua = &resolvers.LuaResolver{}
        s.Resolver.Lua.Content = `function getBackend (tokenInfo, body)
            tt = {}
            tt["BaseUrl"] = "http://www.google.com"
            tt["HeaderMap"] = {}
            tt["HeaderMap"]["X-User"] = tokenInfo.Data.display_name
            return tt
        end`
        s.Cache = &ServiceCache{TTL: 3600}
        tokenInfo, err := vc.Auth().Token().LookupSelf()
        So(err, ShouldBeNil)
        cr := NewCacheResolver(s)
        cr.Resolve(tokenInfo)
        bi := cr.Resolve(tokenInfo)
        So(bi.BaseUrl(), ShouldEqual, "http://www.google.com")
    })
  
}
