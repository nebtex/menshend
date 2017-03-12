package resolvers

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    . "github.com/nebtex/menshend/pkg/apis/menshend/v1"
    . "github.com/nebtex/menshend/pkg/apis/menshend"
    . "github.com/nebtex/menshend/pkg/users"
    . "github.com/nebtex/menshend/pkg/utils"
    "github.com/ansel1/merry"
)

func TestCacheResolver_Resolve(t *testing.T) {
    Convey("should cache by user and subdomain", t, func() {
        
        s := &AdminServiceResource{}
        s.SubDomain = "consul."
        s.ProxyCode = `function getBackend (Username, Groups, AuthProvider)
    tt = {}
    tt["user"] = Username
    tt["AuthProvider"] = AuthProvider
    return "http://www.google.com", tt
end`
        u, err := NewUser("test-acl")
        u.SetExpiresAt(GetNow() + 1000)
        So(err, ShouldBeNil)
        u.UsernamePasswordLogin("test-token")
        cr := &CacheResolver{}
        s.ProxyLanguage = "lua"
        s.Cache.Active = true
        s.Cache.TTL = 3600
        cr.Resolve(s, u)
        bi := cr.Resolve(s, u)
        So(bi.BaseUrl(), ShouldEqual, "http://www.google.com")
        So(bi.Headers()["user"], ShouldEqual, "test-token")
        So(bi.Headers()["AuthProvider"], ShouldEqual, "userpass")
    })
    
    Convey("should fails if receive an unkown language type", t, func(c C) {
        defer func() {
            r := recover()
            if (r == nil) {
                t.Error("did not panicked")
                t.Fail()
            }
            switch x := r.(type) {
            case error:
                c.So(merry.Is(x, InternalError), ShouldBeTrue)
            default:
                t.Errorf("%v", x)
                t.Fail()
            }
        }()
        
        s := &AdminServiceResource{}
        u, err := NewUser("test-acl")
        So(err, ShouldBeNil)
        u.UsernamePasswordLogin("test-token")
        cr := &CacheResolver{}
        s.ProxyLanguage = "Unkown Language"
        s.Cache.Active = true
        s.Cache.TTL = 3600
        cr.Resolve(s, u)
        
    })
    
}
