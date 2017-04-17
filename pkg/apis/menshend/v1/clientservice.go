package v1

import (
    "github.com/emicklei/go-restful"
    "github.com/mitchellh/mapstructure"
    vault "github.com/hashicorp/vault/api"
    "fmt"
    mfilters "github.com/nebtex/menshend/pkg/filters"
    "github.com/nebtex/menshend/pkg/config"
    mutils "github.com/nebtex/menshend/pkg/utils"
)

type listResult struct {
    Keys []string
}

//ClientServiceResource part of the service that is accessible to not admin users
type ClientServiceResource struct {
    Meta                  *ServiceMetadata  `json:"meta"`
    ImpersonateWithinRole bool              `json:"impersonateWithinRole"`
    IsActive              *bool             `json:"isActive"`
    SecretPaths           []string          `json:"secretPaths"`
    FullURL               string            `json:"fullUrl"`
    NeedPortForward       bool              `json:"needPortForward"`
}

//Register ...
func (cs *ClientServiceResource) Register(container *restful.Container) {
    ws := new(restful.WebService).
        Consumes(restful.MIME_JSON).
        Produces(restful.MIME_JSON)
    
    ws.Path("/v1/clientServices").
        Filter(mfilters.LoginFilter).
        Doc("List and search available Services")
    
    ws.Route(ws.GET("").To(cs.listServiceHandler).
        Doc("list services by subdomain").
        Operation("listAvailableServices").
        Param(ws.QueryParameter("subdomain", "service subdomain").DataType("string")).
        Param(ws.QueryParameter("role", "role").DataType("string")).
        Writes([]ClientServiceResource{}))
    
    container.Add(ws)
    
}
//CheckSecretFailIfIsNull panic if vault backend return a empty secret
func CheckSecretFailIfIsNull(s *vault.Secret) {
    if s == nil || s.Data == nil {
        panic(mutils.NotFound)
    }
}

func (cs *ClientServiceResource) listServiceHandler(request *restful.Request, response *restful.Response) {
    user := mfilters.GetTokenFromContext(request)
    subdomain := request.QueryParameter("subdomain")
    role := request.QueryParameter("role")
    ret := []*ClientServiceResource{}
    
    if role != "" && subdomain != "" {
        ValidateSubdomain(subdomain)
        ValidateRole(role)
        key := fmt.Sprintf("%s/roles/%s/%s", config.Config.VaultPath, role, subdomain)
        vaultClient, err := vault.NewClient(vault.DefaultConfig())
        mutils.HttpCheckPanic(err, mutils.InternalError)
        vaultClient.SetToken(user)
        secret, err := vaultClient.Logical().Read(key)
        mutils.HttpCheckPanic(err, mutils.PermissionError)
        CheckSecretFailIfIsNull(secret)
        nService := &ClientServiceResource{}
        aService := &AdminServiceResource{}
        err = mapstructure.Decode(secret.Data, nService)
        mutils.HttpCheckPanic(err, mutils.InternalError.WithUserMessage("error decoding service"))
        err = mapstructure.Decode(secret.Data, aService)
        mutils.HttpCheckPanic(err, mutils.InternalError.WithUserMessage("error decoding service"))
        nService.FullURL = getFullUrl(nService.Meta)
        nService.NeedPortForward = aService.Strategy.PortForward != nil
        ret = append(ret, nService)
        mutils.HttpCheckPanic(response.WriteEntity(ret), mutils.InternalError)
        return
    }
    
    if role != "" {
        mutils.HttpCheckPanic(response.WriteEntity(getServiceByRole(user, role)), mutils.InternalError)
        return
    }
    
    if subdomain != "" {
        mutils.HttpCheckPanic(response.WriteEntity(getServiceBySubdomain(user, subdomain)), mutils.InternalError)
        return
    }
    
    key := fmt.Sprintf("%s/roles", config.Config.VaultPath)
    vaultClient, err := vault.NewClient(vault.DefaultConfig())
    mutils.HttpCheckPanic(err, mutils.InternalError)
    vaultClient.SetToken(user)
    secret, err := vaultClient.Logical().List(key)
    mutils.HttpCheckPanic(err, mutils.PermissionError)
    CheckSecretFailIfIsNull(secret)
    sr := &listResult{}
    err = mapstructure.Decode(secret.Data, sr)
    mutils.HttpCheckPanic(err, mutils.InternalError)
    roleList := sr.Keys
    
    for _, role := range roleList {
        rKey := fmt.Sprintf("%s/roles/%s", config.Config.VaultPath, role)
        rSecret, err := vaultClient.Logical().List(rKey)
        if !(err != nil || rSecret == nil || rSecret.Data == nil) {
            sr := &listResult{}
            err = mapstructure.Decode(rSecret.Data, sr)
            mutils.HttpCheckPanic(err, mutils.InternalError.WithUserMessage("there is something really wrong contact your admin"))
            serviceList := sr.Keys
            for _, service := range serviceList {
                sKey := fmt.Sprintf("%s/roles/%s/%s", config.Config.VaultPath, role, service)
                sSecret, err := vaultClient.Logical().Read(sKey)
                if !(err != nil || sSecret == nil || sSecret.Data == nil) {
                    cs := &ClientServiceResource{}
                    aService := &AdminServiceResource{}
                    err := mapstructure.Decode(sSecret.Data, cs)
                    err = mapstructure.Decode(sSecret.Data, aService)
                    
                    cs.FullURL = getFullUrl(cs.Meta)
                    cs.NeedPortForward = aService.Strategy.PortForward != nil
                    
                    mutils.HttpCheckPanic(err, mutils.InternalError.WithUserMessage("there is something really wrong contact your admin"))
                    ret = append(ret, cs)
                }
            }
        }
    }
    mutils.HttpCheckPanic(response.WriteEntity(ret), mutils.InternalError)
}

func getServiceByRole(user string, role string) []*ClientServiceResource {
    ValidateRole(role)
    key := fmt.Sprintf("%s/roles/%s", config.Config.VaultPath, role)
    vaultClient, err := vault.NewClient(vault.DefaultConfig())
    mutils.HttpCheckPanic(err, mutils.InternalError)
    vaultClient.SetToken(user)
    secret, err := vaultClient.Logical().List(key)
    mutils.HttpCheckPanic(err, mutils.PermissionError)
    CheckSecretFailIfIsNull(secret)
    sr := &listResult{}
    err = mapstructure.Decode(secret.Data, sr)
    mutils.HttpCheckPanic(err, mutils.InternalError)
    sList := sr.Keys
    ret := []*ClientServiceResource{}
    
    for _, subdomain := range sList {
        key := fmt.Sprintf("%s/roles/%s/%s", config.Config.VaultPath, role, subdomain)
        secret, err := vaultClient.Logical().Read(key)
        if !(err != nil || secret == nil || secret.Data == nil) {
            nService := &ClientServiceResource{}
            err = mapstructure.Decode(secret.Data, nService)
            mutils.HttpCheckPanic(err, mutils.InternalError.WithUserMessage("there is something really wrong contact your admin"))
            nService.FullURL = getFullUrl(nService.Meta)
            aService := &AdminServiceResource{}
            err = mapstructure.Decode(secret.Data, aService)
            mutils.HttpCheckPanic(err, mutils.InternalError.WithUserMessage("there is something really wrong contact your admin"))
            nService.NeedPortForward = aService.Strategy.PortForward != nil
            ret = append(ret, nService)
        }
        
    }
    return ret
    
}

func getServiceBySubdomain(user string, subDomain string) []*ClientServiceResource {
    ValidateSubdomain(subDomain)
    key := fmt.Sprintf("%s/roles", config.Config.VaultPath)
    vaultClient, err := vault.NewClient(vault.DefaultConfig())
    mutils.HttpCheckPanic(err, mutils.InternalError)
    vaultClient.SetToken(user)
    secret, err := vaultClient.Logical().List(key)
    mutils.HttpCheckPanic(err, mutils.PermissionError)
    CheckSecretFailIfIsNull(secret)
    sr := &listResult{}
    err = mapstructure.Decode(secret.Data, sr)
    mutils.HttpCheckPanic(err, mutils.InternalError)
    
    roleList := sr.Keys
    
    ret := []*ClientServiceResource{}
    for _, role := range roleList {
        key := fmt.Sprintf("%s/roles/%s/%s", config.Config.VaultPath, role, subDomain)
        secret, err := vaultClient.Logical().Read(key)
        if !(err != nil || secret == nil || secret.Data == nil) {
            nService := &ClientServiceResource{}
            err = mapstructure.Decode(secret.Data, nService)
            mutils.HttpCheckPanic(err, mutils.InternalError.WithUserMessage("there is something really wrong contact your admin"))
            nService.FullURL = getFullUrl(nService.Meta)
            aService := &AdminServiceResource{}
            err = mapstructure.Decode(secret.Data, aService)
            mutils.HttpCheckPanic(err, mutils.InternalError.WithUserMessage("there is something really wrong contact your admin"))
            nService.NeedPortForward = aService.Strategy.PortForward != nil
            
            ret = append(ret, nService)
        }
        
    }
    return ret
}
