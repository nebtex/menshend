package kuper

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "github.com/ansel1/merry"
)

func generateService() *Service {
    s := &Service{IsActive: true}
    s.LuaScript = `
    function getBackend ()
        return "http://yahoo.com"
    end
    `
    return s
}

func TestService_CreateLuaScript(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    Convey("Should genrate the lua script that will be used for get the full" +
        " backend url", t, func() {
        
        Convey("should use the user script if is defined ", func() {
            s := generateService()
            s.LuaScript = `function getBackend (Username, Groups, AuthProvider)
    return "www.google.com"
end
`
            u, err := NewUser("test-acl")
            u.SetExpiresAt(getNow() + 3600 * 1000)
            So(err, ShouldBeNil)
            u.TokenLogin()
            ls := s.CreateLuaScript(u)
            So(err, ShouldBeNil)
            So(ls, ShouldEqual, `math.randomseed(os.time())
math.random(); math.random(); math.random()

username = ""
groups = {""}
authProvider = "token"

function getBackend (Username, Groups, AuthProvider)
    return "www.google.com"
end
`)
        })
        
        Convey("should return default script if a user-script is not defined",
            func() {
                s := generateService()
                s.LuaScript = ""
                u, err := NewUser("test-acl")
                u.SetExpiresAt(getNow() + 3600 * 1000)
                So(err, ShouldBeNil)
                u.TokenLogin()
                ls := s.CreateLuaScript(u)
                So(err, ShouldBeNil)
                
                So(ls, ShouldEqual, `math.randomseed(os.time())
math.random(); math.random(); math.random()

username = ""
groups = {""}
authProvider = "token"

function getBackend (Username, Groups, AuthProvider)
    return ""
end`)
            })
        
        Convey("should show user and groups if the github provider is used",
            func() {
                s := generateService()
                s.LuaScript = ""
                u, err := NewUser("test-acl")
                u.SetExpiresAt(getNow() + 1000)
                So(err, ShouldBeNil)
                u.GitHubLogin("criloz", "admin", "delos", "umbrella")
                ls := s.CreateLuaScript(u)
                So(err, ShouldBeNil)
                
                So(ls, ShouldEqual, `math.randomseed(os.time())
math.random(); math.random(); math.random()

username = "criloz"
groups = {"admin", "delos", "umbrella"}
authProvider = "github"

function getBackend (Username, Groups, AuthProvider)
    return ""
end`)
            })
    })
}

func TestService_GetBackend(t *testing.T) {
    Convey("Should return Full url backend", t, func() {
        Convey("Should fails  if does not return anything", func() {
            s := &Service{}
            u, err := NewUser("test-acl")
            u.SetExpiresAt(getNow() + 1000)
            So(err, ShouldBeNil)
            u.TokenLogin()
            backend, err := s.GetBackend(u)
            So(err, ShouldNotBeNil)
            So(merry.Is(err, BadBackendUrl), ShouldBeTrue)
            So(backend, ShouldEqual, "")
        })
        Convey("test with 1 backend", func() {
            s := &Service{}
            u, err := NewUser("test-acl")
            u.SetExpiresAt(getNow() + 1000)
            So(err, ShouldBeNil)
            u.TokenLogin()
            s.LuaScript = `
function getBackend (Username, Groups, AuthProvider)
    return "http://google.com"
end`
            backend, err := s.GetBackend(u)
            So(err, ShouldBeNil)
            So(backend, ShouldEqual, "http://google.com")
        })
        
        Convey("Should return error if the lua script fails", func() {
            s := &Service{}
            s.LuaScript = "xx(ss)44"
            u, err := NewUser("test-acl")
            u.SetExpiresAt(getNow() + 1000)
            So(err, ShouldBeNil)
            u.TokenLogin()
            _, err = s.GetBackend(u)
            So(err, ShouldNotBeNil)
            So(merry.Is(err, LuaScriptFailed), ShouldBeTrue)
        })
    })
    
}

func TestGetBackend(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    Convey("Should return Full url backend (using the permissions)", t, func() {
        Convey("Should return error if subdomain is not provided", func() {
            u, err := NewUser("test-acl")
            u.SetExpiresAt( getNow() + 3600 * 1000)
            So(err, ShouldBeNil)
            u.TokenLogin()
            bc, err := GetBackend(u, "", "machine-learning-public")
            So(err, ShouldNotBeNil)
            So(merry.Is(err, InvalidSubdomain), ShouldBeTrue)
            So(bc, ShouldBeNil)
        })
        
        Convey("If the user has not access to the service it should return " +
            "a permission error", func() {
            u, err := NewUser("test-acl")
            u.SetExpiresAt(getNow() + 3600 * 1000)
            So(err, ShouldBeNil)
            u.TokenLogin()
            bc, err := GetBackend(u, "term", "ml-team")
            So(err, ShouldNotBeNil)
            So(merry.Is(err, PermissionError), ShouldBeTrue)
            So(bc, ShouldBeNil)
        })
        
        Convey("Should panic if it cant connect to vault", func(c C) {
            defer func() {
                r := recover();
                c.So(r, ShouldNotBeNil)
            }()
            var err error
            VaultConfig.Address = "http://example.com.local:3212"
            u, err := NewUser("test-acl")
            u.SetExpiresAt(getNow() + 3600 * 1000)
            So(err, ShouldBeNil)
            u.TokenLogin()
            GetBackend(u, "term", "ml-team")
        })
        
        Convey("Should return service not found, when the subdomain " +
            "don't exist on vault", func(c C) {
            VaultConfig.Address = "http://localhost:8200"
            
            cleanVault()
            
            var err error
            u, err := NewUser("myroot")
            u.SetExpiresAt(getNow() + 1000)
            So(err, ShouldBeNil)
            u.TokenLogin()
            bc, err := GetBackend(u, "term", "ml-team")
            So(err, ShouldNotBeNil)
            So(merry.Is(err, ServiceNotFound), ShouldBeTrue)
            So(bc, ShouldBeNil)
            
        })
        
        Convey("Should pick a service", func(c C) {
            var err error
            cleanVault()
            populateVault()
            So(err, ShouldBeNil)
            u, err := NewUser("myroot")
            u.SetExpiresAt( getNow() + 3600 * 1000)
            So(err, ShouldBeNil)
            u.TokenLogin()
            bc, err := GetBackend(u, "redis", "admin")
            So(err, ShouldBeNil)
            So(bc.Url.Host, ShouldBeIn, []string{"redis.kv"})
        })
        
        Convey("Should return error if the backend url is not valid",
            func(c C) {
                var err error
                cleanVault()
                populateVault()
                u, err := NewUser("myroot")
                u.SetExpiresAt(getNow() + 3600 * 1000)
                So(err, ShouldBeNil)
                u.TokenLogin()
                bc, err := GetBackend(u, "kubernetes", "admin")
                So(err, ShouldNotBeNil)
                So(bc, ShouldBeNil)
                So(merry.Is(err, BadBackendUrl), ShouldBeTrue)
                
            })
        
        Convey("If service is not active should return error", func(c C) {
            var err error
            cleanVault()
            populateVault()
    
            u, err := NewUser("myroot")
            u.SetExpiresAt(getNow() + 3600 * 1000)
            So(err, ShouldBeNil)
            u.TokenLogin()
            bc, err := GetBackend(u, "gitlab", "ml-team")
            So(err, ShouldNotBeNil)
            So(bc, ShouldBeNil)
            So(merry.Is(err, InactiveService), ShouldBeTrue)
            
        })
        
    })
    
}

