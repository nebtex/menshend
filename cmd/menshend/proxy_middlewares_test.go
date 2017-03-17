package main

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "net/http/httptest"
    "net/http"
    . "github.com/nebtex/menshend/pkg/utils"
    . "github.com/nebtex/menshend/pkg/utils/test"
    . "github.com/nebtex/menshend/pkg/users"
    . "github.com/nebtex/menshend/pkg/apis/menshend"
    "github.com/ansel1/merry"
    "github.com/nebtex/menshend/pkg/config"
    "github.com/nebtex/menshend/pkg/apis/menshend/v1"
    "github.com/nebtex/menshend/pkg/strategy"
)

func TestDetectBrowser(t *testing.T) {
    
    Convey("if the menshend token comes in the header should detect not detect a browser environment", t, func(c C) {
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        httpReq.Header.Add("X-Vault-Token", user.GenerateJWT())
        httpWriter := httptest.NewRecorder()
        
        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
                c.So(req.Context().Value("IsBrowserRequest").(bool), ShouldBeFalse)
            }
            return http.HandlerFunc(fn)
        }
        DetectBrowser(testHandler()).ServeHTTP(httpWriter, httpReq)
    })
    Convey("if the menshend token comes in the cookie should detect a browser environment", t, func(c C) {
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: user.GenerateJWT()}
        oc := &http.Cookie{Path: "/", Name: "asdf", Value: "asdf"}
        
        httpReq.AddCookie(ct)
        httpReq.AddCookie(oc)
        
        httpWriter := httptest.NewRecorder()
        
        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
                c.So(req.Context().Value("IsBrowserRequest").(bool), ShouldBeTrue)
                c.So(req.Header.Get("X-Vault-Token"), ShouldEqual, ct.Value)
                c.So(len(req.Cookies()), ShouldEqual, 1)
            }
            return http.HandlerFunc(fn)
        }
        DetectBrowser(testHandler()).ServeHTTP(httpWriter, httpReq)
    })
}

func TestNeedLogin(t *testing.T) {
    
    Convey("Should pass if user is defined, and delete the menshend header", t, func(c C) {
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: user.GenerateJWT()}
        
        httpReq.AddCookie(ct)
        httpWriter := httptest.NewRecorder()
        
        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
                c.So(req.Context().Value("User").(*User), ShouldNotBeNil)
                c.So(req.Header.Get("X-Vault-Token"), ShouldBeEmpty)
            }
            return http.HandlerFunc(fn)
        }
        DetectBrowser(NeedLogin(testHandler())).ServeHTTP(httpWriter, httpReq)
        
    })
    
    Convey("Should fail if the user is not defined", t, func(c C) {
        defer func() {
            r := recover()
            if (r == nil) {
                t.Error("did not panicked")
                t.Fail()
            }
            switch x := r.(type) {
            case error:
                c.So(merry.Is(x, NotAuthorized), ShouldBeTrue)
            default:
                t.Errorf("%v", x)
                t.Fail()
            }
        }()
        
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        httpWriter := httptest.NewRecorder()
        
        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
            }
            return http.HandlerFunc(fn)
        }
        DetectBrowser(NeedLogin(testHandler())).ServeHTTP(httpWriter, httpReq)
        
    })
    
}

func TestRoleHandler(t *testing.T) {
    
    Convey("if role is defined on cookie should use it", t, func(c C) {
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: user.GenerateJWT()}
        cm := &http.Cookie{Path: "/", Name: "md-role", Value: "admin"}
        
        httpReq.AddCookie(ct)
        httpReq.AddCookie(cm)
        
        httpWriter := httptest.NewRecorder()
        
        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
                c.So(req.Context().Value("role").(string), ShouldEqual, "admin")
            }
            return http.HandlerFunc(fn)
        }
        DetectBrowser(NeedLogin(RoleHandler(testHandler()))).ServeHTTP(httpWriter, httpReq)
        
    })
    
    Convey("if role is not defined should used the default role", t, func(c C) {
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: user.GenerateJWT()}
        
        httpReq.AddCookie(ct)
        
        httpWriter := httptest.NewRecorder()
        
        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
                c.So(req.Context().Value("role").(string), ShouldEqual, config.Config.DefaultRole)
            }
            return http.HandlerFunc(fn)
        }
        DetectBrowser(NeedLogin(RoleHandler(testHandler()))).ServeHTTP(httpWriter, httpReq)
        
    })
    
    Convey("[not browser ]if role is not defined should used the default role", t, func(c C) {
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        
        httpReq.Header.Set("X-Vault-Token", user.GenerateJWT())
        
        httpWriter := httptest.NewRecorder()
        
        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
                c.So(req.Context().Value("role").(string), ShouldEqual, config.Config.DefaultRole)
            }
            return http.HandlerFunc(fn)
        }
        DetectBrowser(NeedLogin(RoleHandler(testHandler()))).ServeHTTP(httpWriter, httpReq)
        
    })
    Convey("[not browser ]if role is defined in the query menshend should use it", t, func(c C) {
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.?md-role=frontend", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        
        httpReq.Header.Set("X-Vault-Token", user.GenerateJWT())
        
        httpWriter := httptest.NewRecorder()
        
        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
                c.So(req.Context().Value("role").(string), ShouldEqual, "frontend")
            }
            return http.HandlerFunc(fn)
        }
        DetectBrowser(NeedLogin(RoleHandler(testHandler()))).ServeHTTP(httpWriter, httpReq)
        
    })
    
    Convey("if role if defined in the query should make redirection and set the cookie", t, func(c C) {
        httpReq, err := http.NewRequest("PUT", "http://consul.menshend.local?md-role=admin&token=xxxx", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: user.GenerateJWT()}
        
        httpReq.AddCookie(ct)
        
        httpWriter := httptest.NewRecorder()
        
        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
            }
            return http.HandlerFunc(fn)
        }
        DetectBrowser(NeedLogin(RoleHandler(testHandler()))).ServeHTTP(httpWriter, httpReq)
        So(httpWriter.Result().Header.Get("Location"), ShouldEqual, "http://consul.menshend.local?token=xxxx")
        So(httpWriter.Result().StatusCode, ShouldEqual, 302)
        So(httpWriter.Result().Cookies()[0].Value, ShouldEqual, "admin")
        
    })
    
}

func TestGetServiceHandler(t *testing.T) {
    config.VaultConfig.Address = "http://127.0.0.1:8200"
    config.Config.Uris.BaseUrl = "http://menshend.com"
    Convey("Should select and service", t, func(c C) {
        CleanVault()
        PopulateVault()
        
        httpReq, err := http.NewRequest("PUT", "http://consul.menshend.com", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: user.GenerateJWT()}
        cm := &http.Cookie{Path: "/", Name: "md-role", Value: "ml-team"}
        
        httpReq.AddCookie(ct)
        httpReq.AddCookie(cm)
        
        httpWriter := httptest.NewRecorder()
        
        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
                c.So(req.Context().Value("service").(*v1.AdminServiceResource).SubDomain, ShouldEqual, "consul.")
            }
            return http.HandlerFunc(fn)
        }
        GetSubDomainHandler(DetectBrowser(NeedLogin(RoleHandler(GetServiceHandler(testHandler()))))).ServeHTTP(httpWriter, httpReq)
        
    })
    
    Convey("Should fail if service is not activated", t, func(c C) {
        CleanVault()
        PopulateVault()
        defer func() {
            r := recover()
            if (r == nil) {
                t.Error("did not panicked")
                t.Fail()
            }
            switch x := r.(type) {
            case error:
                c.So(merry.Is(x, NotAuthorized), ShouldBeTrue)
            default:
                t.Errorf("%v", x)
                t.Fail()
            }
        }()
        httpReq, err := http.NewRequest("PUT", "http://gitlab.menshend.com", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: user.GenerateJWT()}
        cm := &http.Cookie{Path: "/", Name: "md-role", Value: "ml-team"}
        
        httpReq.AddCookie(ct)
        httpReq.AddCookie(cm)
        
        httpWriter := httptest.NewRecorder()
        
        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
            }
            return http.HandlerFunc(fn)
        }
        GetSubDomainHandler(DetectBrowser(NeedLogin(RoleHandler(GetServiceHandler(testHandler()))))).ServeHTTP(httpWriter, httpReq)
        
    })
    
}

func TestImpersonateWithinRoleHandler(t *testing.T) {
    config.VaultConfig.Address = "http://127.0.0.1:8200"
    config.Config.Uris.BaseUrl = "http://menshend.com"
    Convey("If impersonate withing role is active any user can impersonate any other user in the service", t, func(c C) {
        CleanVault()
        PopulateVault()
        
        httpReq, err := http.NewRequest("PUT", "http://consul.menshend.com?md-user=other&md-groups=amazon&md-groups=ikea", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("myroot")
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: user.GenerateJWT()}
        cm := &http.Cookie{Path: "/", Name: "md-role", Value: "ml-team"}
        
        httpReq.AddCookie(ct)
        httpReq.AddCookie(cm)
        
        httpWriter := httptest.NewRecorder()
        
        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
                c.So(req.Context().Value("User").(*User).Menshend.Username, ShouldEqual, "other")
                c.So(req.Context().Value("User").(*User).Menshend.Groups, ShouldContain, "amazon")
                c.So(req.Context().Value("User").(*User).Menshend.Groups, ShouldContain, "ikea")
            }
            return http.HandlerFunc(fn)
        }
        GetSubDomainHandler(DetectBrowser(NeedLogin(RoleHandler(GetServiceHandler(ImpersonateWithinRoleHandler(testHandler())))))).ServeHTTP(httpWriter, httpReq)
        
    })
    
}

func TestProxyHandlersCSRF(t *testing.T) {
    config.VaultConfig.Address = "http://127.0.0.1:8200"
    config.Config.Uris.BaseUrl = "http://menshend.com"
    Convey("test csrf protection on proxy", t, func(c C) {
        CleanVault()
        PopulateVault()
        
        Convey("should only work when browser and csrf is active", func(c C) {
            Convey(" csrf is not pressent", func(c C) {
                
                httpReq, err := http.NewRequest("PUT", "http://consul.menshend.com", nil)
                So(err, ShouldBeNil)
                httpReq.Header.Set("Content-Type", "application/json")
                user, err := NewUser("myroot")
                So(err, ShouldBeNil)
                user.TokenLogin()
                user.SetExpiresAt(GetNow() + 3600)
                ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: user.GenerateJWT()}
                cm := &http.Cookie{Path: "/", Name: "md-role", Value: "ml-team"}
                httpReq.AddCookie(ct)
                httpReq.AddCookie(cm)
                httpWriter := httptest.NewRecorder()
                
                testHandler := func() http.HandlerFunc {
                    fn := func(rw http.ResponseWriter, req *http.Request) {
                    }
                    return http.HandlerFunc(fn)
                }
                GetSubDomainHandler(DetectBrowser(NeedLogin(RoleHandler(GetServiceHandler(ImpersonateWithinRoleHandler(ProxyHandlers(testHandler()))))))).ServeHTTP(httpWriter, httpReq)
                So(httpWriter.Result().StatusCode, ShouldEqual, 403)
                So(httpWriter.Result().Cookies()[0].Domain, ShouldEqual, "consul.menshend.com")
            })
            Convey("csrf is present", func(c C) {
                defer func() {
                    r := recover()
                    if (r == nil) {
                        t.Error("did not panicked")
                        t.Fail()
                    }
                    switch x := r.(type) {
                    case error:
                        c.So(merry.Is(x, strategy.BadGateway), ShouldBeTrue)
                    default:
                        t.Errorf("%v", x)
                        t.Fail()
                    }
                }()
                
                httpReq, err := http.NewRequest("GET", "http://consul.menshend.com", nil)
                So(err, ShouldBeNil)
                user, err := NewUser("myroot")
                So(err, ShouldBeNil)
                user.TokenLogin()
                user.SetExpiresAt(GetNow() + 3600)
                ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: user.GenerateJWT()}
                cm := &http.Cookie{Path: "/", Name: "md-role", Value: "ml-team"}
                httpReq.AddCookie(ct)
                httpReq.AddCookie(cm)
                httpWriter := httptest.NewRecorder()
                
                testHandler := func() http.HandlerFunc {
                    fn := func(rw http.ResponseWriter, req *http.Request) {
                    }
                    return http.HandlerFunc(fn)
                }
                handler := GetSubDomainHandler(DetectBrowser(NeedLogin(RoleHandler(GetServiceHandler(ImpersonateWithinRoleHandler(ProxyHandlers(testHandler())))))))
                handler.ServeHTTP(httpWriter, httpReq)
                gorillaCookie := httpWriter.Result().Cookies()[0]
                httpReq, err = http.NewRequest("POST", "http://consul.menshend.com", nil)
                So(err, ShouldBeNil)
                httpReq.AddCookie(ct)
                httpReq.AddCookie(cm)
                httpReq.AddCookie(gorillaCookie)
                httpReq.Header.Set("X-CSRF-Token", httpWriter.Header().Get("X-Next-CSRF-Token"))
                httpWriter = httptest.NewRecorder()
                handler.ServeHTTP(httpWriter, httpReq)
                
            })
        })
        
        Convey("if csrf protection is not active should pass any case", func(c C) {
            Convey(" csrf was not send", func(c C) {
                defer func() {
                    r := recover()
                    if (r == nil) {
                        t.Error("did not panicked")
                        t.Fail()
                    }
                    switch x := r.(type) {
                    case error:
                        c.So(merry.Is(x, strategy.BadGateway), ShouldBeTrue)
                    default:
                        t.Errorf("%v", x)
                        t.Fail()
                    }
                }()
                
                httpReq, err := http.NewRequest("GET", "http://consul-2.menshend.com", nil)
                So(err, ShouldBeNil)
                user, err := NewUser("myroot")
                So(err, ShouldBeNil)
                user.TokenLogin()
                user.Menshend.Realm = BrowserRealm
                user.SetExpiresAt(GetNow() + 3600)
                ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: user.GenerateJWT()}
                cm := &http.Cookie{Path: "/", Name: "md-role", Value: "ml-team"}
                httpReq.AddCookie(ct)
                httpReq.AddCookie(cm)
                httpWriter := httptest.NewRecorder()
                
                testHandler := func() http.HandlerFunc {
                    fn := func(rw http.ResponseWriter, req *http.Request) {
                    }
                    return http.HandlerFunc(fn)
                }
                handler := GetSubDomainHandler(DetectBrowser(NeedLogin(TokenRealmSecurityHandler(RoleHandler(GetServiceHandler(ImpersonateWithinRoleHandler(ProxyHandlers(testHandler()))))))))
                handler.ServeHTTP(httpWriter, httpReq)
                
            })
            Convey("csrf is present", func(c C) {
                defer func() {
                    r := recover()
                    if (r == nil) {
                        t.Error("did not panicked")
                        t.Fail()
                    }
                    switch x := r.(type) {
                    case error:
                        c.So(merry.Is(x, strategy.BadGateway), ShouldBeTrue)
                    default:
                        t.Errorf("%v", x)
                        t.Fail()
                    }
                }()
                
                httpReq, err := http.NewRequest("GET", "http://consul-2.menshend.com", nil)
                So(err, ShouldBeNil)
                user, err := NewUser("myroot")
                So(err, ShouldBeNil)
                user.TokenLogin()
                user.SetExpiresAt(GetNow() + 3600)
                ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: user.GenerateJWT()}
                cm := &http.Cookie{Path: "/", Name: "md-role", Value: "ml-team"}
                httpReq.AddCookie(ct)
                httpReq.AddCookie(cm)
                httpWriter := httptest.NewRecorder()
                
                testHandler := func() http.HandlerFunc {
                    fn := func(rw http.ResponseWriter, req *http.Request) {
                    }
                    return http.HandlerFunc(fn)
                }
                handler := GetSubDomainHandler(DetectBrowser(NeedLogin(RoleHandler(GetServiceHandler(ImpersonateWithinRoleHandler(ProxyHandlers(testHandler())))))))
                handler.ServeHTTP(httpWriter, httpReq)
                gorillaCookie := httpWriter.Result().Cookies()[0]
                httpReq, err = http.NewRequest("POST", "http://consul.menshend.com", nil)
                So(err, ShouldBeNil)
                httpReq.AddCookie(ct)
                httpReq.AddCookie(cm)
                httpReq.AddCookie(gorillaCookie)
                httpReq.Header.Set("X-CSRF-Token", httpWriter.Header().Get("X-Next-CSRF-Token"))
                httpWriter = httptest.NewRecorder()
                handler.ServeHTTP(httpWriter, httpReq)
                
            })
        })
        Convey("should pass any case when is not a browser request, ", func(c C) {
            Convey(" csrf was not send", func(c C) {
                defer func() {
                    r := recover()
                    if (r == nil) {
                        t.Error("did not panicked")
                        t.Fail()
                    }
                    switch x := r.(type) {
                    case error:
                        c.So(merry.Is(x, strategy.BadGateway), ShouldBeTrue)
                    default:
                        t.Errorf("%v", x)
                        t.Fail()
                    }
                }()
                httpReq, err := http.NewRequest("GET", "http://consul.menshend.com", nil)
                So(err, ShouldBeNil)
                user, err := NewUser("myroot")
                user.Menshend.Realm = ApiRealm
                So(err, ShouldBeNil)
                user.TokenLogin()
                user.SetExpiresAt(GetNow() + 3600)
                httpReq.Header.Set("md-role", "ml-team")
                httpReq.Header.Set("X-Vault-Token", user.GenerateJWT())
                httpWriter := httptest.NewRecorder()
    
                testHandler := func() http.HandlerFunc {
                    fn := func(rw http.ResponseWriter, req *http.Request) {
                    }
                    return http.HandlerFunc(fn)
                }
                handler := GetSubDomainHandler(DetectBrowser(NeedLogin(TokenRealmSecurityHandler(RoleHandler(GetServiceHandler(ImpersonateWithinRoleHandler(ProxyHandlers(testHandler()))))))))
                handler.ServeHTTP(httpWriter, httpReq)
                
            })
            Convey("csrf is present", func(c C) {
                defer func() {
                    r := recover()
                    if (r == nil) {
                        t.Error("did not panicked")
                        t.Fail()
                    }
                    switch x := r.(type) {
                    case error:
                        c.So(merry.Is(x, strategy.BadGateway), ShouldBeTrue)
                    default:
                        t.Errorf("%v", x)
                        t.Fail()
                    }
                }()
    
                httpReq, err := http.NewRequest("GET", "http://consul.menshend.com", nil)
                So(err, ShouldBeNil)
                user, err := NewUser("myroot")
                So(err, ShouldBeNil)
                user.TokenLogin()
                user.SetExpiresAt(GetNow() + 3600)
                user.Menshend.Realm = ApiRealm
                httpReq.Header.Set("md-role", "ml-team")
                httpReq.Header.Set("X-Vault-Token", user.GenerateJWT())
                httpWriter := httptest.NewRecorder()
    
                testHandler := func() http.HandlerFunc {
                    fn := func(rw http.ResponseWriter, req *http.Request) {
                    }
                    return http.HandlerFunc(fn)
                }
                handler := GetSubDomainHandler(DetectBrowser(NeedLogin(RoleHandler(GetServiceHandler(ImpersonateWithinRoleHandler(ProxyHandlers(testHandler())))))))
                handler.ServeHTTP(httpWriter, httpReq)
                gorillaCookie := httpWriter.Result().Cookies()[0]
                httpReq, err = http.NewRequest("POST", "http://consul.menshend.com", nil)
                So(err, ShouldBeNil)
                httpReq.Header.Set("md-role", "ml-team")
                httpReq.Header.Set("X-Vault-Token", user.GenerateJWT())
                httpReq.AddCookie(gorillaCookie)
                httpReq.Header.Set("X-CSRF-Token", httpWriter.Header().Get("X-Next-CSRF-Token"))
                httpWriter = httptest.NewRecorder()
                handler.ServeHTTP(httpWriter, httpReq)
            })
        })
        
    })
    
    
}


func TestPanicHandler(t *testing.T) {
    config.VaultConfig.Address = "http://127.0.0.1:8200"
    config.Config.Uris.BaseUrl = "http://menshend.com"
    Convey("if is a request from the browser, and the ui is enabled store the error message on the flashes and redirect to the login page", t, func(c C) {
        CleanVault()
        PopulateVault()
        config.Config.EnableUI = true
    
        httpReq, err := http.NewRequest("PUT", "http://consul.menshend.com?md-user=other&md-groups=amazon&md-groups=ikea", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("invalid token")
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: user.GenerateJWT()}
        cm := &http.Cookie{Path: "/", Name: "md-role", Value: "ml-team"}
        
        httpReq.AddCookie(ct)
        httpReq.AddCookie(cm)
        
        httpWriter := httptest.NewRecorder()
        
        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
            }
            return http.HandlerFunc(fn)
        }
        GetSubDomainHandler(DetectBrowser(PanicHandler(NeedLogin(RoleHandler(GetServiceHandler(ImpersonateWithinRoleHandler(testHandler()))))))).ServeHTTP(httpWriter, httpReq)
        So(httpWriter.Header().Get("location"), ShouldEqual, "http://menshend.com/ui/login")
    })
    
    Convey("if is a request from the browser, and the ui is disabled should return the status error code and the message in the body", t, func(c C) {
        CleanVault()
        PopulateVault()
        config.Config.EnableUI = false
        
        httpReq, err := http.NewRequest("PUT", "http://consul.menshend.com?md-user=other&md-groups=amazon&md-groups=ikea", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        user, err := NewUser("invalid token")
        So(err, ShouldBeNil)
        user.TokenLogin()
        user.SetExpiresAt(GetNow() + 3600)
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: user.GenerateJWT()}
        cm := &http.Cookie{Path: "/", Name: "md-role", Value: "ml-team"}
        
        httpReq.AddCookie(ct)
        httpReq.AddCookie(cm)
        
        httpWriter := httptest.NewRecorder()
        
        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
            }
            return http.HandlerFunc(fn)
        }
        GetSubDomainHandler(DetectBrowser(PanicHandler(NeedLogin(RoleHandler(GetServiceHandler(ImpersonateWithinRoleHandler(testHandler()))))))).ServeHTTP(httpWriter, httpReq)
        So(httpWriter.Result().StatusCode, ShouldEqual, 403)
    })
    
}
