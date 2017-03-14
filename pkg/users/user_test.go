package users

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "github.com/dgrijalva/jwt-go"
    . "github.com/nebtex/menshend/pkg/config"
    . "github.com/nebtex/menshend/pkg/utils"
)

//TestGetUser
func TestGetUser(t *testing.T) {
    Convey("Should obtain an User struct from a valid jwt token", t, func() {
        u, err := NewUser("test-acl")
        u.SetExpiresAt(GetNow() + 1000)
        So(err, ShouldBeNil)
        u.GitHubLogin("criloz", "delos", "umbrella")
        j := u.GenerateJWT()
        requestUser, err := FromJWT(j)
        So(err, ShouldBeNil)
        So(requestUser.Valid(), ShouldBeNil)
        So(requestUser.Menshend.Username, ShouldEqual, "criloz")
        
        Convey("If the token use another algo should return error", func() {
            u, err := NewUser("test-acl")
            u.SetExpiresAt(GetNow() + 1000)
            So(err, ShouldBeNil)
            u.GitHubLogin("criloz", "delos", "umbrella")
            token := jwt.NewWithClaims(jwt.SigningMethodNone, u)
            j, nErr := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
            So(nErr, ShouldBeNil)
            _, err = FromJWT(j)
            So(err, ShouldNotBeNil)
            
        })
        
        Convey("should ivalidate token generated with other secret key",
            func() {
                Config.HashKey = GenerateRandomString(32)
                u, err := NewUser("test-acl")
                u.SetExpiresAt(GetNow() + 1000)
                So(err, ShouldBeNil)
                u.GitHubLogin("criloz", "delos", "umbrella")
                j := u.GenerateJWT()
                Config.HashKey = GenerateRandomString(32)
                _, err = FromJWT(j)
                So(err, ShouldNotBeNil)
            })
        
    })
    
    Convey("Should mark the token as invalid after the expiration date", t,
        func() {
            Config.HashKey = GenerateRandomString(32)
            u, err := NewUser("test-acl")
            u.SetExpiresAt(GetNow() - 1000)
            So(err, ShouldBeNil)
            u.TokenLogin()
            j := u.GenerateJWT()
            _, err = FromJWT(j)
            So(err, ShouldNotBeNil)
        })
    
    Convey("Should mark the token as invalid if username is not defined " +
        "and someone is using the impersonate feature", t,
        func() {
            Config.HashKey = GenerateRandomString(64)
            u, err := NewUser("test-acl")
            u.SetExpiresAt(GetNow() + 1000)
            So(err, ShouldBeNil)
            u.TokenLogin()
            u.Menshend.ImpersonateBy = &JwtImpersonateInfo{}
            u.Menshend.ImpersonateBy.Username = "criloz"
            j := u.GenerateJWT()
            _, err = FromJWT(j)
            So(err, ShouldNotBeNil)
        })
    
    Convey("Should mark the token as invalid if it has not a vault token", t,
        func() {
            Config.HashKey = GenerateRandomString(32)
            u, err := NewUser("")
            u.SetExpiresAt(GetNow() + 1000)
            So(err, ShouldBeNil)
            u.UsernamePasswordLogin("criloz")
            j := u.GenerateJWT()
            _, err = FromJWT(j)
            So(err, ShouldNotBeNil)
            
        })
    
    Convey("Should not contains user or groups when TokenLogin is used", t,
        func() {
            Config.HashKey = GenerateRandomString(64)
            u, err := NewUser("acl-token")
            u.SetExpiresAt(GetNow() + 1000)
            So(err, ShouldBeNil)
            u.TokenLogin()
            u.Menshend.Username = "criloz"
            u.Menshend.Groups = []string{"admin", "devs"}
            j := u.GenerateJWT()
            _, err = FromJWT(j)
            So(err, ShouldNotBeNil)
        })
    
    Convey("Should not contains groups when UsernamePasswordLogin is used", t,
        func() {
            Config.HashKey = GenerateRandomString(64)
            u, err := NewUser("acl-token")
            u.SetExpiresAt(GetNow() + 1000)
            So(err, ShouldBeNil)
            u.UsernamePasswordLogin("criloz")
            u.Menshend.Groups = []string{"admin", "devs"}
            j := u.GenerateJWT()
            _, err = FromJWT(j)
            So(err, ShouldNotBeNil)
        })
    
    Convey("Should contains a valid AuthProvider", t, func() {
        Config.HashKey = GenerateRandomString(32)
        u, err := NewUser("acl-token")
        u.SetExpiresAt(GetNow() + 1000)
        So(err, ShouldBeNil)
        u.UsernamePasswordLogin("criloz")
        u.Menshend.AuthProvider = "dsfedgergwerg"
        j := u.GenerateJWT()
        _, err = FromJWT(j)
        So(err, ShouldNotBeNil)
    })
    
    Convey("userpass auth should contains an username", t, func() {
        Config.HashKey = GenerateRandomString(32)
        u, err := NewUser("acl-token")
        u.SetExpiresAt(GetNow() + 1000)
        So(err, ShouldBeNil)
        u.UsernamePasswordLogin("")
        j := u.GenerateJWT()
        _, err = FromJWT(j)
        So(err, ShouldNotBeNil)
    })
    Convey("github auth should contains at last an username", t, func() {
        Config.HashKey = GenerateRandomString(32)
        u, err := NewUser("acl-token")
        u.SetExpiresAt(GetNow() + 1000)
        So(err, ShouldBeNil)
        u.GitHubLogin("")
        j := u.GenerateJWT()
        _, err = FromJWT(j)
        So(err, ShouldNotBeNil)
    })
    
    Convey("token  auth should not contains an username", t, func() {
        Config.HashKey = GenerateRandomString(32)
        u, err := NewUser("acl-token")
        u.SetExpiresAt(GetNow() + 1000)
        So(err, ShouldBeNil)
        u.TokenLogin()
        u.Menshend.Username  = "thisSHouldNOtHappen"
        j := u.GenerateJWT()
        _, err = FromJWT(j)
        So(err, ShouldNotBeNil)
    })
    
}

func authProviderPtr(v AuthProviderType) *AuthProviderType {
    return &v
}

//TestGetUser
func TestStopImpersonate(t *testing.T) {
    Convey("Should delete old impersonification", t, func() {
        Convey("if exist",
            func() {
                Config.HashKey = GenerateRandomString(32)
                u, err := NewUser("test-acl")
                So(err, ShouldBeNil)
                
                u.UsernamePasswordLogin("test-user")
                u.SetExpiresAt(GetNow() + 1000)
                ImpErr := u.Impersonate(authProviderPtr(GitHubProvider), StringPtr("criloz"), "admin", "devloper")
                So(ImpErr, ShouldBeNil)
                So(u.Menshend.ImpersonateBy, ShouldNotBeNil)
                u.StopImpersonate()
                So(u.Menshend.ImpersonateBy, ShouldBeNil)
            })
        Convey("Should not remain the same if ImpersonateBy was null",
            func() {
                Config.HashKey = GenerateRandomString(32)
                u, err := NewUser("test-acl")
                So(err, ShouldBeNil)
                u.UsernamePasswordLogin("test-user")
                u.SetExpiresAt(GetNow() + 1000)
                So(u.Menshend.ImpersonateBy, ShouldBeNil)
                u.StopImpersonate()
                So(u.Menshend.ImpersonateBy, ShouldBeNil)
            })
        
    })
    
}


//TestGetUser
func TestImpersonate(t *testing.T) {
    Convey("mpersonate", t, func() {
        Convey("Should faild if new user is empty",
            func() {
                u, err := NewUser("test-acl")
                So(err, ShouldBeNil)
                u.UsernamePasswordLogin("test-user")
                u.SetExpiresAt(GetNow() + 1000)
                ImpErr := u.Impersonate(authProviderPtr(GitHubProvider), nil, "admin", "devloper")
                So(ImpErr, ShouldNotBeNil)
                ImpErr = u.Impersonate(authProviderPtr(GitHubProvider), StringPtr(""), "admin", "devloper")
                So(ImpErr, ShouldNotBeNil)
            })
    
        Convey("Should fail if AuthProvider does not exist",
            func() {
                u, err := NewUser("test-acl")
                So(err, ShouldBeNil)
                u.UsernamePasswordLogin("test-user")
                u.SetExpiresAt(GetNow() + 1000)
                ImpErr := u.Impersonate(authProviderPtr("dsadsdsadasd"), StringPtr("criloz"), "admin", "devloper")
                So(ImpErr, ShouldNotBeNil)
            })
        Convey("Should fail if AuthProvider is Token or user/pass and there are groups",
            func() {
                u, err := NewUser("test-acl")
                So(err, ShouldBeNil)
                u.UsernamePasswordLogin("test-user")
                u.SetExpiresAt(GetNow() + 1000)
                ImpErr := u.Impersonate(authProviderPtr(TokenProvider), StringPtr("criloz"), "admin", "devloper")
                So(ImpErr, ShouldNotBeNil)
            })
    
        Convey("Should fail if user has not username",
            func() {
                u, err := NewUser("test-acl")
                So(err, ShouldBeNil)
                u.TokenLogin()
                u.SetExpiresAt(GetNow() + 1000)
                ImpErr := u.Impersonate(authProviderPtr(TokenProvider), StringPtr("criloz"))
                So(ImpErr, ShouldNotBeNil)
            })
    
    })
    
}
