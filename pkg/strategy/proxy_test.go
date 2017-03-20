package strategy

import (
    "testing"
    "net/http"
    "net/http/httptest"
    . "github.com/smartystreets/goconvey/convey"
    "io/ioutil"
    "encoding/json"
    "github.com/nebtex/menshend/pkg/resolvers"
    . "github.com/nebtex/menshend/pkg/utils"
    vault "github.com/hashicorp/vault/api"
    "github.com/ansel1/merry"
)

func TestProxy_Execute(t *testing.T) {
    Convey("Should proxy http/https", t, func() {
        tb := &resolvers.YAMLResolve{}
        tb.Content = `baseUrl: http://localhost:8200
headersMap:
  X-Vault-Token: myroot
  h2: t2`
        r := &Proxy{}
        httpReq, err := http.NewRequest("GET", "http://vault.menshend.local/v1/sys/seal-status", nil)
        So(err, ShouldBeNil)
        httpWriter := httptest.NewRecorder()
        noBrowserHandler(r.Execute(tb, &vault.Secret{})).ServeHTTP(httpWriter, httpReq)
        So(httpReq.Header.Get("X-Vault-Token"), ShouldEqual, "myroot")
        d, err := ioutil.ReadAll(httpWriter.Body)
        So(err, ShouldBeNil)
        result := map[string]interface{}{}
        err = json.Unmarshal(d, &result)
        So(err, ShouldBeNil)
        So(result["sealed"], ShouldBeFalse)
    })
    
    Convey("Should panic when backend is not online", t, func(c C) {
        defer func() {
            r := recover()
            if (r == nil) {
                t.Error("did not panicked")
                t.Fail()
            }
            switch x := r.(type) {
            case error:
                c.So(merry.Is(x, BadGateway), ShouldBeTrue)
            default:
                t.Errorf("%v", x)
                t.Fail()
            }
        }()
        tb := &resolvers.YAMLResolve{}
        tb.Content = `baseUrl: http://example.local:444
headersMap:
  X-Vault-Token: myroot
  h2: t2`
        
        r := &Proxy{}
        httpReq, err := http.NewRequest("GET", "http://vault.menshend.local/v1/sys/seal-status", nil)
        So(err, ShouldBeNil)
        httpWriter := httptest.NewRecorder()
        noBrowserHandler(r.Execute(tb, &vault.Secret{})).ServeHTTP(httpWriter, httpReq)
    })
}

func TestProxyHandlersCSRF(t *testing.T) {
    Convey("test csrf protection on proxy", t, func(c C) {
        Convey("should only work when browser and csrf is active", func(c C) {
            Convey(" csrf is not pressent", func(c C) {
                httpReq, err := http.NewRequest("PUT", "http://consul.menshend.com", nil)
                So(err, ShouldBeNil)
                tb := &resolvers.YAMLResolve{}
                tb.Content = `baseUrl: http://example.local:444
headersMap:
  X-Vault-Token: myroot
  h2: t2`
                r := &Proxy{CSRF:true}
                httpWriter := httptest.NewRecorder()
                browserHandler(r.Execute(tb, &vault.Secret{})).ServeHTTP(httpWriter, httpReq)
                So(httpWriter.Result().StatusCode, ShouldEqual, 403)
                So(httpWriter.Result().Cookies()[0].Domain, ShouldEqual, "consul.test.local")
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
                        c.So(merry.Is(x, BadGateway), ShouldBeTrue)
                    default:
                        t.Errorf("%v", x)
                        t.Fail()
                    }
                }()
                
                httpReq, err := http.NewRequest("GET", "http://consul.menshend.com", nil)
                So(err, ShouldBeNil)
                tb := &resolvers.YAMLResolve{}
                tb.Content = `baseUrl: http://example.local:444
headersMap:
  X-Vault-Token: myroot
  h2: t2`
                r := &Proxy{CSRF:true}
                handler := browserHandler(r.Execute(tb, &vault.Secret{}))
                httpWriter := httptest.NewRecorder()
                handler.ServeHTTP(httpWriter, httpReq)
                gorillaCookie := httpWriter.Result().Cookies()[0]
                httpReq, err = http.NewRequest("POST", "http://consul.menshend.com", nil)
                So(err, ShouldBeNil)
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
                        c.So(merry.Is(x, BadGateway), ShouldBeTrue)
                    default:
                        t.Errorf("%v", x)
                        t.Fail()
                    }
                }()
                
                httpReq, err := http.NewRequest("GET", "http://consul-2.menshend.com", nil)
                So(err, ShouldBeNil)
                tb := &resolvers.YAMLResolve{}
                tb.Content = `baseUrl: http://example.local:444
headersMap:
  X-Vault-Token: myroot
  h2: t2`
                httpWriter := httptest.NewRecorder()
                r := &Proxy{CSRF:false}
                handler := browserHandler(r.Execute(tb, &vault.Secret{}))
                handler.ServeHTTP(httpWriter, httpReq)
                
            })
            
            Convey("csrf should not be set", func(c C) {
          
                tb := &resolvers.YAMLResolve{}
                
                tb.Content = `baseUrl: http://localhost:8200
headersMap:
  X-Vault-Token: myroot
  h2: t2`
                
                httpReq, err := http.NewRequest("GET", "http://consul.menshend.com", nil)
                So(err, ShouldBeNil)
                r := &Proxy{CSRF:false}
                handler := browserHandler(r.Execute(tb, &vault.Secret{}))
                httpWriter := httptest.NewRecorder()
                handler.ServeHTTP(httpWriter, httpReq)
                So(len(httpWriter.Result().Cookies()), ShouldEqual, 0)
             
                
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
                            c.So(merry.Is(x, BadGateway), ShouldBeTrue)
                        default:
                            t.Errorf("%v", x)
                            t.Fail()
                        }
                    }()
                    
                    httpReq, err := http.NewRequest("GET", "http://consul-2.menshend.com", nil)
                    So(err, ShouldBeNil)
                    tb := &resolvers.YAMLResolve{}
                    tb.Content = `baseUrl: http://example.local:444
headersMap:
  X-Vault-Token: myroot
  h2: t2`
                    httpWriter := httptest.NewRecorder()
                    r := &Proxy{CSRF:true}
                    handler := noBrowserHandler(r.Execute(tb, &vault.Secret{}))
                    handler.ServeHTTP(httpWriter, httpReq)
					
				})
				Convey("csrf is present", func(c C) {
                    
                    tb := &resolvers.YAMLResolve{}
                    
                    tb.Content = `baseUrl: http://localhost:8200
headersMap:
  X-Vault-Token: myroot
  h2: t2`
                    
                    httpReq, err := http.NewRequest("GET", "http://consul.menshend.com", nil)
                    So(err, ShouldBeNil)
                    r := &Proxy{CSRF:true}
                    handler := noBrowserHandler(r.Execute(tb, &vault.Secret{}))
                    httpWriter := httptest.NewRecorder()
                    handler.ServeHTTP(httpWriter, httpReq)
                    So(len(httpWriter.Result().Cookies()), ShouldEqual, 0)
				})
			})
        
    })
    
}

