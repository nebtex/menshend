package v1

import (
    "regexp"
    mutils "github.com/nebtex/menshend/pkg/utils"
    "github.com/nebtex/menshend/pkg/config"
    "strings"
    "fmt"
    vault "github.com/hashicorp/vault/api"
    "github.com/mitchellh/mapstructure"
)

//ValidateRegExp helper for validate any regexp
func ValidateRegExp(s string, r string) bool {
    rc, err := regexp.Compile(r)
    mutils.HttpCheckPanic(err, mutils.InternalError)
    return rc.MatchString(s)
}

//ValidateService validate the service id
func ValidateService(s string) {
    if !ValidateRegExp(s, "^roles/[a-z0-9\\-]+/([a-z0-9\\-]+\\.)+$") {
        panic(mutils.BadRequest.WithUserMessage("Invalid service").WithUserMessage(fmt.Sprintf("Invalid service: %s", s)))
    }
}

//ValidateSubdomain subdomain should end with . and only contains alphanumeric characters and -
func ValidateSubdomain(s string) {
    if !ValidateRegExp(s, "^([a-z0-9\\-]+\\.)+$") {
        panic(mutils.BadRequest.WithUserMessage("Invalid subdmain").WithUserMessage(fmt.Sprintf("Invalid subdmain: %s", s)))
    }
}

//ValidateRole validate roles names
func ValidateRole(s string) {
    if !ValidateRegExp(s, "^[a-z0-9\\-]+$") {
        panic(mutils.BadRequest.WithUserMessage("Invalid role").WithUserMessage(fmt.Sprintf("Invalid role: %s", s)))
    }
}

//ValidateSecret check if a vault secret is associated with the service, panic in case of not
func ValidateSecret(secretID string, user string) (vaultSecretPath string) {
    items := strings.Split(secretID, "/")
    role := items[1]
    ValidateRole(role)
    subdomain := items[2]
    ValidateSubdomain(subdomain)
    serviceID := fmt.Sprintf("roles/%s/%s", role, subdomain)
    vaultSecretPath = strings.Replace(secretID, serviceID + "/", "", 1)
    //load service
    vc, err := vault.NewClient(vault.DefaultConfig())
    mutils.HttpCheckPanic(err, mutils.InternalError)
    vc.SetToken(user)
    secret, err := vc.Logical().Read(fmt.Sprintf("%s/%s", config.Config.VaultPath, serviceID))
    mutils.HttpCheckPanic(err, mutils.PermissionError)
    CheckSecretFailIfIsNull(secret)
    
    nService := &AdminServiceResource{}
    err = mapstructure.Decode(secret.Data, nService)
    mutils.HttpCheckPanic(err, mutils.InternalError.WithUserMessage("error decoding service"))
    //check if secret exist in service
    if !mutils.SliceStringContains(nService.SecretPaths, vaultSecretPath) {
        panic(mutils.NotFound)
    }
    return vaultSecretPath
}
