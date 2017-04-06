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
)

type GithubConfig struct {
    ClientID     string `json:"clientId"`
    ClientSecret string `json:"clientSecret"`
    BaseUrl      string `json:"baseUrl"`
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
    loginUrl := k.Scheme() + "://" + k.Uris.MenshendSubdomain + k.Host() + "/ui/login"
    return loginUrl
}
func (k *MenshendConfig) GetServicePath() string {
    loginUrl := k.Scheme() + "://" + k.Uris.MenshendSubdomain + k.Host() + "/ui/services"
    return loginUrl
}

func (k *MenshendConfig) GetLogo() string {
    return k.Space.Logo
}

func (k *MenshendConfig) GetLongDescription() string {
    return k.Space.LongDescription
}

func (k *MenshendConfig) GetShortDescription() string {
    return k.Space.Description
}

func (k *MenshendConfig) GetName() string {
    return k.Space.Name
}

var Config *MenshendConfig
var ConfigFile *string
var FlashStore *sessions.CookieStore

// generates default config
func init() {
    Config = &MenshendConfig{}
    Config.Uris = Uris{BaseUrl:"http://test.local", MenshendSubdomain: "menshend."}
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
}

func LoadConfig() error {
    //check if vault token is defined
    if os.Getenv(vault.EnvVaultToken) != "" {
        return fmt.Errorf("Menshend servar is not allowed to run when %s is defined, please check you environmental variable and delete %s", vault.EnvVaultToken, vault.EnvVaultToken)
    }
    //check if file is defined
    if ConfigFile != nil && *ConfigFile!="" {
        //load config from file
        data, err := ioutil.ReadFile(*ConfigFile)
        if err != nil {
           return err
        }
        err = yaml.Unmarshal(data, &Config)
        if err != nil {
            return  err
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
    //githubCallbackUrl := fmt.Sprintf("%s://%s/login/github/callback", Config.Scheme, Config.Host)
    //gomniauth.SetSecurityKey(Config.HashKey)
    //gomniauth.WithProviders(github.New(Config.Github.ClientID, Config.Github.ClientSecret, githubCallbackUrl))
    return nil
}
