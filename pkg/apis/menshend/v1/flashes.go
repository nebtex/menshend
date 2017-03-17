package v1

import (
    "github.com/emicklei/go-restful"
    . "github.com/nebtex/menshend/pkg/config"
    . "github.com/nebtex/menshend/pkg/utils"
)

type FlashResource struct {
    Flashes []string `json:"flashes"`
}

func (f *FlashResource) Register(container *restful.Container) {
    ws := new(restful.WebService).
        Consumes(restful.MIME_JSON).
        Produces(restful.MIME_JSON)
    ws.Path("/v1/flashes").
        Doc("get the flases, this only works on browsers")
    ws.Route(ws.GET("").To(f.get).
        Doc("list current flashes and delete them").
        Operation("getFlashes").
        Writes(FlashResource{}))
    container.Add(ws)
}

func (f *FlashResource) get(request *restful.Request, response *restful.Response) {
    session, err := FlashStore.Get(request.Request, "flashes")
    HttpCheckPanic(err, InternalError)
    flashes := []string{}
    for _, v := range session.Flashes() {
        flash := v.(string)
        flashes = append(flashes, flash)
    }
    fn := &FlashResource{}
    fn.Flashes = flashes
    response.WriteEntity(fn)
}
