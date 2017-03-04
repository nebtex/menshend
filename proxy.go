package menshend
//add audit log in squilite
//add configuration headers(middleware)
/*
import (
	"github.com/vulcand/oxy/forward"
	"net/http"
)
//ProxyHandler forward request to the backend services
func ProxyHandler(fwd *forward.Forwarder) *http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value("Username").(*User)
		if !ok {
			
		}
		subDomain := getSubDomain(r.URL.Host)
		url, err := GetBackend(user, subDomain)
		if err!=nil{
			
		}
		r.URL = url
		fwd.ServeHTTP(w, r)
	}
	return handler
}
*/
