package kuper
/*
import (
    "net/http"
    "encoding/json"
    vault "github.com/hashicorp/vault/api"
    "fmt"
    "github.com/Sirupsen/logrus"
    "strings"
)

type TokenLogin struct {
    Token string
}

//GetConsulAclToken return the consul acl token with highest access available
//for the VaultToken
func GetConsulAclToken(VaultToken string) (token string, expireIn int,
err error) {
    var secret *vault.Secret
    vaultClient, err := vault.NewClient(vault.DefaultConfig())
    CheckPanic(err)
    vaultClient.SetToken(VaultToken)
    fmt.Println(KVStoreRoles)
    for _, r := range KVStoreRoles {
        secret, err := vaultClient.Logical().Read(strings.TrimSpace(r))
        fmt.Println(secret!=nil, err, "###################")
        if err != nil {
            logrus.Error(err)
        }
        if secret != nil {
            break
        }
    }
    fmt.Println(secret)
    
    if secret == nil {
        return "", 0, fmt.Errorf("%s", "you can't use any consul role")
    }
    
    v, ok := secret.Data["token"].(string)
    if !ok {
        CheckPanic(fmt.Errorf("%s", "Vault token has a bad format"))
    }
    return v, secret.LeaseDuration, nil
}

// token
func TokenLoginHandler(w http.ResponseWriter, r *http.Request) {
    tl := &TokenLogin{}
    err := json.NewDecoder(r.Body).Decode(tl)
    if err != nil {
        errMsg := `{"success":false, "message": "Please send a valid json"}`
        w.Write([]byte(errMsg))
        return
    }
}

// user/password


// github
*/
