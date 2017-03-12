package backend
/*
import (
    "github.com/ansel1/merry"
    "net/url"
    "fmt"
    "github.com/mitchellh/mapstructure"
    "strings"
    "github.com/yuin/gopher-lua"
    vault "github.com/hashicorp/vault/api"
)

const DefaultScript = `function getBackend (Username, Groups, AuthProvider)
    return ""
end`

type Role map[string]*Service

type menshend struct {
    Roles       map[string]Role
    Impersonate string
    Admin       string
}

type ServiceCache struct {
    // time to live seconds
    TTL    int
    Active bool
}



//CreateLuaScript create the lua script that will be execute to obtain the full
//backend url
func (s *Service)CreateLuaScript(u *User) string {
    getBackendScript := DefaultScript
    script := `math.randomseed(os.time())
math.random(); math.random(); math.random()
`
    script += fmt.Sprintf("\nusername = \"%s\"", u.Username)
    script += fmt.Sprintf("\ngroups = {\"%s\"}",
        strings.Join(u.Groups, "\", \""))
    script += fmt.Sprintf("\nauthProvider = \"%s\"", u.AuthProvider)
    //pick the resolver script if exists other case use the default one
    if len(s.LuaScript) > 0 {
        getBackendScript = s.LuaScript
    }
    return script + "\n\n" + getBackendScript
}

//GetBackend execute the custom or default lua script and calculate the backend
//full url
func (s *Service)GetBackend(u *User) (string, merry.Error) {
    var backend *string
    script := s.CreateLuaScript(u)
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
        backend = &ret
        return 0 // Notify that we pushed 0 value to the stack
    }
    
    l.SetGlobal("returnBackend", l.NewFunction(returnBackend))
    script += "\nreturnBackend(getBackend(username, groups, authProvider))"
    
    if err := l.DoString(script); err != nil {
        return "", LuaScriptFailed.Append(err.Error()).
            WithValue("user", u).WithValue("service", s).
            WithValue("script", script)
    }
    
    _, parseErr := url.ParseRequestURI(*backend)
    if parseErr != nil {
        return "", BadBackendUrl.Append(parseErr.Error()).
            WithValue("user", u).WithValue("service", s).
            WithValue("script", script).WithValue("backend", *backend)
    }
    return *backend, nil
}

type PlainBackend struct {
    Url                   *url.URL
    ImpersonateWithinRole bool
    Proxy                 bool
}

//GetBackend return the full backend url
func GetBackend(u *User, subDomain string, role string) (*PlainBackend, merry.Error) {
    if subDomain == "" {
        return nil, InvalidSubdomain
    }
    vc, vaultErr := vault.NewClient(VaultConfig)
    vc.SetToken(u.Token)
    CheckPanic(vaultErr)
    key := fmt.Sprintf("%s/roles/%s/%s", Config.VaultPath, role, subDomain)
    secret, vaultError := vc.Logical().Read(key)
    if vaultError != nil {
        if strings.Contains(vaultError.Error(), "403") {
            return nil, PermissionError.Append(vaultError.Error()).
                WithValue("user", u).WithValue("subdomain", subDomain)
        }
        CheckPanic(vaultError)
    }
    if (secret == nil) || (secret.Data == nil) {
        return nil, ServiceNotFound.
        WithValue("user", u).WithValue("subdomain", subDomain)
    }
    service := &Service{}
    decodeErr := mapstructure.Decode(secret.Data, service)
    if decodeErr != nil {
        return nil, InvalidService.
        WithValue("subdomain", subDomain).Append(decodeErr.Error())
    }
    if !service.IsActive {
        return nil, InactiveService.WithValue("subdomain", subDomain)
    }
    backendString, err := service.GetBackend(u)
    if err != nil {
        return nil, err
    }
    
    URL, errURL := url.ParseRequestURI(backendString)
    if errURL != nil {
        return nil, InvalidUrl.WithValue("url", backendString)
    }
    ret := &PlainBackend{}
    ret.Url = URL
    ret.ImpersonateWithinRole = service.ImpersonateWithinRole
    ret.Proxy = service.Proxy
    return ret, nil
    
    }
/*
//GetBackend return the full backend url
func GetBackend(u *User, subDomain string, role string) (*PlainBackend, merry.Error) {
    if subDomain == "" {
        return nil, InvalidSubdomain
    }
    vc, vaultErr := vault.NewClient(VaultConfig)
    vc.SetToken(u.Token)
    CheckPanic(vaultErr)
    key := fmt.Sprintf("%s/roles/%s/%s", Config.VaultPath, role, subDomain)
    secret, vaultError := vc.Logical().Read(key)
    if vaultError != nil {
        if strings.Contains(vaultError.Error(), "403") {
            return nil, PermissionError.Append(vaultError.Error()).
                WithValue("user", u).WithValue("subdomain", subDomain)
        }
        CheckPanic(vaultError)
    }
    if (secret == nil) || (secret.Data == nil) {
        return nil, ServiceNotFound.
        WithValue("user", u).WithValue("subdomain", subDomain)
    }
    service := &Service{}
    decodeErr := mapstructure.Decode(secret.Data, service)
    if decodeErr != nil {
        return nil, InvalidService.
        WithValue("subdomain", subDomain).Append(decodeErr.Error())
    }
    if !service.IsActive {
        return nil, InactiveService.WithValue("subdomain", subDomain)
    }
    backendString, err := service.GetBackend(u)
    if err != nil {
        return nil, err
    }
    
    URL, errURL := url.ParseRequestURI(backendString)
    if errURL != nil {
        return nil, InvalidUrl.WithValue("url", backendString)
    }
    ret := &PlainBackend{}
    ret.Url = URL
    ret.ImpersonateWithinRole = service.ImpersonateWithinRole
    ret.Proxy = service.Proxy
    return ret, nil
}*/
    
