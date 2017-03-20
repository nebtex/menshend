package v1

import (
    "regexp"
    . "github.com/nebtex/menshend/pkg/utils"
    . "github.com/nebtex/menshend/pkg/config"
    
    "strings"
    "fmt"
    vault "github.com/hashicorp/vault/api"
    "github.com/mitchellh/mapstructure"
)

func ValidateRegExp(s string, r string) bool {
    rc, err := regexp.Compile(r)
    HttpCheckPanic(err, InternalError)
    return rc.MatchString(s)
}

func ValidateService(s string) {
    if !ValidateRegExp(s, "^roles/[a-z0-9\\-]+/([a-z0-9\\-]+\\.)+$") {
        panic(BadRequest.Append("Invalid service").Append(s))
    }
}

func ValidateSubdomain(s string) {
    if !ValidateRegExp(s, "^([a-z0-9\\-]+\\.)+$") {
        panic(BadRequest.Append("Invalid service").Append(s))
    }
}

func ValidateRole(s string) {
    if !ValidateRegExp(s, "^[a-z0-9\\-]+$") {
        panic(BadRequest.Append("Invalid role").Append(s))
    }
}


func ValidateSecret(secretId string, user string) (vaultSecretPath string) {
    items := strings.Split(secretId, "/")
    role := items[1]
    ValidateRole(role)
    subdomain := items[2]
    ValidateSubdomain(subdomain)
    serviceId := fmt.Sprintf("roles/%s/%s", role, subdomain)
    vaultSecretPath = strings.Replace(secretId, serviceId + "/", "", 1)
    //load service
    vc, err := vault.NewClient(VaultConfig)
    HttpCheckPanic(err, InternalError)
    vc.SetToken(user)
    secret, err := vc.Logical().Read(fmt.Sprintf("%s/%s", Config.VaultPath, serviceId))
    HttpCheckPanic(err, PermissionError)
    CheckSecretFailIfIsNull(secret)
    
    nService := &AdminServiceResource{}
    err = mapstructure.Decode(secret.Data, nService)
    HttpCheckPanic(err, InternalError.Append("error decoding service"))
    //check if secret exist in service
    if !SliceStringContains(nService.SecretPaths, vaultSecretPath) {
        panic(NotFound)
    }
    return vaultSecretPath
}
