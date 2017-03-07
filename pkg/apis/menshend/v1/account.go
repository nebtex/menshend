package v1

import (
    "github.com/stretchr/objx"
    "net/url"
    "time"
    "github.com/Sirupsen/logrus"
    "github.com/hashicorp/vault/vault"
    "github.com/mitchellh/mapstructure"
    "fmt"
    "github.com/gorilla/mux"
    "github.com/stretchr/gomniauth"
)


/*

PUT v/account
GET v/account
DELETE v/account
*/

package menshend

import (
"net/http"
"encoding/json"
vault "github.com/hashicorp/vault/api"
"github.com/Sirupsen/logrus"
"time"
"github.com/mitchellh/mapstructure"
"fmt"
"github.com/gorilla/mux"
"github.com/stretchr/gomniauth"
"github.com/stretchr/objx"
"net/url"
)

type TokenLogin struct {
    Token string `json:"token"`
}

func MakeTimestampMillisecond() int64 {
    return time.Now().UnixNano() / int64(time.Millisecond)
}

// token
func TokenLoginHandler(w http.ResponseWriter, r *http.Request) {
    
    tr := &TokenLogin{}
    err := json.NewDecoder(r.Body).Decode(tr)
    if err != nil {
        logrus.Error(err)
        http.Redirect(w, r, Config.GetLoginPath()+"?token_error=true", 301)
        return
    }
    vc, err := vault.NewClient(VaultConfig)
    CheckPanic(err)
    vc.SetToken(tr.Token)
    secret, err := vc.Auth().Token().LookupSelf()
    if (err != nil) || (secret == nil) || (secret.Data == nil) {
        logrus.Error(err.Error())
        http.Redirect(w, r, Config.GetLoginPath()+"?token_error=true", 301)
        return
    }
    type secretData struct {
        ttl int64
    }
    sd := &secretData{}
    err = mapstructure.Decode(secret.Data, sd)
    CheckPanic(err)
    user, err := NewUser(tr.Token)
    CheckPanic(err)
    user.TokenLogin()
    setToken(user, sd.ttl * 1000, w)
    http.Redirect(w, r, Config.GetServicePath(), 301)
    return
}

type UPLoginType int

const (
    UPVault UPLoginType = iota
)

type UPLogin struct {
    User     string
    Password string
    Type     UPLoginType
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
func UserPasswordHandler(w http.ResponseWriter, r *http.Request) {
    var key string
    
    upr := &UPLogin{}
    err := json.NewDecoder(r.Body).Decode(upr)
    if err != nil {
        logrus.Error(err)
        http.Redirect(w, r, Config.GetLoginPath()+"?user_pass_error=true", 301)
        return
    }
    
    vc, err := vault.NewClient(VaultConfig)
    CheckPanic(err)
    data := map[string]interface{}{"password": upr.Password}
    switch  upr.Type {
    default:
        key = fmt.Sprintf("/auth/userpass/login/%s", upr.User)
    }
    
    secret, err := VaultLogin(vc, key, data)
    if (err != nil) || (secret == nil) {
        http.Redirect(w, r, Config.GetLoginPath()+"?user_pass_error=true", 301)
        return
    }
    
    user, err := NewUser(secret.Auth.ClientToken)
    CheckPanic(err)
    user.UsernamePasswordLogin(upr.User)
    setToken(user, int64(secret.Auth.LeaseDuration) * 1000, w)
    http.Redirect(w, r, Config.GetServicePath(), 301)
    return
}

func OauthLoginHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    provider, err := gomniauth.Provider(vars["provider"])
    CheckPanic(err)
    state := gomniauth.NewState("after", "success")
    options := objx.MSI("scope", "org")
    authUrl, err := provider.GetBeginAuthURL(state, options)
    CheckPanic(err)
    http.Redirect(w, r, authUrl, 301)
}

func urlValuesToObjectsMap(values url.Values) objx.Map {
    m := make(objx.Map)
    for k, vs := range values {
        m.Set(k, vs)
    }
    return m
}
func OauthLoginCallback(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    provider, err := gomniauth.Provider(vars["provider"])
    CheckPanic(err)
    queryParams := urlValuesToObjectsMap(r.URL.Query())
    creds, err := provider.CompleteAuth(queryParams)
    CheckPanic(err)
    user, err := provider.GetUser(creds)
    CheckPanic(err)
    fmt.Println(user)
}

