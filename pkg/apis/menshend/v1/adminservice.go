package v1

import (
    "net/url"
    "io/ioutil"
    "net/http"
    "github.com/emicklei/go-restful"
    "github.com/mitchellh/mapstructure"
    "fmt"
    . "github.com/nebtex/menshend/pkg/apis/menshend"
    . "github.com/nebtex/menshend/pkg/config"
    . "github.com/nebtex/menshend/pkg/utils"
    "github.com/fatih/structs"
    "strings"
    vault "github.com/hashicorp/vault/api"
    "github.com/nebtex/menshend/pkg/strategy"
    "github.com/nebtex/menshend/pkg/resolvers"
)

func init() {
    restful.PrettyPrintResponses = false
}

type ServiceCache struct {
    // time to live seconds
    TTL    int `json:"ttl"`
    Active bool `json:"active"`
}

type ServiceStrategy struct {
    Proxy       *strategy.Proxy `json:"proxy"`
    PortForward *strategy.PortForward `json:"portForward"`
    Redirect    *strategy.Redirect `json:"redirect"`
}

func (ss *ServiceStrategy) Validate() {
    defined := 0
    if ss.Proxy != nil {
        defined += 1
    }
    if ss.PortForward != nil {
        defined += 1
    }
    if ss.Redirect != nil {
        defined += 1
    }
    
    if defined != 1 {
        panic(BadRequest.WithUserMessage("Please define only one strategy, you have not define one!! or have multiples defined"))
    }
}

func (ss *ServiceStrategy)Get() strategy.Strategy {
    if ss.Proxy != nil {
        return ss.Proxy
    }
    if ss.PortForward != nil {
        return ss.PortForward
    }
    return ss.Redirect
}

type ServiceResolver struct {
    Yaml  *resolvers.YAMLResolver `json:"yaml"`
    Lua   *resolvers.LuaResolver `json:"lua"`
    Cache *ServiceCache `json:"cache"`
}

func (sr *ServiceResolver) Validate() {
    defined := 0
    if sr.Yaml != nil {
        defined += 1
    }
    if sr.Lua != nil {
        defined += 1
    }
    
    if defined != 1 {
        panic(BadRequest.WithUserMessage("Please define only one resolver, you have not define one!! or have multiples defined"))
    }
}
func (sr *ServiceResolver) Get() resolvers.Resolver {
    if sr.Lua != nil {
        return sr.Lua
    }
    return sr.Yaml
}

type ServiceMetadata struct {
    ID        string `json:"id"`
    RoleID    string `json:"roleId"`
    SubDomain string `json:"subDomain"`
    Name      string `json:"name"`
}

//Service service definition struct
type AdminServiceResource struct {
    Meta                  *ServiceMetadata  `json:"meta"`
    Logo                  string `json:"logo"`
    ShortDescription      string `json:"shortDescription"`
    LongDescription       string `json:"longDescription"`
    LongDescriptionUrl    string `json:"longDescriptionUrl"`
    Resolver              *ServiceResolver `json:"Resolver"`
    ImpersonateWithinRole bool   `json:"impersonateWithinRole"`
    Strategy              *ServiceStrategy `json:"strategy"`
    IsActive              *bool `json:"isActive"`
    SecretPaths           []string `json:"secretPaths"`
}

func getReadme(url string) ([]byte, error) {
    r, err := http.Get(url)
    if err != nil {
        return []byte{}, err
    }
    defer r.Body.Close()
    return ioutil.ReadAll(r.Body)
}

func (s *AdminServiceResource) LoadLongDescriptionUrl() error {
    if s.LongDescriptionUrl != "" {
        _, err := url.ParseRequestURI(s.LongDescriptionUrl)
        if err != nil {
            return err
        }
        r, err := getReadme(s.LongDescriptionUrl)
        if err != nil {
            return err
        }
        s.LongDescription = string(r)
    }
    return nil
}

func (as *AdminServiceResource) Register(container *restful.Container) {
    ws := new(restful.WebService).
        Consumes(restful.MIME_JSON).
        Produces(restful.MIME_JSON)
    
    ws.Path("/v1/adminServices").
        Filter(LoginFilter).
        Filter(AdminFilter).
        Doc("Manage Services")
    
    ws.Route(ws.GET("").To(as.listServiceHandler).
        Doc("list services by subdomain").
        Operation("listServicesBySubdomain").
        Param(ws.QueryParameter("subdomain", "service subdomain").DataType("string").Required(true)).
        Writes([]AdminServiceResource{}))
    
    ws.Route(ws.GET("/{id:*}").To(as.getService).
        Doc("get a service with all the available info about it").
        Operation("getService").
        Param(ws.PathParameter("id", "identifier of the service").DataType("string")).
        Writes(AdminServiceResource{}))
    
    ws.Route(ws.DELETE("/{id:*}").To(as.deleteServiceHandler).
        Doc("delete a service").
        Operation("deleteService").
        Param(ws.PathParameter("id", "unique identifier of the service").DataType("string")))
    
    ws.Route(ws.PUT("/{id:*}").To(as.putServiceHandler).
        Doc("update or create a service").
        Operation("updateService").
        Param(ws.PathParameter("id", "unique dentifier of the service").DataType("string")).
        Reads(AdminServiceResource{}))
    
    container.Add(ws)
    
}
func getRoleAndSubdomain(id string) (role string, subdomain string) {
    items := strings.Split(id, "/")
    role = items[1]
    ValidateRole(role)
    subdomain = items[2]
    ValidateSubdomain(subdomain)
    return role, subdomain
}

func (as *AdminServiceResource) putServiceHandler(request *restful.Request, response *restful.Response) {
    user := GetTokenFromContext(request)
    serviceId := request.PathParameter("id")
    ValidateService(serviceId)
    nService := &AdminServiceResource{}
    err := request.ReadEntity(nService)
    HttpCheckPanic(err, BadRequest.WithUserMessage("error decoding request"))
    
    if nService.Resolver == nil {
        panic(BadRequest.WithUserMessage("Please create a resolver section"))
    }
    if nService.Strategy == nil {
        panic(BadRequest.WithUserMessage("Please create a strategy section"))
    }
    if nService.Meta == nil {
        panic(BadRequest.WithUserMessage("Service metadata is not defined"))
    }
    
    nService.Meta.RoleID, nService.Meta.SubDomain = getRoleAndSubdomain(serviceId)
    
    nService.Resolver.Validate()
    nService.Strategy.Validate()
    
    err = nService.LoadLongDescriptionUrl()
    HttpCheckPanic(err, BadRequest.WithUserMessage("invalid LongDescriptionUrl or can't connecto to remote address"))
    
    vaultClient, err := vault.NewClient(vault.DefaultConfig())
    HttpCheckPanic(err, InternalError)
    vaultClient.SetToken(user)
    key := fmt.Sprintf("%s/%s", Config.VaultPath, serviceId)
    _, vaultErr := vaultClient.Logical().Write(key, structs.Map(nService))
    HttpCheckPanic(vaultErr, PermissionError)
    response.WriteEntity(nService)
}

//DeleteServiceHandler delete a  service
func (as *AdminServiceResource)deleteServiceHandler(request *restful.Request, response *restful.Response) {
    user := GetTokenFromContext(request)
    serviceId := request.PathParameter("id")
    ValidateService(serviceId)
    
    key := fmt.Sprintf("%s/%s", Config.VaultPath, serviceId)
    vaultClient, err := vault.NewClient(vault.DefaultConfig())
    HttpCheckPanic(err, InternalError)
    vaultClient.SetToken(user)
    HttpCheckPanic(err, InternalError.WithUserMessage("error decoding service"))
    _, err = vaultClient.Logical().Delete(key)
    HttpCheckPanic(err, PermissionError)
    response.WriteEntity(nil)
}

func (as *AdminServiceResource)getService(request *restful.Request, response *restful.Response) {
    user := GetTokenFromContext(request)
    serviceId := request.PathParameter("id")
    ValidateService(serviceId)
    key := fmt.Sprintf("%s/%s", Config.VaultPath, serviceId)
    vaultClient, err := vault.NewClient(vault.DefaultConfig())
    HttpCheckPanic(err, InternalError)
    vaultClient.SetToken(user)
    secret, err := vaultClient.Logical().Read(key)
    HttpCheckPanic(err, PermissionError)
    CheckSecretFailIfIsNull(secret)
    
    nService := &AdminServiceResource{}
    err = mapstructure.Decode(secret.Data, nService)
    HttpCheckPanic(err, InternalError.WithUserMessage("error decoding service"))
    response.WriteEntity(nService)
    
}

func (as *AdminServiceResource) listServiceHandler(request *restful.Request, response *restful.Response) {
    type ListResult struct {
        Keys []string
    }
    user := GetTokenFromContext(request)
    subdomain := request.QueryParameter("subdomain")
    ValidateSubdomain(subdomain)
    key := fmt.Sprintf("%s/roles", Config.VaultPath)
    vaultClient, err := vault.NewClient(vault.DefaultConfig())
    HttpCheckPanic(err, InternalError)
    vaultClient.SetToken(user)
    secret, err := vaultClient.Logical().List(key)
    HttpCheckPanic(err, PermissionError)
    CheckSecretFailIfIsNull(secret)
    
    sr := &ListResult{}
    err = mapstructure.Decode(secret.Data, sr)
    HttpCheckPanic(err, InternalError)
    
    roleList := sr.Keys
    
    ret := []*AdminServiceResource{}
    for _, role := range roleList {
        key := fmt.Sprintf("%s/roles/%s/%s", Config.VaultPath, role, subdomain)
        secret, err := vaultClient.Logical().Read(key)
        if !(err != nil || secret == nil || secret.Data == nil) {
            nService := &AdminServiceResource{}
            err = mapstructure.Decode(secret.Data, nService)
            HttpCheckPanic(err, InternalError.WithUserMessage("error decoding service"))
            ret = append(ret, nService)
        }
    }
    response.WriteEntity(ret)
}

