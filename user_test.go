package kuper

import (
    "testing"
    "net/http/httptest"
    "bytes"
    "io/ioutil"
    "net/http"
    . "github.com/smartystreets/goconvey/convey"
    "github.com/dgrijalva/jwt-go"
    "github.com/ansel1/merry"
)

//TestGetUser
func TestGetUser(t *testing.T) {
    var err error
    Convey("Should obtain an User struct from a valid jwt token", t, func() {
        MySecretKey = GenerateRandomBytes(64)
        u, err := NewUser("test-acl")
        u.SetExpiresAt(getNow() + 1000)
        So(err, ShouldBeNil)
        u.GitHubLogin("criloz", "delos", "umbrella")
        j := u.GenerateJWT()
        requestUser, err := FromJWT(j)
        So(err, ShouldBeNil)
        So(requestUser.Valid(), ShouldBeNil)
        So(requestUser.Username, ShouldEqual, "criloz")
        
        Convey("If the token use another algo should return error", func() {
            u, err := NewUser("test-acl")
            u.SetExpiresAt(getNow() + 1000)
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
                MySecretKey = GenerateRandomBytes(64)
                u, err := NewUser("test-acl")
                u.SetExpiresAt(getNow() + 1000)
                So(err, ShouldBeNil)
                u.GitHubLogin("criloz", "delos", "umbrella")
                j := u.GenerateJWT()
                MySecretKey = GenerateRandomBytes(64)
                _, err = FromJWT(j)
                So(err, ShouldNotBeNil)
            })
        
    })
    
    Convey("Should mark the token as invalid after the expiration date", t,
        func() {
            MySecretKey = GenerateRandomBytes(64)
            So(err, ShouldBeNil)
            u, err := NewUser("test-acl")
            u.SetExpiresAt(getNow() - 1000)
            So(err, ShouldBeNil)
            u.TokenLogin()
            j := u.GenerateJWT()
            _, err = FromJWT(j)
            So(err, ShouldNotBeNil)
        })
    
    Convey("Should mark the token as invalid if it has not a csrf token", t,
        func() {
            MySecretKey = GenerateRandomBytes(64)
            u, err := NewUser("test-acl")
            u.SetExpiresAt(getNow() + 1000)
            So(err, ShouldBeNil)
            u.CSRFToken = ""
            u.TokenLogin()
            j := u.GenerateJWT()
            _, err = FromJWT(j)
            So(err, ShouldNotBeNil)
            Convey("csrf token should be a valid base64", func() {
                u.CSRFToken = "awdrgyjilp"
                j = u.GenerateJWT()
                _, err = FromJWT(j)
                So(err, ShouldNotBeNil)
                
            })
            
        })
    
    Convey("Should mark the token as invalid if username is not defined " +
        "and someone is using the impersonate feature", t,
        func() {
            MySecretKey = GenerateRandomBytes(64)
            u, err := NewUser("test-acl")
            u.SetExpiresAt(getNow() + 1000)
            So(err, ShouldBeNil)
            u.TokenLogin()
            u.ImpersonatedBy = "criloz"
            j := u.GenerateJWT()
            _, err = FromJWT(j)
            So(err, ShouldNotBeNil)
        })
    Convey("Should mark the token as invalid if it has not a acl token", t,
        func() {
            MySecretKey = GenerateRandomBytes(64)
            So(err, ShouldBeNil)
            u, err := NewUser("")
            u.SetExpiresAt(getNow() + 1000)
            So(err, ShouldBeNil)
            u.UsernamePasswordLogin("criloz")
            j := u.GenerateJWT()
            _, err = FromJWT(j)
            So(err, ShouldNotBeNil)
            
        })
    Convey("Should not contains user or groups when TokenLogin is used", t,
        func() {
            MySecretKey = GenerateRandomBytes(64)
            u, err := NewUser("acl-token")
            u.SetExpiresAt(getNow() + 1000)
            So(err, ShouldBeNil)
            u.TokenLogin()
            u.Username = "criloz"
            u.Groups = []string{"admin", "devs"}
            j := u.GenerateJWT()
            _, err = FromJWT(j)
            So(err, ShouldNotBeNil)
        })
    
    Convey("Should not contains groups when UsernamePasswordLogin is used", t,
        func() {
            MySecretKey = GenerateRandomBytes(64)
            So(err, ShouldBeNil)
            u, err := NewUser("acl-token")
            u.SetExpiresAt(getNow() + 1000)
            So(err, ShouldBeNil)
            u.UsernamePasswordLogin("criloz")
            u.Groups = []string{"admin", "devs"}
            j := u.GenerateJWT()
            _, err = FromJWT(j)
            So(err, ShouldNotBeNil)
        })
    
    Convey("Should contains the AuthProvider", t, func() {
        MySecretKey = GenerateRandomBytes(64)
        So(err, ShouldBeNil)
        u, err := NewUser("acl-token")
        u.SetExpiresAt(getNow() + 1000)
        So(err, ShouldBeNil)
        u.UsernamePasswordLogin("criloz")
        u.AuthProvider = ""
        j := u.GenerateJWT()
        _, err = FromJWT(j)
        So(err, ShouldNotBeNil)
    })
    
}

func TestNeedLogin(t *testing.T) {
    Convey("Should panic when there is not jwt token", t,
        func(c C) {
            var u bytes.Buffer
            panicked := false
            testHandler := func() http.HandlerFunc {
                fn := func(rw http.ResponseWriter, req *http.Request) {
                    panic("test entered test handler, this should not happen")
                }
                return http.HandlerFunc(fn)
            }
            panicHandler := func(next http.Handler) http.Handler {
                return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                    defer func() {
                        r := recover()
                        switch x := r.(type) {
                        case error:
                            panicked = true
                            c.So(merry.Is(x, JWTNotFound), ShouldBeTrue)
                        default:
                            t.Errorf("%v", x)
                            t.Fail()
                        }
                    }()
                    next.ServeHTTP(w, r)
                })
            }
            ts := httptest.NewServer(panicHandler(NeedLogin(testHandler())))
            defer ts.Close()
            u.WriteString(string(ts.URL))
            u.WriteString("/hi")
            var DefaultTransport http.RoundTripper = &http.Transport{}
            req, err := http.NewRequest("GET", u.String(), nil)
            So(err, ShouldBeNil)
            DefaultTransport.RoundTrip(req)
            So(panicked, ShouldBeTrue)
            
        })
    
    Convey("Should panic if there is an invalid token", t,
        func(c C) {
            var u bytes.Buffer
            panicked:=false
            testHandler := func() http.HandlerFunc {
                fn := func(rw http.ResponseWriter, req *http.Request) {
                    panic("test entered test handler, this should not happen")
                }
                return http.HandlerFunc(fn)
            }
            panicHandler := func(next http.Handler) http.Handler {
                return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                    defer func() {
                        r := recover()
                        switch x := r.(type) {
                        case error:
                            panicked = true
                            c.So(merry.Is(x, InvalidJWT), ShouldBeTrue)
                        default:
                            t.Errorf("%v", x)
                            t.Fail()
                        }
                    }()
                    next.ServeHTTP(w, r)
                })
            }
            ts := httptest.NewServer(panicHandler(NeedLogin(testHandler())))
            defer ts.Close()
            u.WriteString(string(ts.URL))
            u.WriteString("/hi")
            var DefaultTransport http.RoundTripper = &http.Transport{}
            req, err := http.NewRequest("GET", u.String(), nil)
            So(err, ShouldBeNil)
            usr, err := NewUser("acl")
            usr.SetExpiresAt(getNow() + 1000)
            So(err, ShouldBeNil)
            usr.TokenLogin()
            // if the auth method is TokenLogin, set a username will
            // make the token invalid
            usr.Username = "criloz"
            req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:usr.GenerateJWT()})
            DefaultTransport.RoundTrip(req)
            So(panicked, ShouldBeTrue)
    
        })
    Convey("Should make the user available in the context if the token" +
        " is valid", t, func(c C) {
        var u bytes.Buffer
        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
                user := req.Context().Value("User").(*User)
                c.So(user, ShouldNotBeNil)
                c.So("acl", ShouldEqual, user.Token)
            }
            return http.HandlerFunc(fn)
        }
        ts := httptest.NewServer(NeedLogin(testHandler()))
        defer ts.Close()
        u.WriteString(string(ts.URL))
        u.WriteString("/hi")
        var DefaultTransport http.RoundTripper = &http.Transport{}
        req, err := http.NewRequest("GET", u.String(), nil)
        So(err, ShouldBeNil)
        usr, err := NewUser("acl")
        usr.SetExpiresAt(getNow() + 1000)
        So(err, ShouldBeNil)
        usr.TokenLogin()
        req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:usr.GenerateJWT()})
        res, err := DefaultTransport.RoundTrip(req)
        So(err, ShouldBeNil)
        if res != nil {
            defer res.Body.Close()
        }
        _, err = ioutil.ReadAll(res.Body)
        So(err, ShouldBeNil)
        So(200, ShouldEqual, res.StatusCode)
        
    })
}
