package strategy

import (
    "github.com/nebtex/menshend/pkg/backend"
    . "github.com/nebtex/menshend/pkg/utils"
    "net/http"
    "net/url"
    "github.com/ansel1/merry"
)

var InternalError = merry.New("Internal Error").WithHTTPCode(500)

type Redirect struct {
    
}

func (r *Redirect) Execute(b backend.Backend) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        for key, value := range b.Headers() {
            w.Header().Set(key, value)
        }
        newUrl, err := url.Parse(r.URL.String())
        HttpCheckPanic(err, InternalError)
        bUrl, err := url.Parse(b.BaseUrl())
        
        newUrl.Host = bUrl.Host
        newUrl.User = bUrl.User
        newUrl.Scheme = bUrl.Scheme
        http.Redirect(w, r, newUrl.String(), 301)
    }
}
