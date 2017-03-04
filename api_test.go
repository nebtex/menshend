package menshend

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "fmt"
    "net/http/httptest"
    "bytes"
    "net/http"
    "io/ioutil"
    "encoding/json"
    vault "github.com/hashicorp/vault/api"
    "github.com/fatih/structs"
    "github.com/mitchellh/mapstructure"
    "strings"
)

func cleanVault() {
    type ListResult struct {
        Keys []string
    }
    vc, err := vault.NewClient(VaultConfig)
    CheckPanic(err)
    vc.SetToken("myroot")
    
    key := fmt.Sprintf("%s/roles", Config.VaultPath)
    secret, err := vc.Logical().List(key)
    CheckPanic(err)
    if (secret == nil) || (secret.Data == nil) {
        return
    }
    rr := &ListResult{}
    err = mapstructure.Decode(secret.Data, rr)
    CheckPanic(err)
    roleList := rr.Keys
    
    for _, role := range roleList {
        if !strings.HasSuffix(role, "/") {
            continue
        }
        rKey := fmt.Sprintf("%s/roles/%s", Config.VaultPath, role)
        rSecret, err := vc.Logical().List(rKey)
        if err != nil {
            continue
        }
        
        sr := &ListResult{}
        err = mapstructure.Decode(rSecret.Data, sr)
        CheckPanic(err)
        
        serviceList := sr.Keys
        
        for _, service := range serviceList {
            sKey := fmt.Sprintf("%s/roles/%s/%s", Config.VaultPath, role, service)
            _, err := vc.Logical().Delete(sKey)
            CheckPanic(err)
            
        }
    }
    
}

func populateVault() {
    vc, err := vault.NewClient(VaultConfig)
    vc.SetToken("myroot")
    CheckPanic(err)
    roles := map[string]Role{
        "ml-team": map[string]*Service{
            "consul":{},
            "gitlab":{IsActive:false, LuaScript:`
    function getBackend ()
        return "http://gitlab"
    end
    `},
            "postgres":{},
            "redis":{}},
        "admin":map[string]*Service{
            "kubernetes":{IsActive:true,
                LuaScript:`
    function getBackend ()
        return "invalid_url"
    end
    `},
            "vault":{},
            "redis":{IsActive:true, LuaScript:`
    function getBackend ()
        return "http://redis.kv"
    end
    `},
        }}
    for role, services := range roles {
        for service, val := range services {
            key := fmt.Sprintf("%s/roles/%s/%s", Config.VaultPath, role, service)
            _, err := vc.Logical().Write(key, structs.Map(val))
            CheckPanic(err)
        }
    }
}

func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func TestService_ServiceListHandler(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    Convey("Should return all the availabe servies if the user has access " +
        "to them", t, func() {
        
        Convey("return all the services ", func(c C) {
            var u bytes.Buffer
            cleanVault()
            populateVault()
            
            testHandler := func() http.HandlerFunc {
                return http.HandlerFunc(ServiceListHandler)
            }
            ts := httptest.NewServer(NeedLogin(testHandler()))
            defer ts.Close()
            u.WriteString(string(ts.URL))
            u.WriteString("/api/v1/serviceList")
            
            req, err := http.NewRequest("GET", u.String(), nil)
            So(err, ShouldBeNil)
            user, err := NewUser("myroot")
            user.SetExpiresAt(GetNow() + 1000)
            So(err, ShouldBeNil)
            user.GitHubLogin("criloz", "admin", "delos", "umbrella")
            req.AddCookie(&http.Cookie{Name:"menshend-jwt", Value:user.GenerateJWT()})
            client := &http.Client{}
            response, err := client.Do(req)
            So(err, ShouldBeNil)
            jsonResponse, err := ioutil.ReadAll(response.Body)
            So(err, ShouldBeNil)
            umR := map[string]struct{ Roles []string }{}
            err = json.Unmarshal(jsonResponse, &umR)
            So(err, ShouldBeNil)
            allServices := []string{"consul", "gitlab", "kubernetes", "postgres", "redis", "vault"}
            for k := range umR {
                So(contains(allServices, k), ShouldBeTrue)
            }
            
            So(len(umR["redis"].Roles), ShouldEqual, 2)
            
        })
        Convey("if there are not services defined return an empty  list ",
            func(c C) {
                var u bytes.Buffer
                cleanVault()
                testHandler := func() http.HandlerFunc {
                    return http.HandlerFunc(ServiceListHandler)
                }
                ts := httptest.NewServer(NeedLogin(testHandler()))
                defer ts.Close()
                u.WriteString(string(ts.URL))
                u.WriteString("/api/v1/serviceList")
                
                req, err := http.NewRequest("GET", u.String(), nil)
                So(err, ShouldBeNil)
                user, err := NewUser("myroot")
                user.SetExpiresAt(GetNow() + 3600 * 1000)
                So(err, ShouldBeNil)
                user.GitHubLogin("criloz", "admin", "delos", "umbrella")
                req.AddCookie(&http.Cookie{Name:"menshend-jwt", Value:user.GenerateJWT()})
                client := &http.Client{}
                response, err := client.Do(req)
                So(err, ShouldBeNil)
                jsonResponse, err := ioutil.ReadAll(response.Body)
                So(err, ShouldBeNil)
                umR := map[string]struct{ Roles []string }{}
                err = json.Unmarshal(jsonResponse, &umR)
                So(err, ShouldBeNil)
                So(len(umR), ShouldEqual, 0)
            })
        
        Convey("if the user has not access to any service" +
            " return an empty  list ",
            func() {
                var u bytes.Buffer
                cleanVault()
                populateVault()
                
                testHandler := func() http.HandlerFunc {
                    return http.HandlerFunc(ServiceListHandler)
                }
                ts := httptest.NewServer(NeedLogin(testHandler()))
                defer ts.Close()
                u.WriteString(string(ts.URL))
                u.WriteString("/api/v1/serviceList")
                
                req, err := http.NewRequest("GET", u.String(), nil)
                So(err, ShouldBeNil)
                user, err := NewUser("test-acl")
                user.SetExpiresAt(GetNow() + 3600 * 1000)
                So(err, ShouldBeNil)
                user.GitHubLogin("criloz", "admin", "delos", "umbrella")
                req.AddCookie(&http.Cookie{Name:"menshend-jwt", Value:user.GenerateJWT()})
                client := &http.Client{}
                response, err := client.Do(req)
                So(err, ShouldBeNil)
                jsonResponse, err := ioutil.ReadAll(response.Body)
                So(err, ShouldBeNil)
                umR := map[string]struct{ Roles []string }{}
                err = json.Unmarshal(jsonResponse, &umR)
                So(err, ShouldBeNil)
                So(len(umR), ShouldEqual, 0)
            })
        
        Convey("return only the services available to the user", func() {
            var u bytes.Buffer
            cleanVault()
            populateVault()
            vaultClient, err := vault.NewClient(VaultConfig)
            So(err, ShouldBeNil)
            vaultClient.SetToken("myroot")
            
            err = vaultClient.Sys().PutPolicy("api-test-permissions", `
        path "secret/menshend/Roles/admin/*" { policy = "read" }
        path "secret/menshend/Roles" { capabilities = ["list"] }
            `)
            So(err, ShouldBeNil)
            secret, err := vaultClient.Auth().Token().
                Create(&vault.TokenCreateRequest{
                Policies:[]string{"api-test-permissions"}})
            So(err, ShouldBeNil)
            
            testHandler := func() http.HandlerFunc {
                return http.HandlerFunc(ServiceListHandler)
            }
            ts := httptest.NewServer(NeedLogin(testHandler()))
            defer ts.Close()
            u.WriteString(string(ts.URL))
            u.WriteString("/api/v1/serviceList")
            
            req, err := http.NewRequest("GET", u.String(), nil)
            So(err, ShouldBeNil)
            user, err := NewUser(secret.Auth.ClientToken)
            user.SetExpiresAt(GetNow() + 3600 * 1000)
            So(err, ShouldBeNil)
            user.TokenLogin()
            req.AddCookie(&http.Cookie{Name:"menshend-jwt", Value:user.GenerateJWT()})
            client := &http.Client{}
            response, err := client.Do(req)
            So(err, ShouldBeNil)
            jsonResponse, err := ioutil.ReadAll(response.Body)
            So(err, ShouldBeNil)
            umR := map[string]struct{ Roles []string }{}
            err = json.Unmarshal(jsonResponse, &umR)
            So(err, ShouldBeNil)
            So(len(umR), ShouldEqual, 3)
        })
    })
}

func Test_IsAdmin(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    Convey("This function should indicate if the user is admin or" +
        " not", t, func() {
        Convey("Should return false if th user is not an admin", func() {
            cleanVault()
            user, err := NewUser("test_token")
            user.SetExpiresAt(GetNow() + 3600 * 1000)
            So(err, ShouldBeNil)
            user.TokenLogin()
            So(IsAdmin(user), ShouldBeFalse)
            vaultClient,vaultErr:= vault.NewClient(VaultConfig)
            So(vaultErr, ShouldBeNil)
            vaultClient.SetToken("myroot")
    
            vaultErr = vaultClient.Sys().PutPolicy("check-capabilities", `
        path "/sys/capabilities-self" { policy = "read" }
            `)
            So(vaultErr, ShouldBeNil)
            secret, vaultErr := vaultClient.Auth().Token().
                Create(&vault.TokenCreateRequest{
                Policies:[]string{"check-capabilities"}})
            So(vaultErr, ShouldBeNil)
            user, err = NewUser(secret.Auth.ClientToken)
            So(err, ShouldBeNil)
            user.SetExpiresAt(GetNow() + 3600 * 1000)
            So(err, ShouldBeNil)
            user.TokenLogin()
            So(IsAdmin(user), ShouldBeFalse)
        })
        Convey("Should return true if th user is  an admin", func() {
            cleanVault()
            user, err := NewUser("myroot")
            user.SetExpiresAt(GetNow() + 3600 * 1000)
            So(err, ShouldBeNil)
            user.TokenLogin()
            So(IsAdmin(user), ShouldBeTrue)
        })
    })
}

func Test_CanImpersonateHandler(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    
    Convey("This endpoint should indicate if the user can impersonate or" +
        " not", t, func() {
        Convey("Should return false if th user can't impersonate", func() {
            cleanVault()
            user, err := NewUser("test_token")
            user.SetExpiresAt(GetNow() + 3600 * 1000)
            
            So(err, ShouldBeNil)
            user.TokenLogin()
            So(CanImpersonate(user), ShouldBeFalse)
        })
        
        Convey("Should return true if th user can impersonate", func() {
            cleanVault()
            user, err := NewUser("myroot")
            user.SetExpiresAt(GetNow() + 3600 * 1000)
            
            So(err, ShouldBeNil)
            user.TokenLogin()
            So(CanImpersonate(user), ShouldBeTrue)
        })
    })
    
}

func Test_LoginStatusHandler(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    Convey("Should return the login status", t, func() {
        Convey("when user is logged ", func(c C) {
            var u bytes.Buffer
            cleanVault()
            
            testHandler := func() http.HandlerFunc {
                return http.HandlerFunc(LoginStatusHandler)
            }
            
            ts := httptest.NewServer(testHandler())
            defer ts.Close()
            u.WriteString(string(ts.URL))
            u.WriteString("/api/v1/status")
            
            req, err := http.NewRequest("GET", u.String(), nil)
            So(err, ShouldBeNil)
            user, err := NewUser("myroot")
            user.SetExpiresAt(GetNow() + 3600 * 1000)
            So(err, ShouldBeNil)
            user.GitHubLogin("criloz", "admin", "delos", "umbrella")
            req.AddCookie(&http.Cookie{Name:"menshend-jwt",
                Value:user.GenerateJWT()})
            client := &http.Client{}
            response, err := client.Do(req)
            So(err, ShouldBeNil)
            jsonResponse, err := ioutil.ReadAll(response.Body)
            So(err, ShouldBeNil)
            umR := &LoginStatusResponse{}
            err = json.Unmarshal(jsonResponse, &umR)
            So(err, ShouldBeNil)
            So(umR.IsLogged, ShouldBeTrue)
            So(umR.IsAdmin, ShouldBeTrue)
            So(umR.CanImpersonate, ShouldBeTrue)
            
        })
        
        Convey("when user is not logged ", func(c C) {
            var u bytes.Buffer
            cleanVault()
            
            testHandler := func() http.HandlerFunc {
                return http.HandlerFunc(LoginStatusHandler)
            }
            
            ts := httptest.NewServer(testHandler())
            defer ts.Close()
            u.WriteString(string(ts.URL))
            u.WriteString("/api/v1/status")
            
            req, err := http.NewRequest("GET", u.String(), nil)
            So(err, ShouldBeNil)
            
            client := &http.Client{}
            response, err := client.Do(req)
            So(err, ShouldBeNil)
            jsonResponse, err := ioutil.ReadAll(response.Body)
            So(err, ShouldBeNil)
            umR := &LoginStatusResponse{}
            err = json.Unmarshal(jsonResponse, &umR)
            So(err, ShouldBeNil)
            So(umR.IsLogged, ShouldBeFalse)
            So(umR.IsAdmin, ShouldBeFalse)
            So(umR.CanImpersonate, ShouldBeFalse)
            So(umR.SessionExpiresAt, ShouldBeZeroValue)
            
        })
        
    })
}
