package kuper

import (
    "net/http"
    //"strings"
    "encoding/json"
    vault "github.com/hashicorp/vault/api"
    "fmt"
    "github.com/mitchellh/mapstructure"
    "strings"
    "github.com/Sirupsen/logrus"
)

//ServiceListHandler list all the services, that are allowed to the
// current user
func ServiceListHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
    type ListResult struct {
        Keys []string
    }
    allRoles := map[string]Role{}
    user := GetUserFromContext(r.Context())
    vc, err := vault.NewClient(VaultConfig)
    vc.SetToken(user.Token)
    CheckPanic(err)
    key := fmt.Sprintf("%s/Roles", Config.VaultPath)
    secret, vaultErr := vc.Logical().List(key)
    if vaultErr != nil {
        if strings.Contains(vaultErr.Error(), "403") {
            w.Write([]byte("{}"))
            return
        }
        CheckPanic(vaultErr)
    }
    if (secret == nil) || (secret.Data == nil) {
        w.Write([]byte("{}"))
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
        rKey := fmt.Sprintf("%s/Roles/%s", Config.VaultPath, role)
        rSecret, err := vc.Logical().List(rKey)
        if err != nil {
            continue
        }

        sr := &ListResult{}
        err = mapstructure.Decode(rSecret.Data, sr)
        CheckPanic(err)

        serviceList := sr.Keys

        for _, service := range serviceList {
            sKey := fmt.Sprintf("%s/Roles/%s/%s", Config.VaultPath, role, service)
            sSecret, consulErr := vc.Logical().Read(sKey)
            if consulErr != nil {
                continue
            }
            if _, ok := allRoles[role]; !ok {
                allRoles[role] = Role{}
            }
            if _, ok := allRoles[role][service]; !ok {
                allRoles[role][service] = &Service{}
            }
            if sSecret.Data != nil {
                err := mapstructure.Decode(sSecret.Data, allRoles[role][service])
                CheckPanic(err)
            }

        }
    }

    if len(allRoles) == 0 {
        w.Write([]byte("{}"))
        return
    }

    type oService struct {
        Logo             string
        Name             string
        ShortDescription string
        LongDescription  string
        Subdomain        string
        Roles            []string
    }
    response := map[string]*oService{}

    for role, services := range allRoles {
        for subDomain, service := range services {
            if _, ok := response[subDomain]; !ok {
                response[subDomain] = &oService{
                    Logo: service.Logo,
                    Name: service.Name,
                    ShortDescription: service.ShortDescription,
                    LongDescription: service.LongDescription,
                    Subdomain:subDomain,
                    Roles: []string{role},
                }
            } else {
                response[subDomain].Roles = append(response[subDomain].Roles, role)
            }
        }
    }
    data, err := json.Marshal(response)

    CheckPanic(err)
    w.Write(data)
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
        checkAdminPermission(user, VaultConfig)
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
        checkImpersonatePermission(user, VaultConfig)
    }()
    return ret
}


type LoginStatusResponse struct {
    IsLogged         bool
    IsAdmin          bool
    CanImpersonate   bool
    SessionExpiresAt int64
}

//LoginStatusHandler ...
func LoginStatusHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)

    user, err := getUserFromRequest(r)
    if err != nil {
        logrus.Error(err)
    }
    response := &LoginStatusResponse{}

    if user != nil {
        response.IsLogged = true
        response.IsAdmin = IsAdmin(user)
        response.CanImpersonate = CanImpersonate(user)
        response.SessionExpiresAt = user.ExpiresAt
    }

    data, jsonErr := json.Marshal(response)
    CheckPanic(jsonErr)
    w.Write(data)

}
