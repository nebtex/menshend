package kuper

import (
    "net/http"
    "github.com/ansel1/merry"
    "time"
    "fmt"
    vault "github.com/hashicorp/vault/api"
    "context"
)

//InvalidFormError form validation error
var InvalidFormError = merry.New("Form is invalid, some field is missing")
//InvalidRequest request is not valid
var InvalidRequest = merry.New("request is not valid ")
//InvalidRequest
var UserNotFound = merry.New("User not found in the request context")

func GetUserFromContext(c context.Context) (*User) {
    user, ok := c.Value("User").(*User)
    if !ok {
        panic(UserNotFound)
    }
    return user
}

//ImpersonateHandler allow ro change the username and groups of the token
//only user with access to the KuperPath/Impersonate. can use this endpoint
func ImpersonateHandler(w http.ResponseWriter, r *http.Request) {
    user := GetUserFromContext(r.Context())
    if len(user.Username) == 0 {
        panic(InvalidRequest.Append("Only tokens with an username can use " +
            "the impersonate feature"))
    }
    checkImpersonatePermission(user, VaultConfig)
    err := r.ParseForm()
    CheckPanic(err)
    newUsername := r.PostForm["username"]
    authProvider := r.PostForm["authProvider"]
    if len(newUsername) == 0 {
        panic(InvalidFormError.Append("username is missing"))
    }
    if len(authProvider) == 0 {
        panic(InvalidFormError.Append("authProvider is missing"))
    }
    
    if authProvider[0] == TokenProvider {
        panic(InvalidRequest.Append("Is not possible to impersonate user that " +
            "use the auth token provider method"))
    }
    newGroups := r.PostForm["group"]
    
    if authProvider[0] == UsernamePasswordProvider {
        if len(newGroups) > 0 {
            panic(InvalidRequest.Append("This auth provider don't " +
                "support groups"))
        }
    }
    
    user.ImpersonatedBy = user.Username
    user.ImpersonatedByGroups = user.Groups
    user.ImpersonatedByAuthProvider = user.AuthProvider
    user.Username = newUsername[0]
    user.Groups = newGroups
    user.AuthProvider = authProvider[0]
    setToken(user, 0, w)
    w.WriteHeader(200)
}
//setToken ..
// expiresIn in milliseconds
func setToken(u *User, expiresIn int64, w http.ResponseWriter) {
    expireAt := MakeTimestampMillisecond()
    if expiresIn == 0 {
        expireAt += Config.DefaultTTL
    } else {
        expireAt += expiresIn
    }
    u.SetExpiresAt(expireAt)
    ct := &http.Cookie{Path: "/", Name: "kuper-jwt", Value: u.GenerateJWT(),
        Expires: time.Unix(u.ExpiresAt / 1000, 0),
        HttpOnly:true }
    
    ct.Domain = "." + Config.Host
    
    if Config.Scheme == "https" {
        ct.Secure = true
    }
    http.SetCookie(w, ct)
}

//checkImpersonatePermission check if the user can impersonate other
func checkImpersonatePermission(u *User, vc *vault.Config) {
    key := fmt.Sprintf("%s/%s", Config.VaultPath, "Impersonate")
    client, err := vault.NewClient(vc)
    CheckPanic(err)
    client.SetToken(u.Token)
    cap, err := client.Sys().CapabilitiesSelf(key)
    if err != nil {
        panic(PermissionError.Append(err.Error()).WithValue("user", u))
    }
    if !((SliceStringContains(cap, "read")) ||
        (SliceStringContains(cap, "write")) ||
        (SliceStringContains(cap, "update")) ||
        (SliceStringContains(cap, "root"))) {
        panic(PermissionError.Append(err.Error()).WithValue("user", u))
    }
}
