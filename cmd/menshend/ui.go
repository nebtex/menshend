package main

import (
    "github.com/gorilla/mux"
    "fmt"
    "net/http"
    mutils "github.com/nebtex/menshend/pkg/utils"
    mconfig "github.com/nebtex/menshend/pkg/config"
    vault "github.com/hashicorp/vault/api"
    "github.com/ansel1/merry"
    "github.com/markbates/goth/gothic"
    "time"
    "encoding/json"
    "github.com/nebtex/menshend/pkg/apis/menshend/v1"
    "strings"
    "github.com/vulcand/oxy/testutils"
    "github.com/gorilla/csrf"
    "github.com/Sirupsen/logrus"
)

var LoginError = merry.New("login error")

const (
    userPasswordProvider = iota
    tokenProvider
    githubProvider
)

func setToken(token string, provider int, w http.ResponseWriter) {
    //see if token can lookup itself
    vc, err := vault.NewClient(vault.DefaultConfig())
    mutils.HttpCheckPanic(err, LoginError.WithHTTPCode(500).WithUserMessage("Internal error").WithValue("type", provider))
    vc.SetToken(token)
    tokenInfo, err := vc.Auth().Token().LookupSelf()
    mutils.HttpCheckPanic(err, LoginError.WithHTTPCode(401).WithUserMessage("Bad token permission: auth/token/lookup-self missing").WithValue("type", provider))
    //set token in the cookie
    r1, err := tokenInfo.Data["ttl"].(json.Number).Int64()
    mutils.HttpCheckPanic(err, mutils.InternalError)
    ttl := r1 * 1000
    if ttl != 0 {
        ttl += time.Now().UnixNano() / int64(time.Millisecond)
    }
    ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: token, HttpOnly:true }
    if ttl != 0 {
        ct.Expires = time.Unix(ttl / 1000, 0)
    }
    
    ct.Domain = "." + mconfig.Config.HostWithoutPort()
    
    if mconfig.Config.Scheme() == "https" {
        ct.Secure = true
    }
    http.SetCookie(w, ct)
}

//VaultLogin ...
func vaultLogin(c *vault.Client, path string, data map[string]interface{}, provider int) *vault.Secret {
    r := c.NewRequest("POST", "/v1/" + path)
    err := r.SetJSONBody(data)
    mutils.HttpCheckPanic(err, LoginError.WithHTTPCode(500).WithUserMessage("Internal error").WithValue("type", provider))
    
    resp, err := c.RawRequest(r)
    if resp != nil {
        defer resp.Body.Close()
    }
    mutils.HttpCheckPanic(err, LoginError.WithHTTPCode(500).WithUserMessage("Internal error").WithValue("type", provider))
    
    if resp != nil && resp.StatusCode == 404 {
        panic(LoginError.WithHTTPCode(403).WithUserMessage("Bad username or password").WithValue("type", provider))
    }
    
    secret, err := vault.ParseSecret(resp.Body)
    mutils.HttpCheckPanic(err, LoginError.WithHTTPCode(500).WithUserMessage("Internal error").WithValue("type", provider))
    return secret
}

func getTokenFromGithub(gt string) string {
    vc, err := vault.NewClient(vault.DefaultConfig())
    mutils.HttpCheckPanic(err, LoginError.WithHTTPCode(500).WithUserMessage("Internal error"))
    data := map[string]interface{}{}
    data["token"] = gt
    path := "auth/github/login"
    secret := vaultLogin(vc, path, data, githubProvider)
    return secret.Auth.ClientToken
}

func uiLoginHandler() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var token string
        var provider int
        err := r.ParseForm()
        mutils.HttpCheckPanic(err, mutils.InternalError)
        vc, err := vault.NewClient(vault.DefaultConfig())
        mutils.HttpCheckPanic(err, LoginError.WithHTTPCode(500).WithUserMessage("Internal error"))
        switch r.PostForm.Get("method") {
        case "userpass", "ldap", "radius":
            data := map[string]interface{}{}
            data["password"] = r.PostForm.Get("password")
            path := fmt.Sprintf("auth/%s/login/%s", r.PostForm.Get("method"), r.PostForm.Get("user"))
            secret := vaultLogin(vc, path, data, userPasswordProvider)
            token = secret.Auth.ClientToken
            provider = userPasswordProvider
        case "token":
            token = r.PostForm.Get("token")
            provider = tokenProvider
        default:
            panic(LoginError.WithHTTPCode(403).WithUserMessage("Bad username or password"))
        }
        setToken(token, provider, w)
        vars := mux.Vars(r)
        subdomain := vars["subdomain"]
        if subdomain != "" {
            if !strings.HasSuffix(subdomain, ".") {
                subdomain += "."
            }
            v1.ValidateSubdomain(subdomain)
        }
        if subdomain == "" {
            http.Redirect(w, r, mconfig.Config.GetServicePath(), http.StatusTemporaryRedirect)
        }
        http.Redirect(w, r, mconfig.Config.GetSubdomainFullUrl(subdomain), http.StatusTemporaryRedirect)
    })
}

func providerRedirect() http.Handler {
    return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
        vars := mux.Vars(req)
        subdomain := vars["subdomain"]
        provider := vars["provider"]
        // try to get the user without re-authenticating
        if gothUser, err := gothic.CompleteUserAuth(res, req); err == nil {
            setToken(getTokenFromGithub(gothUser.AccessToken), githubProvider, res)
            if subdomain != "" {
                if !strings.HasSuffix(subdomain, ".") {
                    subdomain += "."
                }
                v1.ValidateSubdomain(subdomain)
            }
            if subdomain == "" {
                http.Redirect(res, req, mconfig.Config.GetServicePath(), http.StatusTemporaryRedirect)
            }
            http.Redirect(res, req, mconfig.Config.GetSubdomainFullUrl(subdomain), http.StatusTemporaryRedirect)
        } else {
            url, err := gothic.GetAuthURL(res, req)
            mutils.HttpCheckPanic(err, LoginError.WithHTTPCode(404).WithUserMessage("Bad request, github is not enabled").WithValue("type", githubProvider))
            redirect := mconfig.Config.GetCallbackUrl(provider, subdomain)
            URL := testutils.ParseURI(url)
            if subdomain != "" {
                values := URL.Query()
                values.Set("redirect_uri", redirect)
                URL.RawQuery = values.Encode()
                
            }
            http.Redirect(res, req, URL.String(), http.StatusTemporaryRedirect)
        }
    })
}

func providerCallback() http.Handler {
    return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
        vars := mux.Vars(req)
        subdomain := vars["subdomain"]
        user, err := gothic.CompleteUserAuth(res, req)
        mutils.HttpCheckPanic(err, LoginError.WithHTTPCode(500).WithUserMessage("Internal error").WithValue("type", githubProvider))
        setToken(getTokenFromGithub(user.AccessToken), githubProvider, res)
        if subdomain != "" {
            if !strings.HasSuffix(subdomain, ".") {
                subdomain += "."
            }
            v1.ValidateSubdomain(subdomain)
        }
        if subdomain == "" {
            http.Redirect(res, req, mconfig.Config.GetServicePath(), http.StatusTemporaryRedirect)
        }
        http.Redirect(res, req, mconfig.Config.GetSubdomainFullUrl(subdomain), http.StatusTemporaryRedirect)
    })
}

//SameOriginHandler ...
func SameOriginHandler(next http.Handler, subdomain string) http.Handler {
    return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
        if req.Method == "GET" {
            next.ServeHTTP(res, req)
            return
        }
        
        checked := false
        if req.Header.Get("Referer") != "" {
            if !strings.HasPrefix(req.Header.Get("Referer"), mconfig.Config.GetSubdomainFullUrl(subdomain)) {
                logrus.Infoln("refer", req.Header.Get("Referer"))
                logrus.Infoln("fullUrl", mconfig.Config.GetSubdomainFullUrl(subdomain))
                panic(merry.New("Same origin policy violated").WithUserMessage("Internal error").WithHTTPCode(500))
            }
            checked = true
        }
        if req.Header.Get("Origin") != "" {
            if !strings.HasPrefix(req.Header.Get("Origin"), mconfig.Config.GetSubdomainFullUrl(subdomain)) {
                logrus.Infoln("origin", req.Header.Get("Origin"))
                logrus.Infoln("fullUrl", mconfig.Config.GetSubdomainFullUrl(subdomain))
                panic(merry.New("Same origin policy violated").WithUserMessage("Internal error").WithHTTPCode(500))
            }
            checked = true
        }
        
        if checked == false {
            logrus.Infoln("not referer or origin header found")
            panic(merry.New("Same origin policy violated").WithUserMessage("Internal error").WithHTTPCode(500))
        }
        next.ServeHTTP(res, req)
    })
}


//UiPanicHandler ..
func UiPanicHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
        providerMap := map[int]string{githubProvider:"Github",
            userPasswordProvider: "Username/Password",
            tokenProvider: "Token",
        }
        
        defer func() {
            rec := recover()
            if rec == nil {
                return
            }
            logrus.Errorln(rec)
            errorMessage := "Internal server error"
            provider := -1
            switch x := rec.(type) {
            case merry.Error:
                logrus.Errorln(merry.Details(x))
                errorMessage = merry.UserMessage(x)
                p := merry.Value(x, "type")
                if p != nil {
                    provider = p.(int)
                }
            }
            // Get a session.
            session, err := mconfig.FlashStore.Get(req, "flashes")
            if err != nil {
                http.Error(res, err.Error(), http.StatusInternalServerError)
                return
            }
            // Set a new flash.
            session.AddFlash(errorMessage)
            session.Save(req, res)
            query := ""
            if val, ok := providerMap[provider]; ok  {
                query = "?loginError=" + val
            }
            
            http.Redirect(res, req, mconfig.Config.GetLoginPath() + query, http.StatusTemporaryRedirect)
        }()
        next.ServeHTTP(res, req)
    })
}

func uiLogoutHandler()http.Handler{
    return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
    
        ct := &http.Cookie{Path: "/", Name: "X-Vault-Token", Value: "", HttpOnly:true }
    
        ct.Domain = "." + mconfig.Config.HostWithoutPort()
    
        if mconfig.Config.Scheme() == "https" {
            ct.Secure = true
        }
    
        http.SetCookie(res, ct)
        http.Redirect(res, req, mconfig.Config.GetLoginPath(), http.StatusTemporaryRedirect)
    
    })
}


func loginRouters() http.Handler {
    router := mux.NewRouter()
    router.Handle("/ui/logout", uiLogoutHandler())
    router.Handle("/ui/login", uiLoginHandler()).Methods("POST")
    router.Handle("/ui/login/{subdomain}", uiLoginHandler()).Methods("POST")
    router.Handle("/ui/auth/{provider}/callback/{subdomain}", providerCallback()).Methods("GET")
    router.Handle("/ui/auth/{provider}/callback", providerCallback()).Methods("GET")
    router.Handle("/ui/auth/{provider}", providerRedirect()).Methods("GET")
    router.Handle("/ui/auth/{provider}/{subdomain}", providerRedirect()).Methods("GET")
    return router
}

func getUiCSRF() func(http.Handler )http.Handler  {
    if mconfig.Config.Scheme()=="http"{
        return csrf.Protect([]byte(mconfig.Config.BlockKey), csrf.Secure(false), csrf.Path("/"), csrf.Domain(mconfig.Config.Uris.MenshendSubdomain + mconfig.Config.HostWithoutPort()))
    }
    return csrf.Protect([]byte(mconfig.Config.BlockKey), csrf.Path("/"), csrf.Domain(mconfig.Config.Uris.MenshendSubdomain + mconfig.Config.HostWithoutPort()))
}

func uiHandler() http.Handler {
    CSRF := getUiCSRF()
    return UiPanicHandler(SameOriginHandler(CSRF(loginRouters()), mconfig.Config.Uris.MenshendSubdomain))
}
