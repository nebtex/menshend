package kuper

import (
    "runtime"
    log "github.com/Sirupsen/logrus"
    vault "github.com/hashicorp/vault/api"
    consul "github.com/hashicorp/consul/api"
    "github.com/ansel1/merry"
    "os"
    "strings"
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

type KuperConfig struct {
    InterfaceURL string
    VaultPath    string
    Scheme       string
    BaseDomain   string
}

var Config *KuperConfig
var VaultConfig *vault.Config

var ConsulClient *consul.Client

func init() {
    var err error
    Config = &KuperConfig{}
    Config.BaseDomain = "test.local"
    Config.InterfaceURL = "http://kuper.test.local/ui/"
    Config.Scheme = "http"
    Config.VaultPath = "secret/kuper"
    
    VaultConfig = vault.DefaultConfig()
    ConsulClient, err = consul.NewClient(consul.DefaultConfig())
    CheckPanic(err)
    
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
var InvalidService = merry.New("service definition on consul is invalid")
//InactiveService service has been manually deactivated by the admin
var InactiveService = merry.New("service is disabled")
//BadBackendUrl ...
var BadBackendUrl = merry.New("Backend return a invalid url")

var KVStoreRoles = []string{}

func init() {
    if ksr := os.Getenv("KV_STORE_ROLES"); ksr != "" {
        KVStoreRoles = strings.Split(ksr, ",")
    }
}
