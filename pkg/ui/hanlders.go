package main

import (
    "github.com/stretchr/gomniauth"
    "github.com/stretchr/objx"
    "net/url"
    "github.com/gorilla/mux"
    "fmt"
)

func main() {
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
