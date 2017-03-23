package v1

import (
    "net/url"
    "io/ioutil"
    "net/http"
    "github.com/emicklei/go-restful"
    "github.com/mitchellh/mapstructure"
    "fmt"
    "github.com/fatih/structs"
    "strings"
    vault "github.com/hashicorp/vault/api"
    "github.com/nebtex/menshend/pkg/strategy"
    "github.com/nebtex/menshend/pkg/resolvers"
    mutils "github.com/nebtex/menshend/pkg/utils"
    mfilters "github.com/nebtex/menshend/pkg/filters"
    mconfig "github.com/nebtex/menshend/pkg/config"
    "github.com/Sirupsen/logrus"
)

func init() {
    restful.PrettyPrintResponses = false
}

//ServiceCache activate a cache for the resolvers result
type ServiceCache struct {
    // time to live seconds
    TTL    int `json:"ttl"`
}

//ServiceStrategy defines how menshend will handle the user request
type ServiceStrategy struct {
    Proxy       *strategy.Proxy       `json:"proxy"`
    PortForward *strategy.PortForward `json:"portForward"`
    Redirect    *strategy.Redirect    `json:"redirect"`
}

//Validate a service can only contains a strategy
func (ss *ServiceStrategy) Validate() {
    defined := 0
    if ss.Proxy != nil {
        defined ++
    }
    if ss.PortForward != nil {
        defined  ++
    }
    if ss.Redirect != nil {
        defined ++
    }
    
    if defined != 1 {
        panic(mutils.BadRequest.WithUserMessage("Please define only one strategy, you have not define one!! or have multiples defined"))
    }
}

//Get returns the active strategy
func (ss *ServiceStrategy)Get() strategy.Strategy {
    if ss.Proxy != nil {
        return ss.Proxy
    }
    if ss.PortForward != nil {
        return ss.PortForward
    }
    return ss.Redirect
}

//ServiceResolver ..
type ServiceResolver struct {
    Yaml *resolvers.YAMLResolver `json:"yaml"`
    Lua  *resolvers.LuaResolver `json:"lua"`
}

//Validate a service can only contains an resolver (lua, yaml, js, etc..)
func (sr *ServiceResolver) Validate() {
    defined := 0
    if sr.Yaml != nil {
        defined  ++
    }
    if sr.Lua != nil {
        defined  ++
    }
    if defined != 1 {
        panic(mutils.BadRequest.WithUserMessage("Please define only one resolver, you have not define one!! or have multiples defined"))
    }
}

//Get returns the active resolver
func (sr *ServiceResolver) Get() resolvers.Resolver {
    if sr.Lua != nil {
        return sr.Lua
    }
    return sr.Yaml
}

//SimpleLongDescription the user defines the contents manually - supports markdown
type SimpleLongDescription struct {
    Content string `json:"content"`
}

//LongDescription ...
func (sld *SimpleLongDescription)LongDescription() string {
    return sld.Content
}

//Load ...
func (sld *SimpleLongDescription) Load() {}

//URLLongDescription allow to the user to load the service long description from a remote url
//a README.md file is preferred
type URLLongDescription struct {
    URL     string `json:"url"`
    Content string `json:"content"`
}

//LongDescription ...
func (uld *URLLongDescription)LongDescription() string {
    return uld.Content
}

//Load query the remote endpoint and load the content from there
func (uld *URLLongDescription) Load() {
    if uld.URL != "" {
        _, err := url.ParseRequestURI(uld.URL)
        mutils.HttpCheckPanic(err, mutils.BadRequest.WithUserMessage("Long description remote url is not valid"))
        r, err := getReadme(uld.URL)
        mutils.HttpCheckPanic(err, mutils.BadRequest.WithUserMessage("Long description remote url not responding"))
        uld.Content = string(r)
    }
}

//ServiceLongDescription long description options
type ServiceLongDescription struct {
    Remote *URLLongDescription `json:"remote"`
    Local  *SimpleLongDescription `json:"local"`
}


//Validate ...
func (sldn *ServiceLongDescription)Validate() {
    defined := 0
    if sldn.Local != nil {
        defined ++
    }
    if sldn.Remote != nil {
        defined ++
    }
    if defined != 1 {
        panic(mutils.BadRequest.WithUserMessage("Please define only a way to pull the service long description"))
    }
}

//Load load long description from remote/local or whatever resource
func (sldn *ServiceLongDescription)Load() {
    if sldn.Local != nil {
        sldn.Local.Load()
    } else if sldn.Remote != nil {
        sldn.Remote.Load()
    }
}

//LongDescription ...
func (sldn *ServiceLongDescription)LongDescription() string {
    if sldn.Local != nil {
        return sldn.Local.LongDescription()
    } else if sldn.Remote != nil {
        return sldn.Remote.LongDescription()
    }
    return ""
}


//ServiceMetadata ...
type ServiceMetadata struct {
    ID              string `json:"id"`
    RoleID          string `json:"roleId"`
    SubDomain       string `json:"subDomain"`
    Name            string `json:"name"`
    Logo            string `json:"logo"`
    Description     string `json:"description"`
    Tags            []string `json:"tags"`
    LongDescription *ServiceLongDescription `json:"longDescription"`
}

//AdminServiceResource service definition struct
type AdminServiceResource struct {
    ImpersonateWithinRole bool             `json:"impersonateWithinRole"`
    IsActive              *bool            `json:"isActive"`
    SecretPaths           []string         `json:"secretPaths"`
    Meta                  *ServiceMetadata `json:"meta"`
    Resolver              *ServiceResolver `json:"Resolver"`
    Strategy              *ServiceStrategy `json:"strategy"`
    Cache                 *ServiceCache    `json:"cache"`
}

func getReadme(url string) ([]byte, error) {
    r, err := http.Get(url)
    if err != nil {
        return []byte{}, err
    }
    defer func() {
        err := r.Body.Close()
        if err != nil {
            logrus.Error(":O")
            logrus.Error(err)
        }
    }()
    return ioutil.ReadAll(r.Body)
}

//Register ...
func (as *AdminServiceResource) Register(container *restful.Container) {
    ws := new(restful.WebService).
        Consumes(restful.MIME_JSON).
        Produces(restful.MIME_JSON)
    
    ws.Path("/v1/adminServices").
        Filter(mfilters.LoginFilter).
        Filter(mfilters.AdminFilter).
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
    user := mfilters.GetTokenFromContext(request)
    serviceID := request.PathParameter("id")
    ValidateService(serviceID)
    nService := &AdminServiceResource{}
    err := request.ReadEntity(nService)
    mutils.HttpCheckPanic(err, mutils.BadRequest.WithUserMessage("error decoding request"))
    
    if nService.Resolver == nil {
        panic(mutils.BadRequest.WithUserMessage("Please create a resolver section"))
    }
    if nService.Strategy == nil {
        panic(mutils.BadRequest.WithUserMessage("Please create a strategy section"))
    }
    if nService.Meta == nil {
        panic(mutils.BadRequest.WithUserMessage("Service metadata is not defined"))
    }
    
    nService.Meta.RoleID, nService.Meta.SubDomain = getRoleAndSubdomain(serviceID)
    
    nService.Resolver.Validate()
    nService.Strategy.Validate()
    
    if nService.Meta.LongDescription != nil {
        nService.Meta.LongDescription.Validate()
        nService.Meta.LongDescription.Load()
    }
    
    vaultClient, err := vault.NewClient(vault.DefaultConfig())
    mutils.HttpCheckPanic(err, mutils.InternalError)
    vaultClient.SetToken(user)
    key := fmt.Sprintf("%s/%s", mconfig.Config.VaultPath, serviceID)
    _, vaultErr := vaultClient.Logical().Write(key, structs.Map(nService))
    mutils.HttpCheckPanic(vaultErr, mutils.PermissionError)
    mutils.HttpCheckPanic(response.WriteEntity(nService), mutils.InternalError)
}

//DeleteServiceHandler delete a  service
func (as *AdminServiceResource)deleteServiceHandler(request *restful.Request, response *restful.Response) {
    user := mfilters.GetTokenFromContext(request)
    serviceID := request.PathParameter("id")
    ValidateService(serviceID)
    
    key := fmt.Sprintf("%s/%s", mconfig.Config.VaultPath, serviceID)
    vaultClient, err := vault.NewClient(vault.DefaultConfig())
    mutils.HttpCheckPanic(err, mutils.InternalError)
    vaultClient.SetToken(user)
    _, err = vaultClient.Logical().Delete(key)
    mutils.HttpCheckPanic(err, mutils.PermissionError)
    
    mutils.HttpCheckPanic(response.WriteEntity(nil), mutils.InternalError)
}

func (as *AdminServiceResource)getService(request *restful.Request, response *restful.Response) {
    user := mfilters.GetTokenFromContext(request)
    serviceID := request.PathParameter("id")
    ValidateService(serviceID)
    key := fmt.Sprintf("%s/%s", mconfig.Config.VaultPath, serviceID)
    vaultClient, err := vault.NewClient(vault.DefaultConfig())
    mutils.HttpCheckPanic(err, mutils.InternalError)
    vaultClient.SetToken(user)
    secret, err := vaultClient.Logical().Read(key)
    mutils.HttpCheckPanic(err, mutils.PermissionError)
    CheckSecretFailIfIsNull(secret)
    
    nService := &AdminServiceResource{}
    err = mapstructure.Decode(secret.Data, nService)
    mutils.HttpCheckPanic(err, mutils.InternalError.WithUserMessage("error decoding service"))
    mutils.HttpCheckPanic(response.WriteEntity(nService), mutils.InternalError)
    
}

func (as *AdminServiceResource) listServiceHandler(request *restful.Request, response *restful.Response) {
    user := mfilters.GetTokenFromContext(request)
    subdomain := request.QueryParameter("subdomain")
    ValidateSubdomain(subdomain)
    key := fmt.Sprintf("%s/roles", mconfig.Config.VaultPath)
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
    
    ret := []*AdminServiceResource{}
    for _, role := range roleList {
        key := fmt.Sprintf("%s/roles/%s/%s", mconfig.Config.VaultPath, role, subdomain)
        secret, err := vaultClient.Logical().Read(key)
        if !(err != nil || secret == nil || secret.Data == nil) {
            nService := &AdminServiceResource{}
            err = mapstructure.Decode(secret.Data, nService)
            mutils.HttpCheckPanic(err, mutils.InternalError.WithUserMessage("error decoding service"))
            ret = append(ret, nService)
        }
    }
    mutils.HttpCheckPanic(response.WriteEntity(ret), mutils.InternalError)
}

