package resolvers

import (
    "github.com/yuin/gopher-lua"
    "net/url"
    "net/http"
    . "github.com/nebtex/menshend/pkg/utils"
    vault "github.com/hashicorp/vault/api"
    "github.com/cjoudrey/gluahttp"
    luajson "layeh.com/gopher-json"
    "github.com/yuin/gluamapper"
    "fmt"
    st "layeh.com/gopher-luar"
)

type LuaResolver struct {
    Content string `json:"content"`
    Body    string `json:"-"`
}

func (lr *LuaResolver) SetBody(s string) {
    lr.Body = s
}

func (lr *LuaResolver)Resolve(v *vault.Secret) (Backend) {
    
    script := lr.Content
    script += "\nreturnBackend(getBackend(TokenInfo, Body))"
    
    l := lua.NewState()
   
    lua.OpenBase(l)
    lua.OpenString(l)
    lua.OpenTable(l)
    lua.OpenMath(l)
    lua.OpenOs(l)
    lua.OpenIo(l)
    lua.OpenPackage(l)
    
    l.PreloadModule("http", gluahttp.NewHttpModule(&http.Client{}).Loader)
    luajson.Preload(l)
    l.SetGlobal("TokenInfo", st.New(l, v))
    l.SetGlobal("Body", st.New(l, lr.Body))
    ret := &SimpleBackend{}
    ret.ys = &backendValues{}
    
    returnBackend := func(L *lua.LState) int {
        table := L.ToTable(1)
        err := gluamapper.Map(table, ret.ys)
        HttpCheckPanic(err, InternalError)
        return 0 // Notify that we pushed 0 value to the stack
    }
    
    l.SetGlobal("returnBackend", l.NewFunction(returnBackend))
    err := l.DoString(script)
    HttpCheckPanic(err, InternalError.WithValue("script", script))
    
    _, parseErr := url.ParseRequestURI(ret.ys.BaseUrl)
    HttpCheckPanic(parseErr, InternalError.WithValue("script", lr.Content).WithValue("backend", fmt.Sprintf("%v", ret)))
    return ret
}



    
