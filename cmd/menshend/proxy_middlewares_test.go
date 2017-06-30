package main

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "net/http/httptest"
    "net/http"
    "github.com/nebtex/menshend/pkg/apis/menshend/v1"
    mutils "github.com/nebtex/menshend/pkg/utils"
    mconfig "github.com/nebtex/menshend/pkg/config"
    "os"
    vault "github.com/hashicorp/vault/api"
    "github.com/ansel1/merry"
    testutils "github.com/nebtex/menshend/pkg/utils/test"
    "fmt"
)

func TestNeedLogin(t *testing.T) {
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))

    Convey("Should pass if token is defined, and delete the menshend header", t, func(c C) {
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        So(err, ShouldBeNil)
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: "myroot"}

        httpReq.AddCookie(ct)
        httpWriter := httptest.NewRecorder()

        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
                c.So(req.Context().Value(mutils.VaultToken).(string), ShouldNotBeNil)
                c.So(req.Context().Value(mutils.TokenInfo).(*vault.Secret), ShouldNotBeNil)
                c.So(req.Header.Get("X-Vault-Token"), ShouldBeEmpty)
                c.So(len(req.Cookies()), ShouldEqual, 0)
            }
            return http.HandlerFunc(fn)
        }
        v1.BrowserDetectorHandler((NeedLogin(testHandler()))).ServeHTTP(httpWriter, httpReq)

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
                c.So(merry.Is(x, mutils.NotAuthorized), ShouldBeTrue)
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
        v1.BrowserDetectorHandler(NeedLogin(testHandler())).ServeHTTP(httpWriter, httpReq)

    })

}

func TestRoleHandler(t *testing.T) {
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))

    Convey("if role is defined on cookie should use it", t, func(c C) {
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: "myroot"}
        cm := &http.Cookie{Path: "/", Name: "md-role", Value: "admin"}

        httpReq.AddCookie(ct)
        httpReq.AddCookie(cm)

        httpWriter := httptest.NewRecorder()

        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
                c.So(req.Context().Value(mutils.Role).(string), ShouldEqual, "admin")
            }
            return http.HandlerFunc(fn)
        }
        v1.BrowserDetectorHandler(NeedLogin(RoleHandler(testHandler()))).ServeHTTP(httpWriter, httpReq)

    })

    Convey("if role is not defined should used the default role", t, func(c C) {
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: "myroot"}

        httpReq.AddCookie(ct)

        httpWriter := httptest.NewRecorder()

        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
                c.So(req.Context().Value(mutils.Role).(string), ShouldEqual, mconfig.Config.DefaultRole)
            }
            return http.HandlerFunc(fn)
        }
        v1.BrowserDetectorHandler(NeedLogin(RoleHandler(testHandler()))).ServeHTTP(httpWriter, httpReq)

    })

    Convey("[not browser ]if role is not defined should used the default role", t, func(c C) {
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")

        httpReq.Header.Set("X-Vault-Token", "myroot")

        httpWriter := httptest.NewRecorder()

        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
                c.So(req.Context().Value(mutils.Role).(string), ShouldEqual, mconfig.Config.DefaultRole)
            }
            return http.HandlerFunc(fn)
        }
        v1.BrowserDetectorHandler(NeedLogin(RoleHandler(testHandler()))).ServeHTTP(httpWriter, httpReq)

    })
    Convey("[not browser ]if role is defined in the query menshend should use it", t, func(c C) {
        httpReq, err := http.NewRequest("PUT", "/v1/adminServices/roles/ml-team/gitlab.?md-role=frontend", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        httpReq.Header.Set("X-Vault-Token", "myroot")
        httpWriter := httptest.NewRecorder()
        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
                c.So(req.Context().Value(mutils.Role).(string), ShouldEqual, "frontend")
            }
            return http.HandlerFunc(fn)
        }
        v1.BrowserDetectorHandler(NeedLogin(RoleHandler(testHandler()))).ServeHTTP(httpWriter, httpReq)

    })

    Convey("if role if defined in the query should make redirection and set the cookie", t, func(c C) {
        httpReq, err := http.NewRequest("PUT", "http://consul.menshend.local?md-role=admin&token=xxxx", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        So(err, ShouldBeNil)
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: "myroot"}

        httpReq.AddCookie(ct)

        httpWriter := httptest.NewRecorder()

        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
            }
            return http.HandlerFunc(fn)
        }
        GetSubDomainHandler(v1.BrowserDetectorHandler(NeedLogin(RoleHandler(testHandler())))).ServeHTTP(httpWriter, httpReq)
        So(httpWriter.Result().Header.Get("Location"), ShouldEqual, "http://consul.menshend.local?token=xxxx")
        So(httpWriter.Result().StatusCode, ShouldEqual, 302)
        So(httpWriter.Result().Cookies()[0].Value, ShouldEqual, "admin")

    })

}

func TestGetServiceHandler(t *testing.T) {
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))
    mconfig.Config.Uris.BaseUrl = "http://menshend.com"
    Convey("Should select and service", t, func(c C) {
        testutils.CleanVault()
        testutils.PopulateVault()

        httpReq, err := http.NewRequest("PUT", "http://consul.menshend.com", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        So(err, ShouldBeNil)
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: "myroot"}
        cm := &http.Cookie{Path: "/", Name: "md-role", Value: "ml-team"}

        httpReq.AddCookie(ct)
        httpReq.AddCookie(cm)

        httpWriter := httptest.NewRecorder()

        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
                c.So(req.Context().Value(mutils.Service).(*v1.AdminServiceResource).Meta.SubDomain, ShouldEqual, "consul.")
            }
            return http.HandlerFunc(fn)
        }
        GetSubDomainHandler(v1.BrowserDetectorHandler(NeedLogin(RoleHandler(GetServiceHandler(testHandler()))))).ServeHTTP(httpWriter, httpReq)

    })

    Convey("Should fail if service is not activated", t, func(c C) {
        testutils.CleanVault()
        testutils.PopulateVault()
        defer func() {
            r := recover()
            if (r == nil) {
                t.Error("did not panicked")
                t.Fail()
            }
            switch x := r.(type) {
            case error:
                c.So(merry.Is(x, mutils.NotAuthorized), ShouldBeTrue)
            default:
                t.Errorf("%v", x)
                t.Fail()
            }
        }()
        httpReq, err := http.NewRequest("PUT", "http://gitlab.menshend.com", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        So(err, ShouldBeNil)
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: "myroot"}
        cm := &http.Cookie{Path: "/", Name: "md-role", Value: "ml-team"}

        httpReq.AddCookie(ct)
        httpReq.AddCookie(cm)

        httpWriter := httptest.NewRecorder()

        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
            }
            return http.HandlerFunc(fn)
        }
        GetSubDomainHandler(v1.BrowserDetectorHandler(NeedLogin(RoleHandler(GetServiceHandler(testHandler()))))).ServeHTTP(httpWriter, httpReq)

    })

}

/*
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
*/

func TestPanicHandler(t *testing.T) {
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))
    mconfig.Config.Uris.BaseUrl = "http://menshend.com"
    Convey("if is a request from the browser, and the ui is enabled store the error message on the flashes and redirect to the login page", t, func(c C) {
        testutils.CleanVault()
        testutils.PopulateVault()
        httpReq, err := http.NewRequest("PUT", "http://consul.menshend.com?md-user=other&md-groups=amazon&md-groups=ikea", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        So(err, ShouldBeNil)
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: "invalid-token"}
        cm := &http.Cookie{Path: "/", Name: "md-role", Value: "ml-team"}
        httpReq.AddCookie(ct)
        httpReq.AddCookie(cm)

        httpWriter := httptest.NewRecorder()

        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
            }
            return http.HandlerFunc(fn)
        }
        GetSubDomainHandler(v1.BrowserDetectorHandler(PanicHandler(NeedLogin(RoleHandler(GetServiceHandler(testHandler())))))).ServeHTTP(httpWriter, httpReq)
        So(httpWriter.Header().Get("location"), ShouldEqual, fmt.Sprintf("%s?subdomain=consul.&r=", mconfig.Config.GetLoginPath()))
    })

    Convey("if is a request from other source should return the status error code and the message in the body", t, func(c C) {
        testutils.CleanVault()
        testutils.PopulateVault()

        httpReq, err := http.NewRequest("PUT", "http://consul.menshend.com?md-user=other&md-groups=amazon&md-groups=ikea", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        httpReq.Header.Set("X-Vault-Token", "invalid-token")



        httpWriter := httptest.NewRecorder()

        testHandler := func() http.HandlerFunc {
            fn := func(rw http.ResponseWriter, req *http.Request) {
            }
            return http.HandlerFunc(fn)
        }
        GetSubDomainHandler(v1.BrowserDetectorHandler(PanicHandler(NeedLogin(RoleHandler(GetServiceHandler(testHandler())))))).ServeHTTP(httpWriter, httpReq)
        So(httpWriter.Result().StatusCode, ShouldEqual, 403)
    })

}


/*TODO:
 1. run without ui and configuration
 2. create build.sh
 3. add entrypoint
 4. add test data to json server
 5. add supervisor
 6. web ui for test crfs
 7. kill tunnel at expiration
 */
