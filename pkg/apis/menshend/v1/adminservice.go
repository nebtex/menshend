package v1

import (
    "net/url"
    "io/ioutil"
    "net/http"
    "github.com/emicklei/go-restful"
    "github.com/mitchellh/mapstructure"
    vault "github.com/hashicorp/vault/api"
    "fmt"
    . "github.com/nebtex/menshend/pkg/apis/menshend"
    . "github.com/nebtex/menshend/pkg/config"
    . "github.com/nebtex/menshend/pkg/utils"
    "github.com/fatih/structs"
    "strings"
)

func init() {
    restful.PrettyPrintResponses = false
}

type ServiceCache struct {
    // time to live seconds
    TTL    int `json:"ttl"`
    Active bool `json:"active"`
}

type StrategyTypes int

var AllStrategyTypes = []string{
    "proxy",
    "redirect",
    "port-forward",
}

type LanguageTypes int

var AllLanguageTypes = []string{
    "lua",
    "yaml",
}
// Options is a configuration container to setup the CORS middleware.
type CorsOptions struct {
    // AllowedOrigins is a list of origins a cross-domain request can be executed from.
    // If the special "*" value is present in the list, all origins will be allowed.
    // An origin may contain a wildcard (*) to replace 0 or more characters
    // (i.e.: http://*.domain.com). Usage of wildcards implies a small performance penality.
    // Only one wildcard can be used per origin.
    // Default value is ["*"]
    AllowedOrigins     []string
    // AllowedMethods is a list of methods the client is allowed to use with
    // cross-domain requests. Default value is simple methods (GET and POST)
    AllowedMethods     []string
    // AllowedHeaders is list of non simple headers the client is allowed to use with
    // cross-domain requests.
    // If the special "*" value is present in the list, all headers will be allowed.
    // Default value is [] but "Origin" is always appended to the list.
    AllowedHeaders     []string
    // ExposedHeaders indicates which headers are safe to expose to the API of a CORS
    // API specification
    ExposedHeaders     []string
    // AllowCredentials indicates whether the request can include user credentials like
    // cookies, HTTP authentication or client side SSL certificates.
    AllowCredentials   bool
    // MaxAge indicates how long (in seconds) the results of a preflight request
    // can be cached
    MaxAge             int
    // OptionsPassthrough instructs preflight to let other potential next handlers to
    // process the OPTIONS method. Turn this on if your application handles OPTIONS.
    OptionsPassthrough bool
    // Debugging flag adds additional output to debug server side CORS issues
    Debug              bool
}

//Service service definition struct
type AdminServiceResource struct {
    ID                    string `json:"id"`
    RoleID                string `json:"roleId"`
    SubDomain             string `json:"subDomain"`
    Logo                  string `json:"logo"`
    Name                  string `json:"name"`
    ShortDescription      string `json:"shortDescription"`
    LongDescription       string `json:"longDescription"`
    LongDescriptionUrl    string `json:"longDescriptionUrl"`
    ProxyCode             string `json:"proxyCode"`
    ProxyLanguage         string `json:"proxyCodeLanguage"`
    Cache                 ServiceCache `json:"cache"`
    ImpersonateWithinRole bool   `json:"impersonateWithinRole"`
    Strategy              string `json:"strategy"`
    IsActive              bool `json:"isActive"`
    SecretPaths           []string `json:"secretPaths"`
    Cors                  CorsOptions `json:"cors"`
    EnableCustomCors      bool `json:"enableCustomCors"`
    CSRF                  bool `json:"csrf"`
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
    user := GetUserFromContext(request)
    serviceId := request.PathParameter("id")
    ValidateService(serviceId)
    nService := &AdminServiceResource{}
    err := request.ReadEntity(nService)
    HttpCheckPanic(err, BadRequest.Append("error decoding request"))
    
    nService.RoleID, nService.SubDomain = getRoleAndSubdomain(serviceId)
    ValidateLanguageTypes(nService.ProxyLanguage)
    ValidateStrategyTypes(nService.Strategy)
    
    err = nService.LoadLongDescriptionUrl()
    HttpCheckPanic(err, BadRequest.Append("invalid LongDescriptionUrl or can't connecto to remote address"))
    
    vaultClient, err := vault.NewClient(VaultConfig)
    HttpCheckPanic(err, InternalError)
    vaultClient.SetToken(user.Menshend.VaultToken)
    key := fmt.Sprintf("%s/%s", Config.VaultPath, serviceId)
    _, vaultErr := vaultClient.Logical().Write(key, structs.Map(nService))
    HttpCheckPanic(vaultErr, PermissionError)
    response.WriteEntity(nService)
}

//DeleteServiceHandler delete a  service
func (as *AdminServiceResource)deleteServiceHandler(request *restful.Request, response *restful.Response) {
    user := GetUserFromContext(request)
    serviceId := request.PathParameter("id")
    ValidateService(serviceId)
    
    key := fmt.Sprintf("%s/%s", Config.VaultPath, serviceId)
    vaultClient, err := vault.NewClient(VaultConfig)
    HttpCheckPanic(err, InternalError)
    vaultClient.SetToken(user.Menshend.VaultToken)
    HttpCheckPanic(err, InternalError.Append("error decoding service"))
    _, err = vaultClient.Logical().Delete(key)
    HttpCheckPanic(err, PermissionError)
    response.WriteEntity(nil)
}

func (as *AdminServiceResource)getService(request *restful.Request, response *restful.Response) {
    user := GetUserFromContext(request)
    serviceId := request.PathParameter("id")
    ValidateService(serviceId)
    key := fmt.Sprintf("%s/%s", Config.VaultPath, serviceId)
    vaultClient, err := vault.NewClient(VaultConfig)
    HttpCheckPanic(err, InternalError)
    vaultClient.SetToken(user.Menshend.VaultToken)
    secret, err := vaultClient.Logical().Read(key)
    HttpCheckPanic(err, PermissionError)
    CheckSecretFailIfIsNull(secret)
    
    nService := &AdminServiceResource{}
    err = mapstructure.Decode(secret.Data, nService)
    HttpCheckPanic(err, InternalError.Append("error decoding service"))
    response.WriteEntity(nService)
    
}

func (as *AdminServiceResource) listServiceHandler(request *restful.Request, response *restful.Response) {
    type ListResult struct {
        Keys []string
    }
    user := GetUserFromContext(request)
    subdomain := request.QueryParameter("subdomain")
    ValidateSubdomain(subdomain)
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
    
    ret := []*AdminServiceResource{}
    for _, role := range roleList {
        key := fmt.Sprintf("%s/roles/%s/%s", Config.VaultPath, role, subdomain)
        secret, err := vaultClient.Logical().Read(key)
        if !(err != nil || secret == nil || secret.Data == nil) {
            nService := &AdminServiceResource{}
            err = mapstructure.Decode(secret.Data, nService)
            HttpCheckPanic(err, InternalError.Append("error decoding service"))
            ret = append(ret, nService)
        }
    }
    response.WriteEntity(ret)
}

