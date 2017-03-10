package menshend

import (
    vault "github.com/hashicorp/vault/api"
    "fmt"
    . "github.com/nebtex/menshend/pkg/utils"
    . "github.com/nebtex/menshend/pkg/users"
    . "github.com/nebtex/menshend/pkg/config"
    "github.com/emicklei/go-restful"
)

func GetUserFromRequest(r *restful.Request) *User {
    jwtCookie := r.HeaderParameter("X-Menshend-Token")
    user, err := FromJWT(jwtCookie)
    HttpCheckPanic(err, NotAuthorized)
    return user
}

//GetUserFromContext ...
func GetUserFromContext(r *restful.Request) *User {
    switch v := r.Attribute("user").(type) {
    case *User:
        return v
    default:
        panic(NotAuthorized)
    }
}

//checkAdminPermission ...
func CheckAdminPermission(u *User, vc *vault.Config) {
    key := fmt.Sprintf("%s/%s", Config.VaultPath, "Admin")
    client, err := vault.NewClient(vc)
    HttpCheckPanic(err, InternalError)
    client.SetToken(u.Menshend.VaultToken)
    cap, err := client.Sys().CapabilitiesSelf(key)
    if err != nil {
        panic(PermissionError.Append(err.Error()).WithValue("user", u))
    }
    if !((SliceStringContains(cap, "read")) ||
        (SliceStringContains(cap, "write")) ||
        (SliceStringContains(cap, "update")) ||
        (SliceStringContains(cap, "root"))) {
        panic(PermissionError.WithValue("user", u))
    }
}

//CheckImpersonatePermission check if the user can impersonate other
func CheckImpersonatePermission(u *User, vc *vault.Config) {
    key := fmt.Sprintf("%s/%s", Config.VaultPath, "Impersonate")
    client, err := vault.NewClient(vc)
    CheckPanic(err)
    client.SetToken(u.Menshend.VaultToken)
    cap, err := client.Sys().CapabilitiesSelf(key)
    if err != nil {
        panic(PermissionError.Append(err.Error()).WithValue("user", u))
    }
    if !((SliceStringContains(cap, "read")) ||
        (SliceStringContains(cap, "write")) ||
        (SliceStringContains(cap, "update")) ||
        (SliceStringContains(cap, "root"))) {
        panic(PermissionError.Append(err.Error()).WithValue("user", u))
    }
}


//IsAdmin ...
func IsAdmin(user *User) bool {
    ret := true
    func() {
        defer func() {
            r := recover()
            if r != nil {
                ret = false
            }
        }()
        CheckAdminPermission(user, VaultConfig)
    }()
    return ret
}

//CanImpersonate ...
func CanImpersonate(user *User) bool {
    ret := true
    func() {
        defer func() {
            r := recover()
            if r != nil {
                ret = false
            }
        }()
        CheckImpersonatePermission(user, VaultConfig)
    }()
    return ret
}


//AdminFilter fail if the user is not an admin
func AdminFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
    user := GetUserFromContext(req)
    CheckAdminPermission(user, VaultConfig)
    chain.ProcessFilter(req, resp)
}


//LoginFilter fail if the user is not logged in
func LoginFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
    user := GetUserFromRequest(req)
    req.SetAttribute("user", user)
    chain.ProcessFilter(req, resp)
}


//ImpersonateFilter ....
func ImpersonateFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
    user := GetUserFromContext(req)
    CheckImpersonatePermission(user, VaultConfig)
    chain.ProcessFilter(req, resp)
}
