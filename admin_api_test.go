package kuper

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "net/http/httptest"
    "bytes"
    "net/http"
    "io/ioutil"
    "encoding/json"
    vault "github.com/hashicorp/vault/api"
    "github.com/mitchellh/mapstructure"
    "github.com/gorilla/mux"
)

func getServicePayload() *ServicePayload {
    c := &ServicePayload{}
    c.Logo = "https://www.consul.io/assets/images/logo_large-475cebb0.png"
    c.Name = "consul"
    c.SubDomain = "consul"
    c.LongDescription = "dummy_description"
    c.LongDescriptionUrl = "https://raw.githubusercontent.com/hashicorp" +
        "/consul/master/README.md"
    c.ShortDescription = "Consul is a tool for service discovery and " +
        "configuration. Consul is distributed, highly available," +
        " and extremely scalable."
    c.Roles = map[string]*ServiceRole{}
    c.Roles["admin"] = &ServiceRole{
        LuaScript: ":)",
        ImpersonateWithinRole: true,
        IsActive: true,
        Proxy:false,
        
    }
    
    c.Roles["devops"] = &ServiceRole{
        LuaScript: ":)",
        ImpersonateWithinRole: true,
        IsActive: true,
        Proxy:false,
    }
    
    return c
}

func getDeleteServicePayload() *DeleteServicePayload {
    c := &DeleteServicePayload{}
    c.SubDomain = "consul"
    c.Roles = []string{"admin", "devops"}
    return c
}

func TestCreateEditServiceHandler(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    Convey("Should create or modify a service", t, func() {
        Convey("Create or modify a service with no roles should fail",
            func(c C) {
                var u bytes.Buffer
                cleanVault()
                
                testHandler := func() http.HandlerFunc {
                    return http.HandlerFunc(CreateEditServiceHandler)
                }
                ts := httptest.NewServer(NeedLogin(NeedAdmin(testHandler())))
                defer ts.Close()
                u.WriteString(string(ts.URL))
                u.WriteString("/admin_api/v1/mkService")
                service := getServicePayload()
                service.Roles = nil
                postBody, err := json.Marshal(service)
                So(err, ShouldBeNil)
                req, err := http.NewRequest("POST", u.String(),
                    bytes.NewReader(postBody))
                So(err, ShouldBeNil)
                user, err := NewUser("myroot")
                user.SetExpiresAt(getNow() + 3600 * 1000)
                So(err, ShouldBeNil)
                user.GitHubLogin("criloz", "admin", "delos", "umbrella")
                req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:user.GenerateJWT()})
                client := &http.Client{}
                response, err := client.Do(req)
                So(err, ShouldBeNil)
                jsonResponse, err := ioutil.ReadAll(response.Body)
                So(err, ShouldBeNil)
                umR := struct {
                    Success bool
                    Message string
                }{}
                err = json.Unmarshal(jsonResponse, &umR)
                So(err, ShouldBeNil)
                So(umR.Success, ShouldBeFalse)
                So(umR.Message, ShouldEqual, "Not role was provided")
            })
        
        Convey("Request with not body or bad json should fail",
            func(c C) {
                var u bytes.Buffer
                cleanVault()
                testHandler := func() http.HandlerFunc {
                    return http.HandlerFunc(CreateEditServiceHandler)
                }
                ts := httptest.NewServer(NeedLogin(NeedAdmin(testHandler())))
                defer ts.Close()
                u.WriteString(string(ts.URL))
                u.WriteString("/admin_api/v1/mkService")
                req, err := http.NewRequest("POST", u.String(), nil)
                So(err, ShouldBeNil)
                user, err := NewUser("myroot")
                user.SetExpiresAt(getNow() + 3600 * 1000)
                So(err, ShouldBeNil)
                user.GitHubLogin("criloz", "admin", "delos", "umbrella")
                req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:user.GenerateJWT()})
                client := &http.Client{}
                response, err := client.Do(req)
                So(err, ShouldBeNil)
                jsonResponse, err := ioutil.ReadAll(response.Body)
                So(err, ShouldBeNil)
                umR := struct {
                    Success bool
                    Message string
                }{}
                err = json.Unmarshal(jsonResponse, &umR)
                So(err, ShouldBeNil)
                So(umR.Success, ShouldBeFalse)
                So(umR.Message, ShouldEqual, "Please send a valid json")
            })
        
        Convey("Should save the service with all the roles",
            func(c C) {
                var u bytes.Buffer
                cleanVault()
                testHandler := func() http.HandlerFunc {
                    return http.HandlerFunc(CreateEditServiceHandler)
                }
                ts := httptest.NewServer(NeedLogin(NeedAdmin(testHandler())))
                defer ts.Close()
                u.WriteString(string(ts.URL))
                u.WriteString("/admin_api/v1/mkService")
                service := getServicePayload()
                postBody, err := json.Marshal(service)
                So(err, ShouldBeNil)
                req, err := http.NewRequest("POST", u.String(),
                    bytes.NewReader(postBody))
                So(err, ShouldBeNil)
                user, err := NewUser("myroot")
                user.SetExpiresAt(getNow() + 3600 * 1000)
                So(err, ShouldBeNil)
                user.GitHubLogin("criloz", "admin", "delos", "umbrella")
                req.AddCookie(&http.Cookie{Name:"kuper-jwt",
                    Value:user.GenerateJWT()})
                client := &http.Client{}
                response, err := client.Do(req)
                So(err, ShouldBeNil)
                jsonResponse, err := ioutil.ReadAll(response.Body)
                
                So(err, ShouldBeNil)
                umR := AdminServiceResponse{}
                err = json.Unmarshal(jsonResponse, &umR)
                
                So(err, ShouldBeNil)
                for _, v := range umR.Roles {
                    So(v.Success, ShouldBeTrue)
                }
            })
        
    })
}

func TestCreateEditServiceHandler_Permissions(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    
    Convey("Should return permission error if the user can't save a " +
        "service in the role", t, func(c C) {
        var u bytes.Buffer
        cleanVault()
        vClient, err := vault.NewClient(VaultConfig)
        So(err, ShouldBeNil)
        vClient.SetToken("myroot")
        err = vClient.Sys().PutPolicy("admin-test-cesh", `
        path "secret/kuper/Roles/devops/*" { policy = "write" }
        path "secret/kuper/Admin" { policy = "write" }`)
        So(err, ShouldBeNil)
        secret, err := vClient.Auth().Token().Create(&vault.TokenCreateRequest{
            Policies:[]string{"admin-test-cesh"}})
        So(err, ShouldBeNil)
        
        testHandler := func() http.HandlerFunc {
            return http.HandlerFunc(CreateEditServiceHandler)
        }
        ts := httptest.NewServer(NeedLogin(NeedAdmin(testHandler())))
        defer ts.Close()
        u.WriteString(string(ts.URL))
        u.WriteString("/admin_api/v1/mkService")
        service := getServicePayload()
        postBody, err := json.Marshal(service)
        So(err, ShouldBeNil)
        req, err := http.NewRequest("POST", u.String(),
            bytes.NewReader(postBody))
        So(err, ShouldBeNil)
        user, err := NewUser("test-acl")
        user.SetExpiresAt(getNow() + 3600 * 1000)
        So(err, ShouldBeNil)
        user.GitHubLogin("criloz", "admin", "delos", "umbrella")
        user.Token = secret.Auth.ClientToken
        req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:user.GenerateJWT()})
        client := &http.Client{}
        response, err := client.Do(req)
        So(err, ShouldBeNil)
        jsonResponse, err := ioutil.ReadAll(response.Body)
        
        So(err, ShouldBeNil)
        umR := AdminServiceResponse{}
        err = json.Unmarshal(jsonResponse, &umR)
        
        So(err, ShouldBeNil)
        So(umR.Roles["devops"].Success, ShouldBeTrue)
        So(umR.Roles["admin"].Success, ShouldBeFalse)
        
    })
}

func Test_LoadLongDescriptionFromUrl(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    
    Convey("if long description url is defined should use it for populate " +
        "long description", t, func(c C) {
        var u bytes.Buffer
        cleanVault()
        testHandler := func() http.HandlerFunc {
            return http.HandlerFunc(CreateEditServiceHandler)
        }
        ts := httptest.NewServer(NeedLogin(NeedAdmin(testHandler())))
        defer ts.Close()
        u.WriteString(string(ts.URL))
        u.WriteString("/admin_api/v1/mkService")
        service := getServicePayload()
        postBody, err := json.Marshal(service)
        So(err, ShouldBeNil)
        req, err := http.NewRequest("POST", u.String(),
            bytes.NewReader(postBody))
        So(err, ShouldBeNil)
        user, err := NewUser("myroot")
        user.SetExpiresAt(getNow() + 3600 * 1000)
        So(err, ShouldBeNil)
        user.GitHubLogin("criloz", "admin", "delos", "umbrella")
        req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:user.GenerateJWT()})
        client := &http.Client{}
        response, err := client.Do(req)
        So(err, ShouldBeNil)
        _, err = ioutil.ReadAll(response.Body)
        So(err, ShouldBeNil)
        cc, err := vault.NewClient(VaultConfig)
        cc.SetToken("myroot")
        So(err, ShouldBeNil)
        secret, err := cc.Logical().Read("secret/kuper/Roles/admin/consul")
        So(err, ShouldBeNil)
        vService := &Service{}
        err = mapstructure.Decode(secret.Data, vService)
        So(err, ShouldBeNil)
        
        So(vService.LongDescription, ShouldNotBeEmpty)
        So(vService.LongDescription, ShouldNotEqual, "dummy_description")
    })
    
    Convey("should return error with a invalid url", t, func(c C) {
        var u bytes.Buffer
        cleanVault()
        testHandler := func() http.HandlerFunc {
            return http.HandlerFunc(CreateEditServiceHandler)
        }
        ts := httptest.NewServer(NeedLogin(NeedAdmin(testHandler())))
        defer ts.Close()
        u.WriteString(string(ts.URL))
        u.WriteString("/admin_api/v1/mkService")
        service := getServicePayload()
        service.LongDescriptionUrl = "invalid_url"
        postBody, err := json.Marshal(service)
        So(err, ShouldBeNil)
        req, err := http.NewRequest("POST", u.String(),
            bytes.NewReader(postBody))
        So(err, ShouldBeNil)
        user, err := NewUser("myroot")
        user.SetExpiresAt(getNow() + 3600 * 1000)
        So(err, ShouldBeNil)
        user.GitHubLogin("criloz", "admin", "delos", "umbrella")
        req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:user.GenerateJWT()})
        client := &http.Client{}
        response, err := client.Do(req)
        So(err, ShouldBeNil)
        jsonResponse, err := ioutil.ReadAll(response.Body)
        
        So(err, ShouldBeNil)
        umR := AdminServiceResponse{}
        err = json.Unmarshal(jsonResponse, &umR)
        So(err, ShouldBeNil)
        So(umR.Success, ShouldBeFalse)
        So(umR.Message, ShouldEqual, "invalid long description url")
    })
    
    Convey("should return error if it cant load the url", t, func(c C) {
        var u bytes.Buffer
        cleanVault()
        testHandler := func() http.HandlerFunc {
            return http.HandlerFunc(CreateEditServiceHandler)
        }
        ts := httptest.NewServer(NeedLogin(NeedAdmin(testHandler())))
        defer ts.Close()
        u.WriteString(string(ts.URL))
        u.WriteString("/admin_api/v1/mkService")
        service := getServicePayload()
        service.LongDescriptionUrl = "http://dummy-123-example.com/readme"
        postBody, err := json.Marshal(service)
        So(err, ShouldBeNil)
        req, err := http.NewRequest("POST", u.String(),
            bytes.NewReader(postBody))
        So(err, ShouldBeNil)
        user, err := NewUser("myroot")
        user.SetExpiresAt(getNow() + 3600 * 1000)
        So(err, ShouldBeNil)
        user.GitHubLogin("criloz", "admin", "delos", "umbrella")
        req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:user.GenerateJWT()})
        client := &http.Client{}
        response, err := client.Do(req)
        So(err, ShouldBeNil)
        jsonResponse, err := ioutil.ReadAll(response.Body)
        So(err, ShouldBeNil)
        umR := AdminServiceResponse{}
        err = json.Unmarshal(jsonResponse, &umR)
        So(err, ShouldBeNil)
        So(umR.Success, ShouldBeFalse)
        So(umR.Message, ShouldEqual, "error quering the provided url")
        
    })
}

func TestDeleteServiceHandler(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    Convey("Should delete a service", t, func() {
        Convey("delete a service with no roles should fail",
            func(c C) {
                var u bytes.Buffer
                cleanVault()
                
                testHandler := func() http.HandlerFunc {
                    return http.HandlerFunc(DeleteServiceHandler)
                }
                ts := httptest.NewServer(NeedLogin(NeedAdmin(testHandler())))
                defer ts.Close()
                u.WriteString(string(ts.URL))
                u.WriteString("/admin_api/v1/deleteService")
                service := getDeleteServicePayload()
                service.Roles = []string{}
                postBody, err := json.Marshal(service)
                So(err, ShouldBeNil)
                req, err := http.NewRequest("POST", u.String(),
                    bytes.NewReader(postBody))
                So(err, ShouldBeNil)
                user, err := NewUser("myroot")
                user.SetExpiresAt(getNow() + 3600 * 1000)
                So(err, ShouldBeNil)
                user.GitHubLogin("criloz", "admin", "delos", "umbrella")
                req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:user.GenerateJWT()})
                client := &http.Client{}
                response, err := client.Do(req)
                So(err, ShouldBeNil)
                jsonResponse, err := ioutil.ReadAll(response.Body)
                So(err, ShouldBeNil)
                umR := struct {
                    Success bool
                    Message string
                }{}
                err = json.Unmarshal(jsonResponse, &umR)
                So(err, ShouldBeNil)
                So(umR.Success, ShouldBeFalse)
                So(umR.Message, ShouldEqual, "Not role was provided")
            })
        
        Convey("Request with not body or bad json should fail",
            func(c C) {
                var u bytes.Buffer
                cleanVault()
                testHandler := func() http.HandlerFunc {
                    return http.HandlerFunc(DeleteServiceHandler)
                }
                ts := httptest.NewServer(NeedLogin(NeedAdmin(testHandler())))
                defer ts.Close()
                u.WriteString(string(ts.URL))
                u.WriteString("/admin_api/v1/deleteService")
                req, err := http.NewRequest("POST", u.String(), nil)
                So(err, ShouldBeNil)
                user, err := NewUser("myroot")
                user.SetExpiresAt(getNow() + 3600 * 1000)
                So(err, ShouldBeNil)
                user.GitHubLogin("criloz", "admin", "delos", "umbrella")
                req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:user.GenerateJWT()})
                client := &http.Client{}
                response, err := client.Do(req)
                So(err, ShouldBeNil)
                jsonResponse, err := ioutil.ReadAll(response.Body)
                So(err, ShouldBeNil)
                umR := struct {
                    Success bool
                    Message string
                }{}
                err = json.Unmarshal(jsonResponse, &umR)
                So(err, ShouldBeNil)
                So(umR.Success, ShouldBeFalse)
                So(umR.Message, ShouldEqual, "Please send a valid json")
            })
        
        Convey("Should delete the service with all the roles",
            func(c C) {
                var u bytes.Buffer
                testHandler := func() http.HandlerFunc {
                    return http.HandlerFunc(DeleteServiceHandler)
                }
                ts := httptest.NewServer(NeedLogin(NeedAdmin(testHandler())))
                defer ts.Close()
                u.WriteString(string(ts.URL))
                u.WriteString("/admin_api/v1/deleteService")
                service := getDeleteServicePayload()
                postBody, err := json.Marshal(service)
                So(err, ShouldBeNil)
                req, err := http.NewRequest("POST", u.String(),
                    bytes.NewReader(postBody))
                So(err, ShouldBeNil)
                user, err := NewUser("myroot")
                user.SetExpiresAt(getNow() + 3600 * 1000)
                So(err, ShouldBeNil)
                user.GitHubLogin("criloz", "admin", "delos", "umbrella")
                req.AddCookie(&http.Cookie{Name:"kuper-jwt",
                    Value:user.GenerateJWT()})
                client := &http.Client{}
                response, err := client.Do(req)
                So(err, ShouldBeNil)
                jsonResponse, err := ioutil.ReadAll(response.Body)
                
                So(err, ShouldBeNil)
                umR := AdminServiceResponse{}
                err = json.Unmarshal(jsonResponse, &umR)
                
                So(err, ShouldBeNil)
                for _, v := range umR.Roles {
                    So(v.Success, ShouldBeTrue)
                }
            })
        
    })
}

func TestDeleteServiceHandler_Permissions(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    
    Convey("Should return permission error if the user can't delete a " +
        "service in the role", t, func(c C) {
        var u bytes.Buffer
        cleanVault()
        vClient, err := vault.NewClient(VaultConfig)
        So(err, ShouldBeNil)
        vClient.SetToken("myroot")
        err = vClient.Sys().PutPolicy("admin-test-cesh", `
        path "secret/kuper/Roles/devops/*" { policy = "write" }
        path "secret/kuper/Admin" { policy = "write" }`)
        So(err, ShouldBeNil)
        secret, err := vClient.Auth().Token().Create(&vault.TokenCreateRequest{
            Policies:[]string{"admin-test-cesh"}})
        So(err, ShouldBeNil)
        
        testHandler := func() http.HandlerFunc {
            return http.HandlerFunc(DeleteServiceHandler)
        }
        ts := httptest.NewServer(NeedLogin(NeedAdmin(testHandler())))
        defer ts.Close()
        u.WriteString(string(ts.URL))
        u.WriteString("/admin_api/v1/mkService")
        service := getDeleteServicePayload()
        postBody, err := json.Marshal(service)
        So(err, ShouldBeNil)
        req, err := http.NewRequest("POST", u.String(),
            bytes.NewReader(postBody))
        So(err, ShouldBeNil)
        user, err := NewUser("test-acl")
        user.SetExpiresAt(getNow() + 3600 * 1000)
        So(err, ShouldBeNil)
        user.GitHubLogin("criloz", "admin", "delos", "umbrella")
        user.Token = secret.Auth.ClientToken
        req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:user.GenerateJWT()})
        client := &http.Client{}
        response, err := client.Do(req)
        So(err, ShouldBeNil)
        jsonResponse, err := ioutil.ReadAll(response.Body)
        
        So(err, ShouldBeNil)
        umR := AdminServiceResponse{}
        err = json.Unmarshal(jsonResponse, &umR)
        
        So(err, ShouldBeNil)
        So(umR.Roles["devops"].Success, ShouldBeTrue)
        So(umR.Roles["admin"].Success, ShouldBeFalse)
        
    })
}

func TestGetServiceHandler(t *testing.T) {
    VaultConfig.Address = "http://localhost:8200"
    Convey("Should get a service by subdomain", t, func() {
        Convey("if service does not exist should return error",
            func(c C) {
                var u bytes.Buffer
                cleanVault()
                r := mux.NewRouter()
                r.HandleFunc("/v1/api/admin/service/{subDomain}", GetServiceHandler)
                ts := httptest.NewServer(NeedLogin(NeedAdmin(r)))
                defer ts.Close()
                u.WriteString(string(ts.URL))
                u.WriteString("/v1/api/admin/service/redis")
                req, err := http.NewRequest("GET", u.String(), nil)
                So(err, ShouldBeNil)
                user, err := NewUser("myroot")
                user.SetExpiresAt(getNow() + 3600 * 1000)
                So(err, ShouldBeNil)
                user.GitHubLogin("criloz", "admin", "delos", "umbrella")
                req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:user.GenerateJWT()})
                client := &http.Client{}
                response, err := client.Do(req)
                So(err, ShouldBeNil)
                jsonResponse, err := ioutil.ReadAll(response.Body)
                So(err, ShouldBeNil)
                umR := struct {
                    Success bool
                    Message string
                }{}
                err = json.Unmarshal(jsonResponse, &umR)
                So(err, ShouldBeNil)
                So(umR.Success, ShouldBeFalse)
                So(umR.Message, ShouldEqual, "service not found")
            })
        
        Convey("Should return  the service with all the roles",
            func(c C) {
                var u bytes.Buffer
                cleanVault()
                populateVault()
                r := mux.NewRouter()
                r.HandleFunc("/v1/api/admin/service/{subDomain}", GetServiceHandler)
                ts := httptest.NewServer(NeedLogin(NeedAdmin(r)))
                defer ts.Close()
                u.WriteString(string(ts.URL))
                u.WriteString("/v1/api/admin/service/redis")
                req, err := http.NewRequest("GET", u.String(), nil)
                So(err, ShouldBeNil)
                user, err := NewUser("myroot")
                user.SetExpiresAt(getNow() + 3600 * 1000)
                So(err, ShouldBeNil)
                user.GitHubLogin("criloz", "admin", "delos", "umbrella")
                req.AddCookie(&http.Cookie{Name:"kuper-jwt", Value:user.GenerateJWT()})
                client := &http.Client{}
                response, err := client.Do(req)
                So(err, ShouldBeNil)
                jsonResponse, err := ioutil.ReadAll(response.Body)
                So(err, ShouldBeNil)
                umR := GetServiceResponse{}
                err = json.Unmarshal(jsonResponse, &umR)
                So(err, ShouldBeNil)
                So(umR.Success, ShouldBeTrue)
                So(umR.Message, ShouldBeEmpty)
            })
        
    })
}

