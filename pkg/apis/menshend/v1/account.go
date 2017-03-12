package v1

import (
    vault "github.com/hashicorp/vault/api"
    "fmt"
    "github.com/emicklei/go-restful"
    . "github.com/nebtex/menshend/pkg/apis/menshend"
    . "github.com/nebtex/menshend/pkg/config"
    . "github.com/nebtex/menshend/pkg/utils"
    . "github.com/nebtex/menshend/pkg/users"
    "time"
    "github.com/mitchellh/mapstructure"
)

type AuthResource struct {
    Data         map[string]interface{} `json:"data"`
    AuthProvider string `json:"authProvider"`
}

type LoginStatus struct {
    IsLogged         bool `json:"isLogged"`
    IsAdmin          bool `json:"isAdmin"`
    CanImpersonate   bool `json:"canImpersonate"`
    SessionExpiresAt int64 `json:"sessionExpiresAt"`
}

type UPLogin struct {
    User     string `json:"user"`
    Password string `json:"password"`
    Type     string `json:"type"`
}

type TokenLogin struct {
    Token string `json:"token"`
}

type GithubLogin struct {
    Token string `json:"token"`
}

func (a *AuthResource) Register(container *restful.Container) {
    ws := new(restful.WebService).
        Consumes(restful.MIME_JSON).
        Produces(restful.MIME_JSON)
    
    ws.Path("/v1/account").
        Doc("login service")
    
    ws.Route(ws.GET("").To(a.accountStatus).
        Doc("get login status").
        Operation("loginStatus").
        Writes(LoginStatus{}))
    
    ws.Route(ws.PUT("").To(a.accountLogin).
        Doc("get login status").
        Operation("login").
        Writes([]LoginStatus{}))
    
    ws.Route(ws.DELETE("").To(a.logout).
        Doc("get login status").
        Operation("logout"))
    container.Add(ws)
    
}
func (a *AuthResource) logout(request *restful.Request, response *restful.Response) {
    defer func() {}()
    user := GetUserFromRequest(request)
    vc, err := vault.NewClient(VaultConfig)
    CheckPanic(err)
    vc.SetToken(user.Menshend.VaultToken)
    err = vc.Auth().Token().RevokeSelf(user.Menshend.VaultToken)
    HttpCheckPanic(err, PermissionError)
}

func MakeTimestampMillisecond() int64 {
    return time.Now().UnixNano() / int64(time.Millisecond)
}
/*
func setToken(u *User, expiresIn int64, response *restful.Response, hasCSRF bool) {
    expireAt := MakeTimestampMillisecond()
    if expiresIn == 0 {
        expireAt += Config.DefaultTTL
    } else {
        expireAt += expiresIn
    }
    u.SetExpiresAt(expireAt)
    
    if !hasCSRF {
        response.AddHeader("X-Menshend-Token", u.GenerateJWT())
        
    } else {
        ct := &http.Cookie{Path: "/", Name: "X-Menshend-Token", Value: u.GenerateJWT(),
            Expires: time.Unix(u.ExpiresAt / 1000, 0),
            HttpOnly:true }
        
        ct.Domain = "." + Config.HostWithoutPort()
        
        if Config.Scheme == "https" {
            ct.Secure = true
        }
        http.SetCookie(response.ResponseWriter, ct)
        
    }
}
*/
func getToken(u *User, expiresIn int64) string {
    expireAt := MakeTimestampMillisecond()
    if expiresIn == 0 {
        expireAt += Config.DefaultTTL
    } else {
        expireAt += expiresIn
    }
    u.SetExpiresAt(expireAt)
    return u.GenerateJWT()
    
}

func (*AuthResource)accountStatus(request *restful.Request, response *restful.Response) {
    defer func() {
        r := recover()
        if r != nil {
            ls := LoginStatus{
                IsLogged: false,
                IsAdmin: false,
                CanImpersonate: false,
                SessionExpiresAt: 0,
            }
            response.WriteEntity(ls)
        }
        
    }()
    
    user := GetUserFromRequest(request)
    ls := LoginStatus{
        IsLogged: true,
        IsAdmin: IsAdmin(user),
        CanImpersonate: CanImpersonate(user),
        SessionExpiresAt: user.ExpiresAt,
    }
    response.WriteEntity(ls)
}

func VaultLogin(c *vault.Client, path string, data map[string]interface{}) (*vault.Secret, error) {
    r := c.NewRequest("POST", "/v1/" + path)
    if err := r.SetJSONBody(data); err != nil {
        return nil, err
    }
    resp, err := c.RawRequest(r)
    if resp != nil {
        defer resp.Body.Close()
    }
    if resp != nil && resp.StatusCode == 404 {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    
    return vault.ParseSecret(resp.Body)
}

// user/password
func UserPasswordHandler(upr *UPLogin) string {
    var key string
    vc, err := vault.NewClient(VaultConfig)
    HttpCheckPanic(err, InternalError)
    data := map[string]interface{}{"password": upr.Password}
    switch  upr.Type {
    default:
        key = fmt.Sprintf("/auth/userpass/login/%s", upr.User)
    }
    fmt.Print(key)
    secret, err := VaultLogin(vc, key, data)
    HttpCheckPanic(err, NotAuthorized)
    user, err := NewUser(secret.Auth.ClientToken)
    HttpCheckPanic(err, InternalError)
    user.UsernamePasswordLogin(upr.User)
    return getToken(user, int64(secret.Auth.LeaseDuration) * 1000)
}


// token
func TokenLoginHandler(tr *TokenLogin) string {
    vc, err := vault.NewClient(VaultConfig)
    CheckPanic(err)
    vc.SetToken(tr.Token)
    secret, err := vc.Auth().Token().LookupSelf()
    HttpCheckPanic(err, NotAuthorized)
    CheckSecretFailIfIsNull(secret)
    type secretData struct {
        ttl int64
    }
    sd := &secretData{}
    err = mapstructure.Decode(secret.Data, sd)
    CheckPanic(err)
    user, err := NewUser(tr.Token)
    CheckPanic(err)
    user.TokenLogin()
    return getToken(user, sd.ttl * 1000)
}

// github login
func GithubLoginHandler(tr *GithubLogin, response *restful.Response, hasCSRF bool) {
    vc, err := vault.NewClient(VaultConfig)
    CheckPanic(err)
    vc.SetToken(tr.Token)
    secret, err := vc.Auth().Token().LookupSelf()
    HttpCheckPanic(err, NotAuthorized)
    CheckSecretFailIfIsNull(secret)
    type secretData struct {
        ttl int64
    }
    sd := &secretData{}
    err = mapstructure.Decode(secret.Data, sd)
    CheckPanic(err)
    user, err := NewUser(tr.Token)
    CheckPanic(err)
    user.TokenLogin()
//    setToken(user, sd.ttl * 1000, response, hasCSRF)
}

func (a *AuthResource) accountLogin(request *restful.Request, response *restful.Response) {
    entry := &AuthResource{}
    var token string
    err := request.ReadEntity(entry)
    HttpCheckPanic(err, BadRequest)
    if entry.Data == nil {
        panic(BadRequest)
    }
    if entry.AuthProvider == "token" {
        tk := &TokenLogin{}
        err = mapstructure.Decode(entry.Data, tk)
        HttpCheckPanic(err, BadRequest)
        token = TokenLoginHandler(tk)
    } else if entry.AuthProvider == "userpass" {
        up := &UPLogin{}
        err = mapstructure.Decode(entry.Data, up)
        HttpCheckPanic(err, BadRequest)
        token = UserPasswordHandler(up)
    }
    response.AddHeader("X-Menshend-Token", token)
    
}
//TODO: account[put[github]], panic handler, ui proxy[account, impersonate], proxy[api], [add static]

