package kuper

import (
    "fmt"
    "net/http"
    "encoding/json"
    "net/url"
    "github.com/gorilla/mux"
    "io/ioutil"
    "github.com/fatih/structs"
    vault "github.com/hashicorp/vault/api"
    "strings"
    "github.com/mitchellh/mapstructure"
)

type ServiceRole struct {
    LuaScript             string `json:"luaScript"`
    ImpersonateWithinRole bool `json:"impersonateWithinRole"`
    Proxy                 bool `json:"proxy"`
    IsActive              bool `json:"isActive"`
}

type ServicePayload struct {
    SubDomain          string `json:"subDomain"`
    Logo               string `json:"logo"`
    Name               string `json:"name"`
    ShortDescription   string `json:"shortDescription"`
    LongDescription    string `json:"longDescription"`
    LongDescriptionUrl string `json:"longDescriptionUrl"`
    Roles              map[string]*ServiceRole `json:"roles"`
}

type DeleteServicePayload struct {
    SubDomain string `json:"subDomain"`
    Roles     []string `json:"roles"`
}

type Response struct {
    Success bool  `json:"success"`
    Message string  `json:"message"`
}

type AdminServiceResponse struct {
    Success bool  `json:"success"`
    Message string  `json:"message"`
    Roles   map[string]Response `json:"roles"`
}

type GetServicePayload struct {
    SubDomain string  `json:"subDomain"`
}

func getReadme(url string) ([]byte, error) {
    r, err := http.Get(url)
    if err != nil {
        return []byte{}, err
    }
    defer r.Body.Close()
    return ioutil.ReadAll(r.Body)
}

//CreateServiceHandler create a new service
func CreateEditServiceHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
    user := GetUserFromContext(r.Context())
    
    csp := &ServicePayload{}
    err := json.NewDecoder(r.Body).Decode(csp)
    if err != nil {
        errMsg := `{"success":false, "message": "Please send a valid json"}`
        w.Write([]byte(errMsg))
        return
    }
    
    if len(csp.Roles) == 0 {
        errMsg := `{"success":false, "message": "Not role was provided"}`
        w.Write([]byte(errMsg))
        return
    }
    if csp.LongDescriptionUrl != "" {
        _, parseErr := url.ParseRequestURI(csp.LongDescriptionUrl)
        if parseErr != nil {
            errMsg := `{"success":false, "message": "invalid long description url"}`
            w.Write([]byte(errMsg))
            return
        }
        r, gErr := getReadme(csp.LongDescriptionUrl)
        if gErr != nil {
            errMsg := `{"success":false, "message": "error quering the provided url"}`
            w.Write([]byte(errMsg))
            return
        }
        csp.LongDescription = string(r)
    }
    
    response := AdminServiceResponse{Roles:map[string]Response{}, Success:true}
    vaultClient, err := vault.NewClient(VaultConfig)
    CheckPanic(err)
    vaultClient.SetToken(user.Token)
    
    for role, val := range csp.Roles {
        newS := &Service{}
        newS.Logo = csp.Logo
        newS.Name = csp.Name
        newS.ShortDescription = csp.ShortDescription
        newS.LongDescription = csp.LongDescription
        newS.LuaScript = val.LuaScript
        newS.ImpersonateWithinRole = val.ImpersonateWithinRole
        newS.Proxy = val.Proxy
        newS.IsActive = val.IsActive
        key := fmt.Sprintf("%s/Roles/%s/%s", Config.VaultPath,
            role, csp.SubDomain)
        _, vaultErr := vaultClient.Logical().Write(key, structs.Map(newS))
        if (vaultErr != nil) {
            response.Roles[role] = Response{Success: false,
                Message: "Permission error"}
        } else {
            response.Roles[role] = Response{Success: true}
        }
    }
    
    data, err := json.Marshal(response)
    
    CheckPanic(err)
    w.Write(data)
}

//DeleteServiceHandler delete a  service
func DeleteServiceHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
    user := GetUserFromContext(r.Context())
    
    dsp := &DeleteServicePayload{}
    err := json.NewDecoder(r.Body).Decode(dsp)
    if err != nil {
        errMsg := `{"success":false, "message": "Please send a valid json"}`
        w.Write([]byte(errMsg))
        return
    }
    
    if len(dsp.Roles) == 0 {
        errMsg := `{"success":false, "message": "Not role was provided"}`
        w.Write([]byte(errMsg))
        return
    }
    
    response := AdminServiceResponse{Roles:map[string]Response{}, Success:true}
    vaultClient, err := vault.NewClient(VaultConfig)
    CheckPanic(err)
    vaultClient.SetToken(user.Token)
    
    for _, role := range dsp.Roles {
        key := fmt.Sprintf("%s/Roles/%s/%s", Config.VaultPath,
            role, dsp.SubDomain)
        _, vaultErr := vaultClient.Logical().Delete(key)
        
        if (vaultErr != nil) {
            response.Roles[role] = Response{Success: false,
                Message: "Permission error"}
        } else {
            response.Roles[role] = Response{Success: true}
        }
    }
    
    data, err := json.Marshal(response)
    
    CheckPanic(err)
    w.Write(data)
}

//checkAdminPermission ...
func checkAdminPermission(u *User, vc *vault.Config) {
    key := fmt.Sprintf("%s/%s", Config.VaultPath, "Admin")
    client, err := vault.NewClient(vc)
    CheckPanic(err)
    client.SetToken(u.Token)
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

//NeedAdmin admin middleware
func NeedAdmin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := GetUserFromContext(r.Context())
        checkAdminPermission(user, VaultConfig)
        next.ServeHTTP(w, r)
    })
}

type GetServiceResponse struct {
    Success bool  `json:"success"`
    Message string  `json:"message"`
    Service *ServicePayload
}


//GetServiceHandler return a service within all the roles
// available to the current user
func GetServiceHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
    user := GetUserFromContext(r.Context())
    vars := mux.Vars(r)
    subDomain := vars["subDomain"]
    vc, err := vault.NewClient(VaultConfig)
    vc.SetToken(user.Token)
    CheckPanic(err)
    key := fmt.Sprintf("%s/Roles", Config.VaultPath)
    secret, vaultErr := vc.Logical().List(key)
    if vaultErr != nil {
        if strings.Contains(vaultErr.Error(), "403") {
            errMsg := `{"success":false, "message": "permission error"}`
            w.Write([]byte(errMsg))
            return
        }
        CheckPanic(vaultErr)
    }
    if (secret == nil) || (secret.Data == nil) {
        errMsg := `{"success":false, "message": "service not found"}`
        w.Write([]byte(errMsg))
        return
    }
    type ListResult struct {
        Keys []string
    }
    
    rr := &ListResult{}
    err = mapstructure.Decode(secret.Data, rr)
    CheckPanic(err)
    roleList := rr.Keys
    serviceFound := &ServicePayload{}
    serviceFound.SubDomain = subDomain
    serviceFound.Roles = map[string]*ServiceRole{}
    
    for _, role := range roleList {
        if !strings.HasSuffix(role, "/") {
            continue
        }
        sKey := fmt.Sprintf("%s/Roles/%s/%s", Config.VaultPath, role, subDomain)
        sSecret, err := vc.Logical().Read(sKey)
        
        if err != nil {
            continue
        }
        if (sSecret == nil) || (sSecret.Data == nil) {
            continue
        }
        nService := &Service{}
        err = mapstructure.Decode(sSecret.Data, nService)
        CheckPanic(err)
        serviceFound.Name = nService.Name
        serviceFound.Logo = nService.Logo
        serviceFound.LongDescription = nService.LongDescription
        serviceFound.ShortDescription = nService.ShortDescription
        serviceFound.Roles[role] = &ServiceRole{}
        serviceFound.Roles[role].ImpersonateWithinRole = nService.ImpersonateWithinRole
        serviceFound.Roles[role].IsActive = nService.IsActive
        serviceFound.Roles[role].LuaScript = nService.LuaScript
        serviceFound.Roles[role].Proxy = nService.Proxy
        
    }
    response := &GetServiceResponse{}
    response.Success = true
    response.Service = serviceFound
    data, err := json.Marshal(response)
    CheckPanic(err)
    w.Write(data)
}
