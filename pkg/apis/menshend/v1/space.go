package v1

import (
    "github.com/emicklei/go-restful"
    "github.com/nebtex/menshend/pkg/config"
    mutils "github.com/nebtex/menshend/pkg/utils"
)

//SpaceResource ...
type SpaceResource struct {
    Logo             string `json:"logo"`
    Name             string `json:"name"`
    ShortDescription string `json:"shortDescription"`
    LongDescription  string `json:"longDescription"`
    Host             string `json:"host"`
}

//Register ...
func (s *SpaceResource) Register(container *restful.Container) {
    ws := new(restful.WebService).
        Consumes(restful.MIME_JSON).
        Produces(restful.MIME_JSON)
    
    ws.Path("/v1/space").
        Doc("Get info about the current environment")
    
    ws.Route(ws.GET("").To(s.info).
        Operation("describeSpace").
        Writes(SpaceResource{}))
    container.Add(ws)
    
}

func (s *SpaceResource) info(request *restful.Request, response *restful.Response) {
    ns := SpaceResource{}
    ns.LongDescription = config.Config.GetLongDescription()
    ns.ShortDescription = config.Config.GetShortDescription()
    ns.Name = config.Config.GetName()
    ns.Logo = config.Config.GetLogo()
    ns.Host = config.Config.Host()
    mutils.HttpCheckPanic(response.WriteEntity(ns), mutils.InternalError)
}

