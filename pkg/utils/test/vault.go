package test

import (
    "fmt"
    vault "github.com/hashicorp/vault/api"
    "github.com/mitchellh/mapstructure"
    . "github.com/nebtex/menshend/pkg/config"
    . "github.com/nebtex/menshend/pkg/utils"
    "strings"
    "github.com/fatih/structs"
)

func CleanVault() {
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

type Role map[string]*Service

//Service service definition struct
type Service struct {
    Logo                  string
    SubDomain             string
    Name                  string
    ShortDescription      string
    LongDescription       string
    ProxyCode             string
    ProxyLanguage         int
    ImpersonateWithinRole bool
    Proxy                 bool
    IsActive              bool
    SecretPaths           []string
}

func PopulateVault() {
    vc, err := vault.NewClient(VaultConfig)
    vc.SetToken("myroot")
    CheckPanic(err)
    roles := map[string]Role{
        "ml-team": map[string]*Service{
            "consul.":{IsActive:true, SubDomain: "consul.", ImpersonateWithinRole: true},
            "gitlab.":{IsActive:false, SecretPaths:[]string{"secret/gitlab/password", Config.VaultPath + "/roles/ml-team/gitlab."},
                ProxyCode:`
    function getBackend ()
        return "http://gitlab"
    end
    `},
            "postgres.":{},
            "redis.":{}},
        "admin":map[string]*Service{
            "kubernetes":{IsActive:true,
                ProxyCode:`
    function getBackend ()
        return "invalid_url"
    end
    `},
            "vault.":{},
            "redis.":{IsActive:true, ProxyCode:`
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
