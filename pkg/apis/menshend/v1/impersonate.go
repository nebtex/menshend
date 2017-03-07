package v1

import (
    "github.com/emicklei/go-restful"
    . "github.com/nebtex/menshend/pkg/apis/menshend"
    . "github.com/nebtex/menshend/pkg/users"
    . "github.com/nebtex/menshend/pkg/utils"
)

func AuthProviderPtr(v AuthProviderType) *AuthProviderType {
    return &v
}
//Space ..
type ImpersonateResource struct {
    Active       bool  `json:"active"`
    User         *string `json:"user,omitempty"`
    Groups       []string `json:"groups,omitempty"`
    AuthProvider *AuthProviderType `json:"authProviderType,omitempty"`
}

func (i *ImpersonateResource) Register(container *restful.Container) {
    ws := new(restful.WebService).
        Consumes(restful.MIME_JSON).
        Produces(restful.MIME_JSON).
        Filter(LoginFilter).
        Filter(ImpersonateFilter)
    
    ws.Path("/v1/impersonate").
        Doc("impersonate another user")
    
    ws.Route(ws.GET("").To(i.read).
        Operation("impersonateStatus").
        Doc("Check if there is an impersonification going on").
        Writes(ImpersonateResource{}))
    
    ws.Route(ws.PUT("").To(i.impersonate).
        Reads(ImpersonateResource{}).
        Operation("impersonate").
        Writes(ImpersonateResource{}))
    
    ws.Route(ws.DELETE("").To(i.stopImpersonate).
        Operation("stopImpersonate"))
    container.Add(ws)
}

func (i *ImpersonateResource) read(request *restful.Request, response *restful.Response) {
    user := GetUserFromContext(request)
    if user.Menshend.ImpersonateBy == nil {
        response.WriteEntity(ImpersonateResource{Active:false})
        return
    }
    response.WriteEntity(ImpersonateResource{
        Active:true,
        User:StringPtr(user.Menshend.Username),
        Groups:user.Menshend.Groups,
        AuthProvider:AuthProviderPtr(user.Menshend.AuthProvider),
    })
}

func (i *ImpersonateResource) impersonate(request *restful.Request, response *restful.Response) {
    user := GetUserFromContext(request)
    ipr := &ImpersonateResource{}
    err := request.ReadEntity(ipr)
    HttpCheckPanic(err, BadRequest.Append("invalid request format"))
    
    err = user.Impersonate(ipr.AuthProvider, ipr.User, ipr.Groups...)
    HttpCheckPanic(err, BadRequest)
    
    response.AddHeader("X-Menshend-Token", user.GenerateJWT())
    response.WriteEntity(ImpersonateResource{
        Active:true,
        User:StringPtr(user.Menshend.Username),
        Groups:user.Menshend.Groups,
        AuthProvider:AuthProviderPtr(user.Menshend.AuthProvider),
    })
}

func (i *ImpersonateResource) stopImpersonate(request *restful.Request, response *restful.Response) {
    user := GetUserFromContext(request)
    user.StopImpersonate()
    response.AddHeader("X-Menshend-Token", user.GenerateJWT())
    response.WriteEntity(nil)
    
}
