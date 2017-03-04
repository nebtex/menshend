package v1

import (
    "github.com/emicklei/go-restful"
    "github.com/mitchellh/mapstructure"
    vault "github.com/hashicorp/vault/api"
    "fmt"
    . "github.com/nebtex/menshend/pkg/apis/menshend"
    . "github.com/nebtex/menshend/pkg/config"
    . "github.com/nebtex/menshend/pkg/utils"
)

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

func (cs *ClientServiceResource) listServiceHandler(request *restful.Request, response *restful.Response) {
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
    if secret == nil || secret.Data == nil {
        panic(NotFound)
    }
    
    sr := &ListResult{}
    err = mapstructure.Decode(secret.Data, sr)
    HttpCheckPanic(err, InternalError)
    
    roleList := sr.Keys
    
    ret := []*AdminServiceResource{}
    for _, role := range roleList {
        key := fmt.Sprintf("%s/roles/%s/%s", Config.VaultPath, role, subdomain)
        secret, err := vaultClient.Logical().Read(key)
        if err != nil || secret == nil || secret.Data == nil {
            continue
        }
        nService := &AdminServiceResource{}
        err = mapstructure.Decode(secret.Data, nService)
        HttpCheckPanic(err, InternalError.Append("error decoding service"))
        ret = append(ret, nService)
    }
    response.WriteEntity(ret)
}

