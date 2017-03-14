package resolvers

import (
    "strings"
    "fmt"
    "github.com/yuin/gopher-lua"
    "net/url"
    . "github.com/nebtex/menshend/pkg/apis/menshend/v1"
    . "github.com/nebtex/menshend/pkg/users"
    . "github.com/nebtex/menshend/pkg/apis/menshend"
    . "github.com/nebtex/menshend/pkg/utils"
)

type LuaResolver struct {
}

func (lr *LuaResolver)Resolve(c *AdminServiceResource, u *User) (*backendImplementation) {
    ys := GetBackend(c, u)
    return &backendImplementation{ys:ys}
}



//CreateLuaScript create the lua script that will be execute to obtain the full
//backend url
func CreateLuaScript(s *AdminServiceResource, u *User) string {
    script := `math.randomseed(os.time())
math.random(); math.random(); math.random()
`
    script += fmt.Sprintf("\nusername = \"%s\"", u.Menshend.Username)
    script += fmt.Sprintf("\ngroups = {\"%s\"}", strings.Join(u.Menshend.Groups, "\", \""))
    script += fmt.Sprintf("\nauthProvider = \"%s\"", u.Menshend.AuthProvider)
    
    return script + "\n\n" + s.ProxyCode
}

//GetBackend execute the custom or default lua script and calculate the backend
//full url
//TODO: add vault token metadata
func GetBackend(s *AdminServiceResource, u *User) (*backendValues) {
    var baseUrl *string
    var headerMaps map[string]string
    
    script := CreateLuaScript(s, u)
    l := lua.NewState()
    lua.OpenBase(l)
    lua.OpenString(l)
    lua.OpenTable(l)
    lua.OpenMath(l)
    lua.OpenOs(l)
    lua.OpenIo(l)
    lua.OpenPackage(l)
    
    returnBackend := func(L *lua.LState) int {
        ret := L.ToString(1)
        table := L.ToTable(2)
        headerMaps = map[string]string{}
        table.ForEach(func(k lua.LValue, v lua.LValue) {
            headerMaps[k.String()] = v.String()
        })
        baseUrl = &ret
        return 0 // Notify that we pushed 0 value to the stack
    }
    
    l.SetGlobal("returnBackend", l.NewFunction(returnBackend))
    script += "\nreturnBackend(getBackend(username, groups, authProvider))"
    
    err := l.DoString(script)
    HttpCheckPanic(err, InternalError.WithValue("user", u).WithValue("service", s).
        WithValue("script", script))
    
    _, parseErr := url.ParseRequestURI(*baseUrl)
    HttpCheckPanic(parseErr, InternalError.WithValue("user", u).WithValue("service", s).
        WithValue("script", script).WithValue("backend", *baseUrl))
    return &backendValues{BaseUrl:*baseUrl, HeadersMap:headerMaps}
}
    
    
    
    
