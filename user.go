package kuper

import (
    "github.com/dgrijalva/jwt-go"
    "github.com/ansel1/merry"
    "fmt"
    "encoding/base64"
    "net/http"
    "context"
)

type User struct {
    jwt.StandardClaims
    //username only works with github and user/password backend
    Username                   string `json:"usr,omitempty"`
    //vault
    Token                      string `json:"alt,omitempty"`
    //user group only works with github backend
    Groups                     []string `json:"grp,omitempty"`
    //github, token or user/password
    AuthProvider               string `json:"atp,omitempty"`
    //CSRF token for this jwt token (used in the kuper ui interface)
    CSRFToken                  string `json:"cst,omitempty"`
    //person that is impersonating this user
    ImpersonatedBy             string `json:"ipb,omitempty"`
    ImpersonatedByGroups       []string `json:"ipbg,omitempty"`
    ImpersonatedByAuthProvider string `json:"ipbap,omitempty"`
}

func (u *User) Valid() error {
    csrf, err := base64.URLEncoding.DecodeString(u.CSRFToken)
    if err != nil {
        return err
    }
    if len(csrf) != CSRFTokenLen {
        return fmt.Errorf("jwt token has not a csfr token")
    }
    if len(u.Token) < 1 {
        return fmt.Errorf("jwt token has not an acl token")
    }
    if len(u.AuthProvider) < 1 {
        return fmt.Errorf("jwt token is currupted")
    }
    if len(u.ImpersonatedBy) > 0 {
        if (len(u.Username) < 1) || (len(u.AuthProvider) < 0) {
            return fmt.Errorf("ImpersonatedBy  only allowed if" +
                " the username and authprovider  are defined")
        }
    }
    switch u.AuthProvider {
    case TokenProvider:
        if len(u.Username) > 0 || len(u.Groups) > 0 {
            return fmt.Errorf("Currupted token")
        }
    case UsernamePasswordProvider:
        if len(u.Groups) > 0 {
            return fmt.Errorf("Currupted token")
        }
    default:
    }
    return u.StandardClaims.Valid()
}
//NewUser create a new user struct
func NewUser(acl string, expireAt int64) (*User, merry.Error) {
    user := &User{Token:acl}
    user.ExpiresAt = expireAt
    user.CSRFToken = GenerateRandomString(CSRFTokenLen)
    
    return user, nil
}

///GitHubLogin save github important info
func (u *User)GitHubLogin(un string, gs ...string) {
    u.Username = un
    u.Groups = gs
    u.AuthProvider = GitHubProvider
}
///TokenLogin ...
func (u *User)TokenLogin() {
    u.Username = ""
    u.Groups = []string{}
    u.AuthProvider = TokenProvider
}

///UsernamePasswordLogin ...
func (u *User)UsernamePasswordLogin(un string) {
    u.Username = un
    u.Groups = []string{}
    u.AuthProvider = UsernamePasswordProvider
}
//GenerateJWT create a new jwt token
func (u *User)GenerateJWT() string {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, u)
    
    // Sign and get the complete encoded token as a string using the secret
    tokenString, err := token.SignedString(MySecretKey)
    CheckPanic(err)
    return tokenString
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
        return MySecretKey, nil
    })
    
    if err != nil {
        return nil, InvalidJWT.Append(err.Error()).WithValue("kuper-jwt", jwtRaw)
    }
    claims, _ := token.Claims.(*User);
    return claims, nil
    
}

func getUserFromRequest(r *http.Request) (*User, merry.Error) {
    jwtCookie, err := r.Cookie("kuper-jwt")
    if err != nil {
        return nil, JWTNotFound.Append(err.Error())
    }
    user, mErr := FromJWT(jwtCookie.Value)
    if mErr != nil {
        return nil, mErr
    }
    
    return user, nil
}

//NeedLogin auth middleware, for router that need the jwt token
func NeedLogin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user, err := getUserFromRequest(r)
        if err != nil {
            panic(err)
        }
        ctx := context.WithValue(r.Context(), "User", user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
