package v1

import (
    "github.com/emicklei/go-restful"
    "github.com/nebtex/menshend/pkg/config"
    mutils "github.com/nebtex/menshend/pkg/utils"
)

//FlashResource  this allow to store some error messages in the cookie
//only useful in the browser
type FlashResource struct {
    Flashes []string `json:"flashes"`
}

//Register ...
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
    session, err := config.FlashStore.Get(request.Request, "flashes")
    mutils.HttpCheckPanic(err, mutils.InternalError)
    flashes := []string{}
    for _, v := range session.Flashes() {
        flash := v.(string)
        flashes = append(flashes, flash)
    }
    fn := &FlashResource{}
    fn.Flashes = flashes
    mutils.HttpCheckPanic(response.WriteEntity(fn), mutils.InternalError)
}
