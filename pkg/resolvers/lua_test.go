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

func TestService_CreateLuaScript(t *testing.T) {
    Convey("Should genrate the lua script that will be used for get the full" +
        " backend url", t, func() {
        
        Convey("should use the user script if is defined ", func() {
            s := &AdminServiceResource{}
            s.ProxyCode = `function getBackend (Username, Groups, AuthProvider)
    return "www.google.com", {}
end`
            u, err := NewUser("xxx")
            u.SetExpiresAt(GetNow() + 3600 * 1000)
            u.TokenLogin()
            So(err, ShouldBeNil)
            ls := CreateLuaScript(s, u)
            So(ls, ShouldEqual, `math.randomseed(os.time())
math.random(); math.random(); math.random()

username = ""
groups = {""}
authProvider = "token"

function getBackend (Username, Groups, AuthProvider)
    return "www.google.com", {}
end`)
        })
        
        Convey("should show user and groups if the github provider is used",
            func() {
                s := &AdminServiceResource{}
                s.ProxyCode = `function getBackend (Username, Groups, AuthProvider)
    return "www.google.com", {}
end`
                u, err := NewUser("test-acl")
                u.SetExpiresAt(GetNow() + 1000)
                So(err, ShouldBeNil)
                u.GitHubLogin("criloz", "admin", "delos", "umbrella")
                ls := CreateLuaScript(s, u)
                So(err, ShouldBeNil)
                
                So(ls, ShouldEqual, `math.randomseed(os.time())
math.random(); math.random(); math.random()

username = "criloz"
groups = {"admin", "delos", "umbrella"}
authProvider = "github"

function getBackend (Username, Groups, AuthProvider)
    return "www.google.com", {}
end`)
            })
    })
}

func TestService_GetBackend(t *testing.T) {
    Convey("Should return Full url backend", t, func() {
        
        Convey("Should return url and headers", func() {
            s := &AdminServiceResource{}
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
            bi := cr.Resolve(s, u)
            So(bi.BaseUrl(), ShouldEqual, "http://www.google.com")
            So(bi.Headers()["user"], ShouldEqual, "test-token")
            So(bi.Headers()["AuthProvider"], ShouldEqual, "userpass")
            
        })
        Convey("Should return error if the lua script fails", func(c C) {
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
            s.ProxyCode = `bad line of code`
            
            u, err := NewUser("test-acl")
            u.SetExpiresAt(GetNow() + 1000)
            So(err, ShouldBeNil)
            u.UsernamePasswordLogin("test-token")
            cr := &CacheResolver{}
            s.ProxyLanguage = "lua"
            cr.Resolve(s, u)
        })
        
        Convey("Should return error if the backend url is not valid",
            func(c C) {
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
                s.ProxyCode = `function getBackend (Username, Groups, AuthProvider)
    tt = {}
    tt["user"] = Username
    tt["AuthProvider"] = AuthProvider
    return "&www.$google.com", tt
end`
    
                u, err := NewUser("test-acl")
                u.SetExpiresAt(GetNow() + 1000)
                So(err, ShouldBeNil)
                u.UsernamePasswordLogin("test-token")
                cr := &CacheResolver{}
                s.ProxyLanguage = "lua"
                cr.Resolve(s, u)

                
            })
        
    })
}


