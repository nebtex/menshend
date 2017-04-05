package resolvers

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    vault "github.com/hashicorp/vault/api"
    mutils "github.com/nebtex/menshend/pkg/utils"
    "github.com/ansel1/merry"
    "os"
)

func TestService_GetBackend(t *testing.T) {
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))
    vc, err := vault.NewClient(vault.DefaultConfig())
    mutils.CheckPanic(err)
    vc.SetToken("myroot")
    Convey("Should return Full url backend", t, func() {
        
        Convey("Should return url and headers", func() {
            lr := &LuaResolver{}
            lr.Content = `

function getBackend (tokenInfo, body)
    tt = {}
    tt["BaseUrl"] = "http://www.google.com"
    tt["HeaderMap"] = {}
    tt["HeaderMap"]["X-User"] = tokenInfo.Data.display_name
    return tt
end`
            
            tokenInfo, err := vc.Auth().Token().LookupSelf()
            So(err, ShouldBeNil)
            bi := lr.Resolve(tokenInfo)
            So(bi.BaseUrl(), ShouldEqual, "http://www.google.com")
            So(bi.Headers()["X-User"], ShouldEqual, "token")
            So(bi.Passed(), ShouldEqual, true)
            
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
                    c.So(merry.Is(x, mutils.InternalError), ShouldBeTrue)
                default:
                    t.Errorf("%v", x)
                    t.Fail()
                }
            }()
            
            tokenInfo, err := vc.Auth().Token().LookupSelf()
            So(err, ShouldBeNil)
            lr := &LuaResolver{}
            lr.Content = `bad line of code`
            lr.Resolve(tokenInfo)
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
                        c.So(merry.Is(x, mutils.InternalError), ShouldBeTrue)
                    default:
                        t.Errorf("%v", x)
                        t.Fail()
                    }
                }()
                s := &LuaResolver{}
                s.Content = `function getBackend (Username, Groups, AuthProvider)
    tt = {}
    tt["user"] = Username
    tt["AuthProvider"] = AuthProvider
    return "&www.$google.com", tt
end`
                tokenInfo, err := vc.Auth().Token().LookupSelf()
                So(err, ShouldBeNil)
                So(err, ShouldBeNil)
                s.Resolve(tokenInfo)
                
            })
        Convey("Should be able to read the body if this is set", func(c C) {
            
            lr := &LuaResolver{}
            lr.Content = `

function getBackend (tokenInfo, request)
    local json = require("json")
    tt = {}
    tt["BaseUrl"] = "http://www.google.com"
    tt["HeaderMap"] = {}
    tt["Passed"] = false
    tt["HeaderMap"]["X-Operation"] = json.decode(request.Body).operation
    return tt
end`
            
            tokenInfo, err := vc.Auth().Token().LookupSelf()
            So(err, ShouldBeNil)
            lr.SetRequest("GET", `{"operation": "post"}`)
            bi := lr.Resolve(tokenInfo)
            So(bi.Headers()["X-Operation"], ShouldEqual, "post")
            So(bi.Passed(), ShouldEqual, false)
    
        })
        
    })
}


