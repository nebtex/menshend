package kuper

import (
    "runtime"
    log "github.com/Sirupsen/logrus"
    vault "github.com/hashicorp/vault/api"
    "github.com/ansel1/merry"
    "github.com/stretchr/gomniauth"
    "github.com/stretchr/gomniauth/providers/github"
    "fmt"
    "github.com/gorilla/sessions"
)

const (
    //GitHubProvider means that the auth method used was github
    GitHubProvider = "github"
    //TokenProvider means that the user used a vault token
    TokenProvider = "token"
    //UsernamePasswordProvider vault user/password auth
    UsernamePasswordProvider = "username/password"
    //CSRFTokenLen size of the CSRF token
    CSRFTokenLen = 32
)
//MySecretKey secret key for sign the jwt token, this global is generated
//automatically each time that the server start
//A restart of the server means that all the issued keys will become invalid
var MySecretKey []byte

type GithubConfig struct {
    ClientID     string
    ClientSecret string
}
type KuperConfig struct {
    Salt         string
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

var Config *KuperConfig
var VaultConfig *vault.Config
var FlashStore *sessions.CookieStore

func init() {
    Config = &KuperConfig{}
    Config.Host = "test.local"
    Config.InterfaceURL = "http://kuper.test.local/ui/"
    Config.Scheme = "http"
    Config.VaultPath = "secret/kuper"
    Config.DefaultTTL = 24 * 60 * 60 * 1000
    Config.Salt = GenerateRandomString(32)
    Config.ListenPort = 8080
    
    VaultConfig = vault.DefaultConfig()
    
    githubCallbackUrl := fmt.Sprintf("%s://%s/login/github/callback", Config.Scheme, Config.Host)
    gomniauth.SetSecurityKey(Config.Salt)
    gomniauth.WithProviders(github.New(Config.Github.ClientID, Config.Github.ClientSecret, githubCallbackUrl))
    FlashStore = sessions.NewCookieStore([]byte(Config.Salt))
    FlashStore.Options.Domain = "." + Config.Host
    FlashStore.Options.Path = "/"
    if Config.Scheme == "http" {
        FlashStore.Options.Secure = false
    }
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

//InvalidJWT
var InvalidJWT = merry.New("jwt is invalid")
//JWTNotFound is not available
var JWTNotFound = merry.New("Could not read the jwt cookie")
//ServiceNotFound not backend defined for service
var ServiceNotFound = merry.New("I could not find the service")
//LuaScriptFailed the lua script failed
var LuaScriptFailed = merry.New("There is a issue with the lua script")
//InvalidSubdomain the subdomain provided is invalid
var InvalidSubdomain = merry.New("subdomain provided is invalid ")
//PermissionError this mean that the acl token has not access to x key on consul
var PermissionError = merry.New("User has not access to resource")
//InvalidUrl url returned is not valid
var InvalidUrl = merry.New("The url returned by the lua script is not valid")
//InvalidService service definition on consul is invalid
var InvalidService = merry.New("service definition on vault is invalid")
//InactiveService service has been manually deactivated by the admin
var InactiveService = merry.New("service is disabled")
//BadBackendUrl ...
var BadBackendUrl = merry.New("Backend return a invalid url")
