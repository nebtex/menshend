package main

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/markbates/goth/gothic"
	"github.com/nebtex/menshend/pkg/apis/menshend/v1"
	mutils "github.com/nebtex/menshend/pkg/utils"
	"github.com/rakyll/statik/fs"

	"github.com/ansel1/merry"
	"github.com/gorilla/mux"
	vault "github.com/hashicorp/vault/api"
	. "github.com/nebtex/menshend/statik"
    "github.com/hashicorp/vault/http"
)

var LoginError = merry.New("login error")

func mainHandler(response http.ResponseWriter, request *http.Request) {
	//detect menshend host
	//use proxy server

}

func proxyServer() http.Handler {
	return PanicHandler(GetSubDomainHandler(v1.BrowserDetectorHandler(
		NeedLogin(RoleHandler(GetServiceHandler(ProxyHandler()))))))
}

func ui() http.Handler {
	r := mux.NewRouter()
	statikFS, _ := fs.New()
	r.PathPrefix("/static").Handler(http.FileServer(statikFS))
	r.PathPrefix("/").Handler(Index())
	return r
}

func gothLogin(res http.ResponseWriter, req *http.Request) {
	// try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(res, req); err == nil {
		setToken(VaultGithubLogin(gothUser.AccessToken))

	} else {
		gothic.BeginAuthHandler(res, req)
	}
}

func(res http.ResponseWriter, req *http.Request) {

		user, err := gothic.CompleteUserAuth(res, req)
		if err != nil {
			fmt.Fprintln(res, err)
			return
		}
        //redirect
	}

func uiLogin() http.Handler {
	//add csfr protection
	// use flash for error [error hanlder ]
	// redirect to service [cache]
	//check same origin policy
	//read form, if github use gomniauth
	//token use token login
	//vault, rados, ldap
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string
		err := r.ParseForm()
		mutils.HttpCheckPanic(err, mutils.InternalError)
		vc, err := vault.NewClient(vault.DefaultConfig())
		mutils.HttpCheckPanic(err, LoginError.WithHTTPCode(500).WithUserMessage("Internal error"))
		switch r.PostForm.Get("method") {
		case "userpass", "ldap", "radius":
			data := map[string]interface{}{}
			data["password"] = r.PostForm.Get("password")
			path := fmt.Sprintf("auth/%s/login/%s", r.PostForm.Get("method"), r.PostForm.Get("user"))
			secret := VaultLogin(vc, path, data)
			token = secret.Auth.ClientToken
		case "token":
			token = r.PostForm.Get("token")
		case "github":
			//ddsdsds
			return
		default:
			panic(LoginError.WithHTTPCode(403).WithUserMessage("Bad username or password"))
		}

		setToken(token)

	})
}

//VaultLogin ...
func VaultLogin(c *vault.Client, path string, data map[string]interface{}) *vault.Secret {
	r := c.NewRequest("POST", "/v1/"+path)
	err := r.SetJSONBody(data)
	mutils.HttpCheckPanic(err, LoginError.WithHTTPCode(500).WithUserMessage("Internal error").WithValue("type", "user/password"))

	resp, err := c.RawRequest(r)
	if resp != nil {
		defer resp.Body.Close()
	}
	mutils.HttpCheckPanic(err, LoginError.WithHTTPCode(500).WithUserMessage("Internal error").WithValue("type", "user/password"))

	if resp != nil && resp.StatusCode == 404 {
		panic(LoginError.WithHTTPCode(403).WithUserMessage("Bad username or password").WithValue("type", "user/password"))
	}

	secret, err := vault.ParseSecret(resp.Body)
	mutils.HttpCheckPanic(err, LoginError.WithHTTPCode(500).WithUserMessage("Internal error").WithValue("type", "user/password"))
	return secret
}

func menshendServer(address, port string) error {
	// /ui
	// /uilogin
	// /uilogout
	// /v1 - api
	//http.Handle("/api", v1.APIHandler())
	http.Handle("/", ui())

	logrus.Infof("Server listing on %s:%s", address, port)
	return http.ListenAndServe(fmt.Sprintf("%s:%s", address, port), nil)

	//r.Host("{subdomain:[a-z\\-]+}." + Config.HostWithoutPort()).Handler(handler)
}

/*
func setToken(u *User, expiresIn int64, response *restful.Response, hasCSRF bool) {
    expireAt := MakeTimestampMillisecond()
    if expiresIn == 0 {
        expireAt += Config.DefaultTTL
    } else {
        expireAt += expiresIn
    }
    u.SetExpiresAt(expireAt)

    if !hasCSRF {
        response.AddHeader("X-Menshend-Token", u.GenerateJWT())

    } else {
        ct := &http.Cookie{Path: "/", Name: "X-Menshend-Token", Value: u.GenerateJWT(),
            Expires: time.Unix(u.ExpiresAt / 1000, 0),
            HttpOnly:true }

        ct.Domain = "." + Config.HostWithoutPort()

        if Config.Scheme == "https" {
            ct.Secure = true
        }
        http.SetCookie(response.ResponseWriter, ct)

    }
}
func setExpirationTime(u *User, expiresIn int64) *User {
    expireAt := MakeTimestampMillisecond()
    if expiresIn == 0 {
        expireAt += Config.DefaultTTL
    } else {
        expireAt += expiresIn
    }
    u.SetExpiresAt(expireAt)
    return u

}
*/
