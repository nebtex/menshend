package v1

import (
    "github.com/emicklei/go-restful"
    "github.com/mitchellh/mapstructure"
    vault "github.com/hashicorp/vault/api"
    "fmt"
    . "github.com/nebtex/menshend/pkg/apis/menshend"
    . "github.com/nebtex/menshend/pkg/config"
    . "github.com/nebtex/menshend/pkg/utils"
    . "github.com/nebtex/menshend/pkg/users"
)

type ListResult struct {
    Keys []string
}
//Service service definition struct
type ClientServiceResource struct {
    ID                    string `json:"id"`
    RoleID                string `json:"roleId"`
    SubDomain             string `json:"subDomain"`
    Logo                  string `json:"logo"`
    Name                  string `json:"name"`
    ShortDescription      string `json:"shortDescription"`
    LongDescription       string `json:"longDescription"`
    ImpersonateWithinRole bool   `json:"impersonateWithinRole"`
    SecretPaths           []string `json:"secretPaths"`
}

func (cs *ClientServiceResource) Register(container *restful.Container) {
    ws := new(restful.WebService).
        Consumes(restful.MIME_JSON).
        Produces(restful.MIME_JSON)
    
    ws.Path("/v1/clientServices").
        Filter(LoginFilter).
        Doc("List and search available Services")
    
    ws.Route(ws.GET("").To(cs.listServiceHandler).
        Doc("list services by subdomain").
        Operation("listAvailableServices").
        Param(ws.QueryParameter("subdomain", "service subdomain").DataType("string")).
        Param(ws.QueryParameter("role", "role").DataType("string")).
        Writes([]ClientServiceResource{}))
    
    container.Add(ws)
    
}
func CheckSecretFailIfIsNull(s *vault.Secret) {
    if s == nil || s.Data == nil {
        panic(NotFound)
    }
}

func (cs *ClientServiceResource) listServiceHandler(request *restful.Request, response *restful.Response) {
    user := GetUserFromContext(request)
    subdomain := request.QueryParameter("subdomain")
    role := request.QueryParameter("role")
    ret := []*ClientServiceResource{}
    
    if role != "" && subdomain != "" {
        ValidateSubdomain(subdomain)
        ValidateRole(role)
        key := fmt.Sprintf("%s/roles/%s/%s", Config.VaultPath, role, subdomain)
        vaultClient, err := vault.NewClient(VaultConfig)
        HttpCheckPanic(err, InternalError)
        vaultClient.SetToken(user.Menshend.VaultToken)
        secret, err := vaultClient.Logical().Read(key)
        HttpCheckPanic(err, PermissionError)
        CheckSecretFailIfIsNull(secret)
        nService := &ClientServiceResource{}
        err = mapstructure.Decode(secret.Data, nService)
        HttpCheckPanic(err, InternalError.Append("error decoding service"))
        ret = append(ret, nService)
        response.WriteEntity(ret)
        return
    }
    
    if role != "" {
        response.WriteEntity(getServiceByRole(user, role))
        return
    }
    
    if subdomain != "" {
        response.WriteEntity(getServiceBySubdomain(user, subdomain))
        return
    }
    
    key := fmt.Sprintf("%s/roles", Config.VaultPath)
    vaultClient, err := vault.NewClient(VaultConfig)
    HttpCheckPanic(err, InternalError)
    vaultClient.SetToken(user.Menshend.VaultToken)
    secret, err := vaultClient.Logical().List(key)
    HttpCheckPanic(err, PermissionError)
    CheckSecretFailIfIsNull(secret)
    sr := &ListResult{}
    err = mapstructure.Decode(secret.Data, sr)
    HttpCheckPanic(err, InternalError)
    roleList := sr.Keys
    
    for _, role := range roleList {
        rKey := fmt.Sprintf("%s/roles/%s", Config.VaultPath, role)
        rSecret, err := vaultClient.Logical().List(rKey)
        if !(err != nil || rSecret == nil || rSecret.Data == nil) {
            sr := &ListResult{}
            err = mapstructure.Decode(rSecret.Data, sr)
            HttpCheckPanic(err, InternalError.Append("there is something really wrong contact your admin"))
            serviceList := sr.Keys
            for _, service := range serviceList {
                sKey := fmt.Sprintf("%s/roles/%s/%s", Config.VaultPath, role, service)
                sSecret, err := vaultClient.Logical().Read(sKey)
                if !(err != nil || sSecret == nil || sSecret.Data == nil) {
                    cs:=&ClientServiceResource{}
                    err := mapstructure.Decode(sSecret.Data, cs)
                    HttpCheckPanic(err, InternalError.Append("there is something really wrong contact your admin"))
                    ret = append(ret, cs)
                }
            }
        }
    }
    response.WriteEntity(ret)
}

func getServiceByRole(user *User, role string) []*ClientServiceResource {
    ValidateRole(role)
    key := fmt.Sprintf("%s/roles/%s", Config.VaultPath, role)
    vaultClient, err := vault.NewClient(VaultConfig)
    HttpCheckPanic(err, InternalError)
    vaultClient.SetToken(user.Menshend.VaultToken)
    secret, err := vaultClient.Logical().List(key)
    HttpCheckPanic(err, PermissionError)
    CheckSecretFailIfIsNull(secret)
    sr := &ListResult{}
    err = mapstructure.Decode(secret.Data, sr)
    HttpCheckPanic(err, InternalError)
    sList := sr.Keys
    ret := []*ClientServiceResource{}
    
    for _, subdomain := range sList {
        key := fmt.Sprintf("%s/roles/%s/%s", Config.VaultPath, role, subdomain)
        secret, err := vaultClient.Logical().Read(key)
        if !(err != nil || secret == nil || secret.Data == nil) {
            nService := &ClientServiceResource{}
            err = mapstructure.Decode(secret.Data, nService)
            HttpCheckPanic(err, InternalError.Append("there is something really wrong contact your admin"))
            ret = append(ret, nService)
        }
        
    }
    return ret
    
}

func getServiceBySubdomain(user *User, subDomain string) []*ClientServiceResource {
    ValidateSubdomain(subDomain)
    key := fmt.Sprintf("%s/roles", Config.VaultPath)
    vaultClient, err := vault.NewClient(VaultConfig)
    HttpCheckPanic(err, InternalError)
    vaultClient.SetToken(user.Menshend.VaultToken)
    secret, err := vaultClient.Logical().List(key)
    HttpCheckPanic(err, PermissionError)
    CheckSecretFailIfIsNull(secret)
    sr := &ListResult{}
    err = mapstructure.Decode(secret.Data, sr)
    HttpCheckPanic(err, InternalError)
    
    roleList := sr.Keys
    
    ret := []*ClientServiceResource{}
    for _, role := range roleList {
        key := fmt.Sprintf("%s/roles/%s/%s", Config.VaultPath, role, subDomain)
        secret, err := vaultClient.Logical().Read(key)
        if !(err != nil || secret == nil || secret.Data == nil) {
            nService := &ClientServiceResource{}
            err = mapstructure.Decode(secret.Data, nService)
            HttpCheckPanic(err, InternalError.Append("there is something really wrong contact your admin"))
            ret = append(ret, nService)
        }
        
    }
    return ret
    
}
