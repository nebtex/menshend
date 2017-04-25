package test

import (
    "fmt"
    vault "github.com/hashicorp/vault/api"
    "github.com/mitchellh/mapstructure"
    mconfig "github.com/nebtex/menshend/pkg/config"
    mutils "github.com/nebtex/menshend/pkg/utils"
    "strings"
    "github.com/fatih/structs"
    "github.com/nebtex/menshend/pkg/strategy"
    "github.com/nebtex/menshend/pkg/resolvers"
)

func CleanVault() {
    type ListResult struct {
        Keys []string
    }
    vc, err := vault.NewClient(vault.DefaultConfig())
    mutils.CheckPanic(err)
    vc.SetToken("myroot")
    
    key := fmt.Sprintf("%s/roles", mconfig.Config.VaultPath)
    secret, err := vc.Logical().List(key)
    mutils.CheckPanic(err)
    if (secret == nil) || (secret.Data == nil) {
        return
    }
    rr := &ListResult{}
    err = mapstructure.Decode(secret.Data, rr)
    mutils.CheckPanic(err)
    roleList := rr.Keys
    
    for _, role := range roleList {
        if !strings.HasSuffix(role, "/") {
            continue
        }
        rKey := fmt.Sprintf("%s/roles/%s", mconfig.Config.VaultPath, role)
        rSecret, err := vc.Logical().List(rKey)
        if err != nil {
            continue
        }
        if (rSecret == nil) || (rSecret.Data == nil) {
            return
        }
        
        sr := &ListResult{}
        err = mapstructure.Decode(rSecret.Data, sr)
        
        mutils.CheckPanic(err)
        
        serviceList := sr.Keys
        
        for _, service := range serviceList {
            sKey := fmt.Sprintf("%s/roles/%s/%s", mconfig.Config.VaultPath, role, service)
            _, err := vc.Logical().Delete(sKey)
            mutils.CheckPanic(err)
        }
    }
}

type Role map[string]*Service

type TestServiceStrategy struct {
    Proxy       *strategy.Proxy `json:"proxy"`
    PortForward *strategy.PortForward `json:"portForward"`
    Redirect    *strategy.Redirect `json:"redirect"`
}

type TestServiceMetadata struct {
    ID        string `json:"id"`
    RoleID    string `json:"roleId"`
    SubDomain string `json:"subDomain"`
    Name      string `json:"name"`
}

type TestServiceCache struct {
    // time to live seconds
    TTL    int `json:"ttl"`
    Active bool `json:"active"`
}
type TestServiceResolver struct {
    Yaml  *resolvers.YAMLResolver `json:"yaml"`
    Lua   *resolvers.LuaResolver `json:"lua"`
    Cache *TestServiceCache `json:"cache"`
}

//Service service definition struct
type Service struct {
    Meta                  *TestServiceMetadata `json:"meta"`
    Logo                  string `json:"logo"`
    ShortDescription      string `json:"shortDescription"`
    LongDescription       string
    ProxyCode             string
    ProxyLanguage         string
    ImpersonateWithinRole bool
    Proxy                 bool
    IsActive              bool
    SecretPaths           []string
    CSRF                  bool
    Strategy              *TestServiceStrategy `json:"strategy"`
    Resolver              *TestServiceResolver `json:"resolver"`
}

func PopulateVault() {
    vc, err := vault.NewClient(vault.DefaultConfig())
    vc.SetToken("myroot")
    mutils.CheckPanic(err)
    roles := map[string]Role{
        "ml-team": map[string]*Service{
            "consul.":{IsActive:true,
                Meta: &TestServiceMetadata{SubDomain: "consul."},
                ImpersonateWithinRole: true,
                CSRF: true,
                Strategy: &TestServiceStrategy{Proxy:&strategy.Proxy{}},
                Resolver: &TestServiceResolver{Lua: &resolvers.LuaResolver{Content: `function getBackend (tokenInfo, request)
    return "http://localhost:5454", {}
end`}}, },
            "consul-2.":{IsActive:true,
                Meta: &TestServiceMetadata{SubDomain: "consul-2."},
                ImpersonateWithinRole: true,
                CSRF: false,
                Strategy: &TestServiceStrategy{Proxy:&strategy.Proxy{}},
                Resolver: &TestServiceResolver{Lua: &resolvers.LuaResolver{Content: `function getBackend (tokenInfo, request)
    return "http://localhost:5454", {}
end`}}, },
            "gitlab.":{
                Meta: &TestServiceMetadata{SubDomain: "consul-2."},
                IsActive:false, SecretPaths:[]string{"secret/gitlab/password", mconfig.Config.VaultPath + "/roles/ml-team/gitlab."},
            },
            "postgres.":{},
            "redis.":{IsActive:true, Meta: &TestServiceMetadata{SubDomain: "redis."}, ShortDescription: "redisdb"}, },
        "admin":map[string]*Service{
            "kubernetes":{IsActive:true},
            "vault.":{},
            "redis.":{IsActive:true, Meta: &TestServiceMetadata{SubDomain: "redis."}, ShortDescription: "redisdb"},
        }}
    for role, services := range roles {
        for service, val := range services {
            key := fmt.Sprintf("%s/roles/%s/%s", mconfig.Config.VaultPath, role, service)
            _, err := vc.Logical().Write(key, structs.Map(val))
            mutils.CheckPanic(err)
        }
    }
}
