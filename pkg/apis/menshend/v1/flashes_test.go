package v1

import (
    "testing"
    "net/http/httptest"
    "net/http"
    "github.com/nebtex/menshend/pkg/config"
    "io/ioutil"
    . "github.com/smartystreets/goconvey/convey"
    "encoding/json"
    mutils "github.com/nebtex/menshend/pkg/utils"

)

func addFlashHandlerHelper(w http.ResponseWriter, r *http.Request){
    // Get a session.
    session, err := config.FlashStore.Get(r, "flashes")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    // Set a new flash.
    session.AddFlash("1 flash")
    session.AddFlash("2 flash")
    session.AddFlash("3 flash")
    session.AddFlash("4 flash")
    session.AddFlash("5 flash")
    session.AddFlash("6 flash")
    mutils.CheckPanic(session.Save(r, w))
    
}

func Test_Flashes(t *testing.T) {
    Convey("shuld return available flashes and delete them when they are read", t, func(c C) {
        httpReq, err := http.NewRequest("GET", "/v1/flashes", nil)
        So(err, ShouldBeNil)
        httpReq.Header.Set("Content-Type", "application/json")
        So(err, ShouldBeNil)
        httpReq.Header.Add("X-Vault-Token", "myroot")
        httpWriter := httptest.NewRecorder()
        addFlashHandlerHelper(httpWriter, httpReq)
        httpReq.Header.Set("Cookie", "")
        httpReq.AddCookie(httpWriter.Result().Cookies()[0])
        httpWriter = httptest.NewRecorder()
        wsContainer := APIHandler()
        wsContainer.ServeHTTP(httpWriter, httpReq)
        jsres, err := ioutil.ReadAll(httpWriter.Body)
        So(err, ShouldBeNil)
        result := &FlashResource{}
        err = json.Unmarshal(jsres, result)
        So(err, ShouldBeNil)
        So(len(result.Flashes), ShouldEqual, 6)
        So(len(httpWriter.Result().Cookies()), ShouldEqual, 1)
        
    })
    
   
}
