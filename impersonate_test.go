package kuper

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "github.com/ansel1/merry"
    "net/http/httptest"
    "bytes"
    "net/http"
    "net/url"
    "strconv"
    vault "github.com/hashicorp/vault/api"
    "fmt"
    "io/ioutil"
)


//TestImpersonateHandler
func TestCheckImpersonatePermission(t *testing.T) {
    var err error
    VaultConfigLocal := vault.DefaultConfig()
    VaultConfigLocal.Address = "http://localhost:8200"
    CheckPanic(err)
    Convey("TestCheckImpersonatePermission",
        t, func(c C) {
            Convey("Should panic if the user has not access to the" +
                " impersonate path", func() {
                defer func() {
                    r := recover()
                    switch x := r.(type) {
                    case error:
                        c.So(merry.Is(x, PermissionError), ShouldBeTrue)
                    default:
                        t.Errorf("%v", x)
                        t.Fail()
                    }
                }()
                u, err := NewUser("test-acl")
                u.SetExpiresAt(getNow() + 3600 * 1000)
                So(err, ShouldBeNil)
                u.GitHubLogin("criloz", "delos", "umbrella")
                checkImpersonatePermission(u, VaultConfigLocal)
            })
            Convey("If the impersonate path not exist it should not fail",
                func() {
                    u, err := NewUser("myroot")
                    u.SetExpiresAt(getNow() + 3600 * 1000)
                    So(err, ShouldBeNil)
                    u.GitHubLogin("criloz", "delos", "umbrella")
                    checkImpersonatePermission(u, VaultConfigLocal)
                })
            
            Convey("Should panic if there is not a vault backend online",
                func() {
                    defer func() {
                        r := recover()
                        c.So(r, ShouldNotBeNil)
                    }()
                    u, err := NewUser("test-acl")
                    u.SetExpiresAt(getNow() + 3600 * 1000)
                    So(err, ShouldBeNil)
                    u.GitHubLogin("criloz", "delos", "umbrella")
                    u.Token = "test_token"
                    dc := vault.DefaultConfig()
                    dc.Address = "http://example.com.local:3212"
                    checkImpersonatePermission(u, dc)
                })
        })
}

//TestSetToken
func TestSetToken(t *testing.T) {
    
    Convey("TestSetToken",
        t, func(c C) {
            Convey("Cookie should has the same expiration date that the token",
                func() {
                    Config.Scheme = "http"
                    u, err := NewUser("test-acl")
                    So(err, ShouldBeNil)
                    u.GitHubLogin("criloz", "delos", "umbrella")
                    w := &httptest.ResponseRecorder{}
                    expTime := MakeTimestampMillisecond() + 3600 * 1000
                    setToken(u, 3600 * 1000, w)
                    r := w.Result()
                    c := r.Cookies()[0]
                    So(c.Value, ShouldEqual, u.GenerateJWT())
                    So(c.HttpOnly, ShouldEqual, true)
                    So(c.Expires.Unix(), ShouldEqual, expTime/1000)
                    So(c.Secure, ShouldEqual, false)
                })
            Convey("Cookie should be secure if kuber is behind an https proxy",
                func() {
                    Config.Scheme = "https"
                    u, err := NewUser("test-acl")
                    u.SetExpiresAt(getNow() + 3600 * 1000)
                    So(err, ShouldBeNil)
                    u.GitHubLogin("criloz", "delos", "umbrella")
                    w := &httptest.ResponseRecorder{}
                    expTime := MakeTimestampMillisecond() + 3600 * 1000
                    setToken(u, 3600 * 1000, w)
                    r := w.Result()
                    c := r.Cookies()[0]
                    So(c.Value, ShouldEqual, u.GenerateJWT())
                    So(c.HttpOnly, ShouldEqual, true)
                    So(c.Expires.Unix(), ShouldEqual, expTime/1000)
                    So(c.Secure, ShouldEqual, true)
                })
        })
}

//TestImpersonateHandler
func TestImpersonateHandler(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    Convey("TestImpersonateHandler",
        t, func(c C) {
            Convey("Should impersonate requested user and groups",
                func() {
                    var u bytes.Buffer
                    testHandler := func() http.HandlerFunc {
                        return http.HandlerFunc(ImpersonateHandler)
                    }
                    ts := httptest.NewServer(NeedLogin(testHandler()))
                    defer ts.Close()
                    u.WriteString(string(ts.URL))
                    u.WriteString("/impersonate")
                    form := url.Values{}
                    form.Add("username", "xUser")
                    form.Add("group", "admin")
                    form.Add("group", "devs")
                    form.Add("authProvider", GitHubProvider)
                    
                    req, err := http.NewRequest("POST", u.String(),
                        bytes.NewBufferString(form.Encode()))
                    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
                    req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))
                    So(err, ShouldBeNil)
                    usr, err := NewUser("myroot")
                    usr.SetExpiresAt(getNow() + 3600 * 1000)
                    So(err, ShouldBeNil)
                    usr.GitHubLogin("criloz", "delos", "umbrella")
                    req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:usr.GenerateJWT()})
                    client := &http.Client{}
                    
                    res, err := client.Do(req)
                    So(err, ShouldBeNil)
                    if res != nil {
                        defer res.Body.Close()
                    }
                    _, err = ioutil.ReadAll(res.Body)
                    So(err, ShouldBeNil)
                    So(200, ShouldEqual, res.StatusCode)
                    jwt := res.Cookies()[0].Value
                    resUser, err := FromJWT(jwt)
                    So(err, ShouldBeNil)
                    So(resUser.Username, ShouldEqual, "xUser")
                    So(resUser.Groups, ShouldContain, "admin")
                    So(resUser.Groups, ShouldContain, "devs")
                    So(resUser.AuthProvider, ShouldEqual, GitHubProvider)
                    So(resUser.ImpersonatedBy, ShouldEqual, "criloz")
                    So(resUser.ImpersonatedByGroups, ShouldContain, "delos")
                    So(resUser.ImpersonatedByGroups, ShouldContain, "umbrella")
                    So(resUser.ImpersonatedByAuthProvider, ShouldEqual,
                        GitHubProvider)
                })
            
            Convey("if the token has not username should panic",
                func(c C) {
                    var u bytes.Buffer
                    panic := false
                    testHandler := func() http.HandlerFunc {
                        checkp := func(w http.ResponseWriter, r *http.Request) {
                            defer func() {
                                r := recover()
                                switch x := r.(type) {
                                case error:
                                    c.So(merry.Is(x, InvalidRequest), ShouldBeTrue)
                                    panic = true
                                default:
                                    t.Errorf("%v", x)
                                    t.Fail()
                                }
                            }()
                            ImpersonateHandler(w, r)
                        }
                        
                        return http.HandlerFunc(checkp)
                    }
                    ts := httptest.NewServer(NeedLogin(testHandler()))
                    defer ts.Close()
                    u.WriteString(string(ts.URL))
                    u.WriteString("/impersonate")
                    form := url.Values{}
                    form.Add("username", "xUser")
                    form.Add("group", "admin")
                    form.Add("group", "devs")
                    form.Add("authProvider", GitHubProvider)
                    
                    req, err := http.NewRequest("POST", u.String(),
                        bytes.NewBufferString(form.Encode()))
                    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
                    req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))
                    So(err, ShouldBeNil)
                    usr, err := NewUser("test_token")
                    usr.SetExpiresAt(getNow() + 3600 * 1000)
                    So(err, ShouldBeNil)
                    usr.TokenLogin()
                    req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:usr.GenerateJWT()})
                    client := &http.Client{}
                    client.Do(req)
                    So(panic, ShouldBeTrue)
                    
                })
            
            Convey("if the new username is not set in the form, the method " +
                "should panic",
                func(c C) {
                    var u bytes.Buffer
                    panic := false
                    testHandler := func() http.HandlerFunc {
                        checkp := func(w http.ResponseWriter, r *http.Request) {
                            defer func() {
                                r := recover()
                                fmt.Println(r, "$$$$$$$$$")
                                switch x := r.(type) {
                                case error:
                                    c.So(merry.Is(x, InvalidFormError),
                                        ShouldBeTrue)
                                    panic = true
                                default:
                                    t.Errorf("%v", x)
                                    t.Fail()
                                }
                            }()
                            ImpersonateHandler(w, r)
                        }
                        
                        return http.HandlerFunc(checkp)
                    }
                    ts := httptest.NewServer(NeedLogin(testHandler()))
                    defer ts.Close()
                    u.WriteString(string(ts.URL))
                    u.WriteString("/impersonate")
                    form := url.Values{}
                    form.Add("group", "admin")
                    form.Add("group", "devs")
                    form.Add("authProvider", GitHubProvider)
                    
                    req, err := http.NewRequest("POST", u.String(),
                        bytes.NewBufferString(form.Encode()))
                    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
                    req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))
                    So(err, ShouldBeNil)
                    usr, err := NewUser("myroot")
                    usr.SetExpiresAt(getNow() + 1000)
                    So(err, ShouldBeNil)
                    usr.GitHubLogin("criloz", "admin")
                    req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:usr.GenerateJWT()})
                    client := &http.Client{}
                    client.Do(req)
                    So(panic, ShouldBeTrue)
                    
                })
            
            Convey("if the new authProvider is not set in the form, the method " +
                "should panic",
                func(c C) {
                    var u bytes.Buffer
                    panic := false
                    testHandler := func() http.HandlerFunc {
                        checkp := func(w http.ResponseWriter, r *http.Request) {
                            defer func() {
                                r := recover()
                                switch x := r.(type) {
                                case error:
                                    c.So(merry.Is(x, InvalidFormError),
                                        ShouldBeTrue)
                                    panic = true
                                default:
                                    t.Errorf("%v", x)
                                    t.Fail()
                                }
                            }()
                            ImpersonateHandler(w, r)
                        }
                        
                        return http.HandlerFunc(checkp)
                    }
                    ts := httptest.NewServer(NeedLogin(testHandler()))
                    defer ts.Close()
                    u.WriteString(string(ts.URL))
                    u.WriteString("/impersonate")
                    form := url.Values{}
                    form.Add("username", "xUser")
                    form.Add("group", "admin")
                    form.Add("group", "devs")
                    
                    req, err := http.NewRequest("POST", u.String(),
                        bytes.NewBufferString(form.Encode()))
                    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
                    req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))
                    So(err, ShouldBeNil)
                    usr, err := NewUser("myroot")
                    usr.SetExpiresAt(getNow() + 1000)
                    So(err, ShouldBeNil)
                    usr.GitHubLogin("criloz", "admin")
                    req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:usr.GenerateJWT()})
                    client := &http.Client{}
                    client.Do(req)
                    So(panic, ShouldBeTrue)
                    
                })
            
            Convey("if the new authProvider is the TokenProvider " +
                "should panic",
                func(c C) {
                    var u bytes.Buffer
                    panicked := false
                    testHandler := func() http.HandlerFunc {
                        checkp := func(w http.ResponseWriter, r *http.Request) {
                            defer func() {
                                r := recover()
                                switch x := r.(type) {
                                case error:
                                    c.So(merry.Is(x, InvalidRequest),
                                        ShouldBeTrue)
                                    panicked = true
                                default:
                                    t.Errorf("%v", x)
                                    t.Fail()
                                }
                            }()
                            ImpersonateHandler(w, r)
                        }
                        
                        return http.HandlerFunc(checkp)
                    }
                    ts := httptest.NewServer(NeedLogin(testHandler()))
                    defer ts.Close()
                    u.WriteString(string(ts.URL))
                    u.WriteString("/impersonate")
                    form := url.Values{}
                    form.Add("username", "xUser")
                    form.Add("group", "admin")
                    form.Add("group", "devs")
                    form.Add("authProvider", TokenProvider)
                    
                    req, err := http.NewRequest("POST", u.String(),
                        bytes.NewBufferString(form.Encode()))
                    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
                    req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))
                    So(err, ShouldBeNil)
                    usr, err := NewUser("myroot")
                    usr.SetExpiresAt(getNow() + 1000)
                    So(err, ShouldBeNil)
                    usr.GitHubLogin("criloz", "admin")
                    req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:usr.GenerateJWT()})
                    client := &http.Client{}
                    client.Do(req)
                    So(panicked, ShouldBeTrue)
                    
                })
            Convey("Should panic when try to use groups with the  " +
                "Username/PasswordProvider",
                func(c C) {
                    var u bytes.Buffer
                    panicked := false
                    testHandler := func() http.HandlerFunc {
                        checkp := func(w http.ResponseWriter, r *http.Request) {
                            defer func() {
                                r := recover()
                                switch x := r.(type) {
                                case error:
                                    c.So(merry.Is(x, InvalidRequest),
                                        ShouldBeTrue)
                                    panicked = true
                                default:
                                    t.Errorf("%v", x)
                                    t.Fail()
                                }
                            }()
                            ImpersonateHandler(w, r)
                        }
                        
                        return http.HandlerFunc(checkp)
                    }
                    ts := httptest.NewServer(NeedLogin(testHandler()))
                    defer ts.Close()
                    u.WriteString(string(ts.URL))
                    u.WriteString("/impersonate")
                    form := url.Values{}
                    form.Add("username", "xUser")
                    form.Add("group", "admin")
                    form.Add("group", "devs")
                    form.Add("authProvider", UsernamePasswordProvider)
                    
                    req, err := http.NewRequest("POST", u.String(),
                        bytes.NewBufferString(form.Encode()))
                    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
                    req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))
                    So(err, ShouldBeNil)
                    usr, err := NewUser("myroot")
                    usr.SetExpiresAt(getNow() + 1000)
                    So(err, ShouldBeNil)
                    usr.GitHubLogin("criloz", "admin")
                    req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:usr.GenerateJWT()})
                    client := &http.Client{}
                    client.Do(req)
                    So(panicked, ShouldBeTrue)
                    
                })
        })
    
}

//TestGetUserFromContext
func TestGetUserFromContext(t *testing.T) {
    Convey("TestGetUserFromContext",
        t, func(c C) {
            Convey("Should panic if user is not in the context",
                func() {
                    var u bytes.Buffer
                    panicked := false
                    
                    testHandler := func() http.HandlerFunc {
                        checkp := func(w http.ResponseWriter, r *http.Request) {
                            defer func() {
                                r := recover()
                                switch x := r.(type) {
                                case error:
                                    c.So(merry.Is(x, UserNotFound),
                                        ShouldBeTrue)
                                    panicked = true
                                default:
                                    t.Errorf("%v", x)
                                    t.Fail()
                                }
                            }()
                            GetUserFromContext(r.Context())
                        }
                        
                        return http.HandlerFunc(checkp)
                    }
                    ts := httptest.NewServer(testHandler())
                    defer ts.Close()
                    u.WriteString(string(ts.URL))
                    u.WriteString("/impersonate")
                    req, err := http.NewRequest("Get", u.String(), nil)
                    So(err, ShouldBeNil)
                    client := &http.Client{}
                    client.Do(req)
                    So(panicked, ShouldBeTrue)
                })
            
        })
}

