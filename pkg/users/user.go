package users

import (
    "github.com/dgrijalva/jwt-go"
    "github.com/ansel1/merry"
    "fmt"
    . "github.com/nebtex/menshend/pkg/config"
    . "github.com/nebtex/menshend/pkg/utils"
)

//InvalidJWT
var InvalidJWT = merry.New("jwt is invalid")
//JWTNotFound is not available

type AuthProviderType string

const (
    //GitHubProvider means that the auth method used was github
    GitHubProvider AuthProviderType = "github"
    //TokenProvider means that the user used a vault token
    TokenProvider = "token"
    //UsernamePasswordProvider vault user/password auth
    UsernamePasswordProvider = "userpass"
)

var AllAuthProviders = []AuthProviderType{
    GitHubProvider,
    TokenProvider,
    UsernamePasswordProvider,
}

func SliceAuthProviderTypeContains(s []AuthProviderType, e AuthProviderType) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

type JwtImpersonateInfo struct {
    //person that is impersonating this user
    Username     string `json:"usr,omitempty"`
    //user group only works with github backend
    Groups       []string `json:"grp,omitempty"`
    //github, token or user/password
    AuthProvider AuthProviderType `json:"atp,omitempty"`
}

type JwtMenshendInfo struct {
    //username only works with github and user/password backend
    Username      string `json:"usr,omitempty"`
    //vault
    VaultToken    string `json:"alt,omitempty"`
    //user group only works with github backend
    Groups        []string `json:"grp,omitempty"`
    //github, token or user/password
    AuthProvider  AuthProviderType `json:"atp,omitempty"`
    
    // Indicate that the current user is currently impersonate by other
    ImpersonateBy *JwtImpersonateInfo
}

type User struct {
    jwt.StandardClaims
    Menshend    *JwtMenshendInfo `json:"-"`
    EncodedData string `json:"dta,omitempty"`
}

func (u *User) Valid() error {
    u.Decode()
    
    if len(u.Menshend.VaultToken) < 1 {
        return fmt.Errorf("%s", "jwt token has not an vault token")
    }
    
    if u.Menshend.ImpersonateBy != nil {
        if (len(u.Menshend.Username) == 0) {
            return fmt.Errorf("%s", "ImpersonatedBy  only allowed if the username is defined")
        }
    }
    switch u.Menshend.AuthProvider {
    case TokenProvider:
        if len(u.Menshend.Groups) > 0 {
            return fmt.Errorf("%s", "Currupted token")
        }
        if len(u.Menshend.Username) > 0 {
            return fmt.Errorf("%s", "Currupted token")
        }
    case UsernamePasswordProvider:
        if len(u.Menshend.Groups) > 0 {
            return fmt.Errorf("%s", "Currupted token")
        }
        if len(u.Menshend.Username) == 0 {
            return fmt.Errorf("%s", "Currupted token")
        }
    case GitHubProvider:
        if len(u.Menshend.Username) == 0 {
            return fmt.Errorf("%s", "Currupted token")
        }
    default:
        return fmt.Errorf("%s", "Currupted token")
    }
    return u.StandardClaims.Valid()
}
//NewUser create a new user struct
func NewUser(vaultToken string) (*User, merry.Error) {
    user := &User{}
    user.Menshend = &JwtMenshendInfo{}
    user.Menshend.VaultToken = vaultToken
    return user, nil
}

///GitHubLogin save github important info
func (u *User)GitHubLogin(un string, gs ...string) {
    u.Menshend.Username = un
    u.Menshend.Groups = gs
    u.Menshend.AuthProvider = GitHubProvider
}

///TokenLogin ...
func (u *User)TokenLogin() {
    u.Menshend.Groups = []string{}
    u.Menshend.AuthProvider = TokenProvider
}

///UsernamePasswordLogin ...
func (u *User)UsernamePasswordLogin(un string) {
    u.Menshend.Username = un
    u.Menshend.Groups = []string{}
    u.Menshend.AuthProvider = UsernamePasswordProvider
}
//GenerateJWT create a new jwt token
func (u *User)GenerateJWT() string {
    u.Encoded()
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, u)
    
    // Sign and get the complete encoded token as a string using the secret
    tokenString, err := token.SignedString([]byte(Config.HashKey))
    CheckPanic(err)
    return tokenString
}

func (u *User)Encoded() {
    ed, err := SecureCookie.Encode("X-Menshend-Token", u.Menshend)
    CheckPanic(err)
    u.EncodedData = ed
}

func (u *User)Decode() {
    dst := &JwtMenshendInfo{}
    err := SecureCookie.Decode("X-Menshend-Token", u.EncodedData, dst)
    CheckPanic(err)
    u.Menshend = dst
}

func (u *User)SetExpiresAt(value int64) {
    u.ExpiresAt = value
}

//Impersonate ..
func (u *User) Impersonate(authProvider *AuthProviderType, user *string, groups ...string) error {
    
    //validate input
    if user == nil || len(*user) == 0 {
        return fmt.Errorf("%s", "new username is missing")
    }
    
    if authProvider == nil || !SliceAuthProviderTypeContains(AllAuthProviders, *authProvider) {
        return fmt.Errorf("%s", "invalid authProvider")
    }
    
    if (*authProvider == UsernamePasswordProvider) || (*authProvider == TokenProvider) {
        if len(groups) > 0 {
            return fmt.Errorf("%s", "this  auth provider does not support groups")
        }
    }
    
    if len(u.Menshend.Username) == 0 {
        if (u.Menshend.ImpersonateBy == nil) {
            return fmt.Errorf("%s", "only tokens with an username can use the impersonate feature")
        }
    }
    
    if u.Menshend.ImpersonateBy == nil {
        u.Menshend.ImpersonateBy = &JwtImpersonateInfo{}
        u.Menshend.ImpersonateBy.Username = u.Menshend.Username
        u.Menshend.ImpersonateBy.Groups = u.Menshend.Groups
        u.Menshend.ImpersonateBy.AuthProvider = u.Menshend.AuthProvider
    }
    
    u.Menshend.AuthProvider = *authProvider
    u.Menshend.Username = *user
    u.Menshend.Groups = groups
    return nil
}
//StopImpersonate ..
func (u *User)StopImpersonate() {
    if u.Menshend.ImpersonateBy == nil {
        return
    }
    
    u.Menshend.Username = u.Menshend.ImpersonateBy.Username
    u.Menshend.Groups = u.Menshend.ImpersonateBy.Groups
    u.Menshend.AuthProvider = u.Menshend.ImpersonateBy.AuthProvider
    u.Menshend.ImpersonateBy = nil
    
    return
}
//FromJWT read a jwt token, parse and validate it then obtain an User
func FromJWT(jwtRaw string) (*User, merry.Error) {
    token, err := jwt.ParseWithClaims(jwtRaw, &User{}, func(token *jwt.Token) (interface{}, error) {
        // Don't forget to validate the alg is what you expect:
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("Unexpected signing method: %v",
                token.Header["alg"])
        }
        // MySecretKey is a []byte containing your secret,
        // e.g. []byte("my_secret_key")
        return []byte(Config.HashKey), nil
    })
    
    if err != nil {
        return nil, InvalidJWT.Append(err.Error()).WithValue("kuper-jwt", jwtRaw)
    }
    claims, _ := token.Claims.(*User);
    return claims, nil
    
}

