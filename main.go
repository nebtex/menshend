package kuper
/*
import (
	"fmt"
	"net/http"
	oauth2 "github.com/goincremental/negroni-oauth2"
	sessions "github.com/goincremental/negroni-sessions"
	"github.com/goincremental/negroni-sessions/cookiestore"
	"github.com/urfave/negroni"
	"github.com/gorilla/mux"
	"time"
)
const PermissionErrorRedirect = "/?service=%s&&error=true&&errorType=permission"
const BackendErrorRedirect = "/?service=%s&&error=true&&errorType=backend"
const MySecretKey = "a"
var BaseDomain = "test.local"

func KuperServerHandlers(r *mux.Route) {
	//r.Handler("/graphql", func(){})
	//r.Handler("/ui/*", func(){})
}
func ProxyHandlers(r *mux.Route) {
	r.Handler("/*", func(){})
}

func Server(){
	r := mux.NewRouter()
	KuperServerHandlers(r.Host(BaseDomain).Subrouter())
	KuperServerHandlers(r.Host("kuper." + BaseDomain).Subrouter())
	ProxyHandlers(r.Host("{subdomain:.+}." + BaseDomain).Subrouter())
}
func main() {
	
/*	//load kuper server handlers
	KuperServerHandlers(r.Host(BaseDomain).Subrouter())
	KuperServerHandlers(r.Host("kuper." + BaseDomain).Subrouter())
	
	appR := mux.NewRouter()
	appR.Host("{subdomain:.*}." + BaseDomain).Subrouter()
	
	secureMux := http.NewServeMux()
	
	// Routes that require a logged in user
	// can be protected by using a separate route handler
	// If the user is not authenticated, they will be
	// redirected to the login path.
	secureMux.HandleFunc("/restrict", func(w http.ResponseWriter, req *http.Request) {
		token := oauth2.GetToken(req)
		fmt.Fprintf(w, "OK: %s", token.Access())
	})
	
	secure := negroni.New()
	secure.Use(oauth2.LoginRequired())
	secure.UseHandler(secureMux)
	
	n := negroni.New()
	n.Use(sessions.Sessions("my_session", cookiestore.New([]byte("secret123"))))
	n.Use(oauth2.Google(&oauth2.Config{
		ClientID:     "client_id",
		ClientSecret: "client_secret",
		RedirectURL:  "refresh_url",
		Scopes:       []string{"https://www.googleapis.com/auth/drive"},
	}))
	
	router := http.NewServeMux()
	
	//routes added to mux do not require authentication
	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		token := oauth2.GetToken(req)
		if token == nil || !token.Valid() {
			fmt.Fprintf(w, "not logged in, or the access token is expired")
			return
		}
		fmt.Fprintf(w, "logged in")
		return
	})
	
	//There is probably a nicer way to handle this than repeat the restricted routes again
	//of course, you could use something like gorilla/mux and define prefix / regex etc.
	router.Handle("/restrict", secure)
	
	n.UseHandler(router)
	srv := &http.Server{
		Handler:      r,
		Addr:         ":8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	
	
	n.Run(":8080")
}*/
