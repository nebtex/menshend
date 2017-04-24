package config

import (
    "github.com/gorilla/sessions"
    "strings"
    mutils "github.com/nebtex/menshend/pkg/utils"
    "net/url"
    "os"
    vault "github.com/hashicorp/vault/api"
    "github.com/Sirupsen/logrus"
    "io/ioutil"
    "github.com/ghodss/yaml"
    "fmt"
    "math"
    "github.com/markbates/goth/gothic"
    "github.com/markbates/goth/providers/github"
    "github.com/markbates/goth"
)

type GithubConfig struct {
    ClientID     string `json:"clientId"`
    ClientSecret string `json:"clientSecret"`
    Host         string `json:"host"`
}

type Uris struct {
    BaseUrl           string `json:"baseUrl"`
    MenshendSubdomain string `json:"menshendSubdomain"`
}

type Space struct {
    Logo            string `json:"logo"`
    Description     string `json:"description"`
    LongDescription string `json:"longDescription"`
    Name            string `json:"name"`
}

type MenshendConfig struct {
    HashKey     string        `json:"hashKey"`
    BlockKey    string        `json:"blockKey"`
    VaultPath   string        `json:"vaultPath"`
    Uris        Uris          `json:"uris"`
    DefaultRole string        `json:"defaultRole"`
    Github      GithubConfig  `json:"github"`
    Space       Space         `json:"space"`
}

func (k *MenshendConfig) Host() string {
    URL, err := url.Parse(k.Uris.BaseUrl)
    mutils.CheckPanic(err)
    return URL.Host
}

func (k *MenshendConfig) Scheme() string {
    URL, err := url.Parse(k.Uris.BaseUrl)
    mutils.CheckPanic(err)
    return URL.Scheme
}
func (k *MenshendConfig) HostWithoutPort() string {
    return strings.Split(k.Host(), ":")[0]
}

func (k *MenshendConfig) GetLoginPath() string {
    loginUrl := k.Scheme() + "://" + k.Uris.MenshendSubdomain + k.Host() + "/login"
    return loginUrl
}
func (k *MenshendConfig) GetServicePath() string {
    return k.Scheme() + "://" + k.Uris.MenshendSubdomain + k.Host() + "/services"
}

func (k *MenshendConfig) GetLogo() string {
    return k.Space.Logo
}

func (k *MenshendConfig) GetLongDescription() string {
    return k.Space.LongDescription
}

func (k *MenshendConfig) GetCallbackUrl(provider string, subdomain string) string {
    if subdomain != "" {
        return k.Scheme() + "://" + k.Uris.MenshendSubdomain + k.Host() + "/ui" + fmt.Sprintf("/auth/%s/callback/%s", provider, subdomain)
    }
    return k.Scheme() + "://" + k.Uris.MenshendSubdomain + k.Host() + "/ui" + fmt.Sprintf("/auth/%s/callback", provider)
}

func (k *MenshendConfig) GetShortDescription() string {
    return k.Space.Description
}

func (k *MenshendConfig) GetName() string {
    return k.Space.Name
}
func (k*MenshendConfig)GetSubdomainFullUrl(sub string) string {
    return k.Scheme() + "://" + sub + k.Host()
}

var Config *MenshendConfig
var ConfigFile *string
var FlashStore *sessions.CookieStore

// generates default config
func init() {
    Config = &MenshendConfig{}
    Config.Uris = Uris{BaseUrl:"http://test.local", MenshendSubdomain: ""}
    Config.DefaultRole = "default"
    Config.VaultPath = "secret/menshend"
    Config.HashKey = string(mutils.GenerateRandomBytes(32))
    Config.BlockKey = string(mutils.GenerateRandomBytes(32))
    
    FlashStore = sessions.NewCookieStore([]byte(Config.HashKey), []byte(Config.BlockKey))
    FlashStore.Options.Domain = "." + Config.HostWithoutPort()
    FlashStore.Options.Path = "/"
    if Config.Scheme() == "http" {
        FlashStore.Options.Secure = false
    }
    
    store := sessions.NewFilesystemStore(os.TempDir(), []byte(Config.HashKey), []byte(Config.BlockKey))
    
    // set the maxLength of the cookies stored on the disk to a larger number to prevent issues with:
    // securecookie: the value is too long
    // when using OpenID Connect , since this can contain a large amount of extra information in the id_token
    
    // Note, when using the FilesystemStore only the session.ID is written to a browser cookie, so this is explicit for the storage on disk
    store.MaxLength(math.MaxInt64)
    
    gothic.Store = store
}

func LoadConfig() error {
    //check if vault token is defined
    if os.Getenv(vault.EnvVaultToken) != "" {
        return fmt.Errorf("Menshend servar is not allowed to run when %s is defined, please check you environmental variable and delete %s", vault.EnvVaultToken, vault.EnvVaultToken)
    }
    //check if file is defined
    if ConfigFile != nil && *ConfigFile != "" {
        //load config from file
        data, err := ioutil.ReadFile(*ConfigFile)
        if err != nil {
            return err
        }
        err = yaml.Unmarshal(data, &Config)
        if err != nil {
            return err
        }
    }
    //pre-flight check
    if Config.HashKey == "" {
        Config.HashKey = string(mutils.GenerateRandomBytes(32))
    }
    if Config.BlockKey == "" {
        Config.BlockKey = string(mutils.GenerateRandomBytes(32))
    }
    if len(Config.HashKey) != 32 || len(Config.BlockKey) != 32 {
        logrus.Fatal("HashKey and BlockKey only support secret phrase of 32 characters")
    }
    
    if Config.VaultPath == "" {
        Config.VaultPath = "secret/menshend"
    }
    
    if Config.Uris.BaseUrl == "" {
        logrus.Fatal("please define the base url of menshend 'Config.Uris.BaseUrl'")
    }
    
    //pass
    FlashStore = sessions.NewCookieStore([]byte(Config.HashKey), []byte(Config.BlockKey))
    FlashStore.Options.Domain = "." + Config.HostWithoutPort()
    FlashStore.Options.Path = "/"
    if Config.Scheme() == "http" {
        FlashStore.Options.Secure = false
    }
    if (Config.Github.ClientID != "" && Config.Github.ClientSecret != "") {
        if Config.Github.Host == "" {
            Config.Github.Host = "github.com"
            
        }
        github.AuthURL = fmt.Sprintf("https://%s/login/oauth/authorize", Config.Github.Host)
        github.TokenURL = fmt.Sprintf("https://%s/login/oauth/access_token", Config.Github.Host)
        github.ProfileURL = fmt.Sprintf("https://api.%s/user", Config.Github.Host)
        github.EmailURL = fmt.Sprintf("https://api.%s/user/emails", Config.Github.Host)
        callback := Config.Scheme() + "://" + Config.Uris.MenshendSubdomain + Config.Host() + "/ui" + "/auth/github/callback"
        goth.UseProviders(
            github.New(Config.Github.ClientID, Config.Github.ClientSecret, callback, "read:org"),
        )
    }
    //TODO: check menshend subdomain
    
    store := sessions.NewFilesystemStore(os.TempDir(), []byte(Config.HashKey), []byte(Config.BlockKey))
    store.MaxLength(math.MaxInt64)
    gothic.Store = store
    return nil
}
