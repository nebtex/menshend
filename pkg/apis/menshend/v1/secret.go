package v1

import (
    "github.com/emicklei/go-restful"
    . "github.com/nebtex/menshend/pkg/config"
    vault "github.com/hashicorp/vault/api"
    . "github.com/nebtex/menshend/pkg/utils"
    . "github.com/nebtex/menshend/pkg/apis/menshend"
)

//Space ..
type SecretResource struct {
}

func (s *SecretResource) Register(container *restful.Container) {
    ws := new(restful.WebService).
        Consumes(restful.MIME_JSON).
        Produces(restful.MIME_JSON).Filter(LoginFilter)
    
    ws.Path("/v1/secret").
        Doc("return secret associate with a service")
    
    ws.Route(ws.GET("/{id:*}").To(s.read).
        Operation("readSecret").
        Param(ws.PathParameter("id", "secret path").DataType("string")).
        Writes(vault.Secret{}))
    
    container.Add(ws)
}

func (s *SecretResource) read(request *restful.Request, response *restful.Response) {
    secretID := request.PathParameter("id")
    user := GetUserFromContext(request)
    key := ValidateSecret(secretID, user)
    vaultClient, err := vault.NewClient(VaultConfig)
    HttpCheckPanic(err, InternalError)
    vaultClient.SetToken(user.Menshend.VaultToken)
    secret, err := vaultClient.Logical().Read(key)
    HttpCheckPanic(err, PermissionError)
    if secret == nil || secret.Data == nil {
        panic(NotFound)
    }
    response.WriteEntity(secret)
}

