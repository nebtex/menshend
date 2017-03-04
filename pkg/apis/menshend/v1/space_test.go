package v1

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "net/http/httptest"
    "github.com/emicklei/go-restful"
    "net/http"
)

func Test_SpaceEndpoint(t *testing.T) {
    Convey("sgould return info about the envirenment", t, func(c C) {
        
        httpReq, err := http.NewRequest("GET", "/v1/space", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        httpWriter := httptest.NewRecorder()
        wsContainer := restful.NewContainer()
        s := SpaceResource{}
        s.Register(wsContainer)
        wsContainer.ServeHTTP(httpWriter, httpReq)
        So(httpWriter.Body, ShouldNotBeNil)
        So(httpWriter.Code, ShouldEqual, 200)
        
    })
}
