package kuper

import (
    "github.com/dgrijalva/jwt-go"
    "github.com/ansel1/merry"
    "fmt"
    "net/http"
    "context"
    "github.com/Sirupsen/logrus"
    "encoding/json"
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
    //person that is impersonating this user
    ImpersonatedBy             string `json:"ipb,omitempty"`
    ImpersonatedByGroups       []string `json:"ipbg,omitempty"`
    ImpersonatedByAuthProvider string `json:"ipbap,omitempty"`
}

func (u *User) Valid() error {
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
func NewUser(acl string) (*User, merry.Error) {
    user := &User{Token:acl}
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
    tokenString, err := token.SignedString([]byte(Config.HashKey))
    CheckPanic(err)
    return tokenString
}

func (u *User)SetExpiresAt(value int64) {
    u.ExpiresAt = value
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
        return  []byte(Config.HashKey), nil
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


//PanicHandler middleware for recover from panic attacks
func PanicHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            rec := recover()
            if rec != nil {
                logrus.Errorf("%v", rec)
                switch v := rec.(type) {
                case merry.Error:
                    subDomain := getSubDomain(r.URL.Host)
                    if len(subDomain) == 0 {
                        w.WriteHeader(500)
                        response := Response{Success:false,
                            Message:v.Error()}
                        data, err := json.Marshal(response)
                        if (err != nil) {
                            logrus.Error(err)
                        }
                        w.Write(data)
                        return
                        
                    } else {
                        // Get a session.
                        session, err := FlashStore.Get(r, "errors")
                        if err != nil {
                            http.Error(w, err.Error(), http.StatusInternalServerError)
                            logrus.Error(err)
                            return
                        }
                        session.AddFlash(v.Error())
                        session.Save(r, w)
                        redirectUrl := fmt.Sprintf("%s://%s/ui/login", Config.Scheme, Config.Host)
                        http.Redirect(w, r, redirectUrl, 301)
                        return
                    }
                
                default:
                    w.WriteHeader(500)
                    errMsg := `{"success":false, "message": "Unkown error (please contact your admin)"}`
                    w.Write([]byte(errMsg))
                    return
                }
            }
        }()
        next.ServeHTTP(w, r)
    })
}
