package v1

import (
    "github.com/emicklei/go-restful"
    vault "github.com/hashicorp/vault/api"
    mutils "github.com/nebtex/menshend/pkg/utils"
    mfilters "github.com/nebtex/menshend/pkg/filters"
)

//SecretResource ...
type SecretResource struct {
}
//Register ...
func (s *SecretResource) Register(container *restful.Container) {
    ws := new(restful.WebService).
        Consumes(restful.MIME_JSON).
        Produces(restful.MIME_JSON).Filter(mfilters.LoginFilter)
    
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
    user := mfilters.GetTokenFromContext(request)
    key := ValidateSecret(secretID, user)
    vaultClient, err := vault.NewClient(vault.DefaultConfig())
    mutils.HttpCheckPanic(err, mutils.InternalError)
    vaultClient.SetToken(user)
    secret, err := vaultClient.Logical().Read(key)
    mutils.HttpCheckPanic(err, mutils.PermissionError)
    CheckSecretFailIfIsNull(secret)
    mutils.HttpCheckPanic(response.WriteEntity(secret), mutils.InternalError)
}

