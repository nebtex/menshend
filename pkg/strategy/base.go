package strategy

import "github.com/nebtex/menshend/pkg/backend"
import (
    "net/http"
)

type Strategy  interface {
    Execute(backend.Backend) http.HandlerFunc
}
