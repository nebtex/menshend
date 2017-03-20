package strategy

import (
    "net/http"
    "github.com/nebtex/menshend/pkg/resolvers"
    vault "github.com/hashicorp/vault/api"
)

type Strategy  interface {
    Execute(resolvers.Resolver, *vault.Secret) http.Handler
}
