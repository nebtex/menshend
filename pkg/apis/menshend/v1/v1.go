package v1

import (
    "net/http"
    "github.com/emicklei/go-restful"
)

func ApiHandler() http.Handler {
    wsContainer := restful.NewContainer()
    account:= &AuthResource{}
    account.Register(wsContainer)
    admin := &AdminServiceResource{}
    admin.Register(wsContainer)
    client := &ClientServiceResource{}
    client.Register(wsContainer)
    secret := SecretResource{}
    secret.Register(wsContainer)
    space := SpaceResource{}
    space.Register(wsContainer)
    return wsContainer
}
