package menshend

import (
    "runtime"
    log "github.com/Sirupsen/logrus"
    vault "github.com/hashicorp/vault/api"
    "github.com/ansel1/merry"
    "github.com/stretchr/gomniauth"
    "github.com/stretchr/gomniauth/providers/github"
    "fmt"
    "github.com/gorilla/sessions"
    "github.com/gorilla/securecookie"
    "strings"
)

const (
    //GitHubProvider means that the auth method used was github
    GitHubProvider = "github"
    //TokenProvider means that the user used a vault token
    TokenProvider = "token"
    //UsernamePasswordProvider vault user/password auth
    UsernamePasswordProvider = "username/password"
)
//MySecretKey secret key for sign the jwt token, this global is generated
//automatically each time that the server start
//A restart of the server means that all the issued keys will become invalid

type GithubConfig struct {
    ClientID     string
    ClientSecret string
}
type menshendConfig struct {
    HashKey      string
    BlockKey     string
    ListenPort   int
    InterfaceURL string
    VaultPath    string
    Scheme       string
    Host         string
    Github       GithubConfig
    //default time to live for the jwt token in seconds
    //this value will be used only when the expiration time cant be guessed
    //using the vault api
    DefaultTTL   int64
}

func (k *menshendConfig) HostWithoutPort() string {
    return strings.Split(k.Host, ":")[0]
}
func (k *menshendConfig) GetLoginPath() string {
    loginUrl := k.Scheme + "://" + k.Host + "/ui/login"
    return loginUrl
}
func (k *menshendConfig) GetServicePath() string {
    loginUrl := k.Scheme + "://" + k.Host + "/ui/services"
    return loginUrl
}

var Config *menshendConfig
var VaultConfig *vault.Config
var FlashStore *sessions.CookieStore
var SecureCookie *securecookie.SecureCookie

func init() {
    Config = &menshendConfig{}
    Config.Host = "test.local"
    Config.InterfaceURL = "http://menshend.test.local/ui/"
    Config.Scheme = "http"
    Config.VaultPath = "secret/menshend"
    Config.DefaultTTL = 24 * 60 * 60 * 1000
    Config.HashKey = GenerateRandomString(32)
    Config.BlockKey = GenerateRandomString(32)
    
    Config.ListenPort = 8080
    
    VaultConfig = vault.DefaultConfig()
    
    githubCallbackUrl := fmt.Sprintf("%s://%s/login/github/callback", Config.Scheme, Config.Host)
    gomniauth.SetSecurityKey(Config.HashKey)
    gomniauth.WithProviders(github.New(Config.Github.ClientID, Config.Github.ClientSecret, githubCallbackUrl))
    FlashStore = sessions.NewCookieStore([]byte(Config.HashKey), []byte(Config.BlockKey))
    FlashStore.Options.Domain = "." + Config.HostWithoutPort()
    FlashStore.Options.Path = "/"
    if Config.Scheme == "http" {
        FlashStore.Options.Secure = false
    }
    SecureCookie = securecookie.New([]byte(Config.HashKey), []byte(Config.BlockKey))
}

//CheckPanic if error exist exit and log the file and line from where
//the errors comes
func CheckPanic(e error) {
    _, file, line, _ := runtime.Caller(1)
    if e != nil {
        log.WithFields(log.Fields{"file": file,
            "line": line,
        }).Panic(e)
    }
}


//Errors


//ServiceNotFound not backend defined for service
var ServiceNotFound = merry.New("I could not find the service")
//LuaScriptFailed the lua script failed
var LuaScriptFailed = merry.New("There is a issue with the lua script")
//InvalidSubdomain the subdomain provided is invalid
var InvalidSubdomain = merry.New("subdomain provided is invalid ")
//PermissionError this mean that the acl token has not access to x key on consul
var PermissionError = merry.New("Permission Error").WithHTTPCode(403)
//InvalidUrl url returned is not valid
var InvalidUrl = merry.New("The url returned by the lua script is not valid")
//InvalidService service definition on consul is invalid
var InvalidService = merry.New("service definition on vault is invalid")
//InactiveService service has been manually deactivated by the admin
var InactiveService = merry.New("service is disabled")
//BadBackendUrl ...
var BadBackendUrl = merry.New("Backend return a invalid url")
//BadRequest ...
var BadRequest = merry.New("Bad request").WithHTTPCode(400)

//BadRequest ...
var InternalError = merry.New("Internal Error").WithHTTPCode(500)
var NotFound = merry.New("Resource not found").WithHTTPCode(404)
