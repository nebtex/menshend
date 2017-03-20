package menshend

import (
    vault "github.com/hashicorp/vault/api"
    "fmt"
    . "github.com/nebtex/menshend/pkg/utils"
    . "github.com/nebtex/menshend/pkg/config"
    "github.com/emicklei/go-restful"
    "strings"
)

func parseBearerAuth(auth string) (string, bool) {
    const prefix = "Bearer "
    if !strings.HasPrefix(auth, prefix) {
        return "", false
    }
    return auth[len(prefix):], true
}

func GetTokenFromRequest(r *restful.Request) string {
    bearerToken, _ := parseBearerAuth(r.Request.Header.Get("Authorization"))
    vaultToken := r.HeaderParameter("X-Vault-Token")
    if bearerToken != "" {
        if vaultToken != "" {
            vaultToken = bearerToken
        }
    }
    
    return vaultToken
}

//GetUserFromContext ...
func GetTokenFromContext(r *restful.Request) string {
    switch v := r.Attribute("VaultToken").(type) {
    case string:
        return v
    default:
        panic(NotAuthorized)
    }
}

//checkAdminPermission ...
func CheckAdminPermission(vaultToken string, vc *vault.Config) {
    key := fmt.Sprintf("%s/%s", Config.VaultPath, "Admin")
    client, err := vault.NewClient(vc)
    HttpCheckPanic(err, InternalError)
    client.SetToken(vaultToken)
    cap, err := client.Sys().CapabilitiesSelf(key)
    if err != nil {
        panic(PermissionError.Append(err.Error()))
    }
    if !((SliceStringContains(cap, "read")) ||
        (SliceStringContains(cap, "write")) ||
        (SliceStringContains(cap, "update")) ||
        (SliceStringContains(cap, "root"))) {
        panic(PermissionError)
    }
}

//CheckImpersonatePermission check if the user can impersonate other
func CheckImpersonatePermission(vaultToken string, vc *vault.Config) {
    key := fmt.Sprintf("%s/%s", Config.VaultPath, "Impersonate")
    client, err := vault.NewClient(vc)
    CheckPanic(err)
    client.SetToken(vaultToken)
    cap, err := client.Sys().CapabilitiesSelf(key)
    if err != nil {
        panic(PermissionError.Append(err.Error()))
    }
    if !((SliceStringContains(cap, "read")) ||
        (SliceStringContains(cap, "write")) ||
        (SliceStringContains(cap, "update")) ||
        (SliceStringContains(cap, "root"))) {
        panic(PermissionError.Append(err.Error()))
    }
}


//IsAdmin ...
func IsAdmin(vaultToken string) bool {
    ret := true
    func() {
        defer func() {
            r := recover()
            if r != nil {
                ret = false
            }
        }()
        CheckAdminPermission(vaultToken, vault.DefaultConfig())
    }()
    return ret
}

//CanImpersonate ...
func CanImpersonate(vaultToken string) bool {
    ret := true
    func() {
        defer func() {
            r := recover()
            if r != nil {
                ret = false
            }
        }()
        CheckImpersonatePermission(vaultToken, vault.DefaultConfig())
    }()
    return ret
}


//AdminFilter fail if the user is not an admin
func AdminFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
    CheckAdminPermission(GetTokenFromContext(req), vault.DefaultConfig())
    chain.ProcessFilter(req, resp)
}


//LoginFilter fail if the user is not logged in
func LoginFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
    vaultToken := GetTokenFromRequest(req)
    if vaultToken==""{
        panic(NotAuthorized)
    }
    req.SetAttribute("VaultToken", vaultToken)
    chain.ProcessFilter(req, resp)
}


//ImpersonateFilter ....
func ImpersonateFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
    CheckImpersonatePermission(GetTokenFromContext(req), vault.DefaultConfig())
    chain.ProcessFilter(req, resp)
}
