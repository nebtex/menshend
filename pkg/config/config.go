package config

import (
    vault "github.com/hashicorp/vault/api"
    "github.com/stretchr/gomniauth"
    "github.com/stretchr/gomniauth/providers/github"
    "fmt"
    "github.com/gorilla/sessions"
    "github.com/gorilla/securecookie"
    "strings"
    . "github.com/nebtex/menshend/pkg/utils"
    "os"
    "net/url"
)

type GithubConfig struct {
    ClientID     string
    ClientSecret string
}

type Uris struct {
    BaseUrl     string
    UiSubdomain string
    Api         string
}

type MenshendConfig struct {
    MenshendSubdomain string
    HashKey           string
    BlockKey          string
    ListenPort        int
    DevMode           bool
    VaultPath         string
    EnableUI          bool
    Uris              *Uris
    DefaultRole       string
    DefaultRoleMap    map[string]string
    Github            GithubConfig
    Logo              string
    Description       string
    LongDescription   string
    Name              string
    //default time to live for the jwt token in seconds
    //this value will be used only when the expiration time cant be guessed
    //using the vault api
    DefaultTTL        int64
}

func (k *MenshendConfig) Host() string {
    URL, err := url.Parse(k.Uris.BaseUrl)
    CheckPanic(err)
    return URL.Host
}

func (k *MenshendConfig) Scheme() string {
    URL, err := url.Parse(k.Uris.BaseUrl)
    CheckPanic(err)
    return URL.Scheme
}
func (k *MenshendConfig) HostWithoutPort() string {
    return strings.Split(k.Host(), ":")[0]
}

func (k *MenshendConfig) GetLoginPath() string {
    loginUrl := k.Scheme() + "://" + k.Uris.UiSubdomain + k.Host() + "/ui/login"
    return loginUrl
}
func (k *MenshendConfig) GetServicePath() string {
    loginUrl := k.Scheme() + "://" + k.Uris.UiSubdomain + k.Host() + "/ui/services"
    return loginUrl
}

func (k *MenshendConfig) GetLogo() string {
    return k.Logo
}

func (k *MenshendConfig) GetLongDescription() string {
    return k.LongDescription
}

func (k *MenshendConfig) GetShortDescription() string {
    return k.Description
}

func (k *MenshendConfig) GetName() string {
    return k.Name
}

var Config *MenshendConfig
var VaultConfig *vault.Config
var FlashStore *sessions.CookieStore
var SecureCookie *securecookie.SecureCookie

func init() {
    Config = &MenshendConfig{}
    Config.Uris = &Uris{BaseUrl:"http://test.local", UiSubdomain:"menshend.", Api: "http://test.local"}
    Config.DevMode = os.Getenv("MENSHEND_DEV_MODE") != ""
    Config.DefaultRole = "default"
    Config.VaultPath = "secret/menshend"
    Config.DefaultTTL = 24 * 60 * 60 * 1000
    Config.HashKey = string(GenerateRandomBytes(32))
    Config.BlockKey = string(GenerateRandomBytes(32))
    Config.ListenPort = 8080
    VaultConfig = vault.DefaultConfig()
    githubCallbackUrl := fmt.Sprintf("%s://%s/login/github/callback", Config.Scheme, Config.Host)
    gomniauth.SetSecurityKey(Config.HashKey)
    gomniauth.WithProviders(github.New(Config.Github.ClientID, Config.Github.ClientSecret, githubCallbackUrl))
    FlashStore = sessions.NewCookieStore([]byte(Config.HashKey), []byte(Config.BlockKey))
    FlashStore.Options.Domain = "." + Config.HostWithoutPort()
    FlashStore.Options.Path = "/"
    if Config.Scheme() == "http" {
        FlashStore.Options.Secure = false
    }
    SecureCookie = securecookie.New([]byte(Config.HashKey), []byte(Config.BlockKey))
    
}
