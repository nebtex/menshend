package menshend


/*
//PanicHandler middleware for recover from panic attacks
func PanicHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            rec := recover()
            if rec != nil {
                logrus.Errorf("%v", rec)
                switch v := rec.(type) {
                case merry.Error:
                    subDomain := getSubDomain(r.URL.Host)
                    if len(subDomain) == 0 {
                        w.WriteHeader(500)
                        response := Response{Success:false,
                            Message:v.Error()}
                        data, err := json.Marshal(response)
                        if (err != nil) {
                            logrus.Error(err)
                        }
                        w.Write(data)
                        return
                        
                    } else {
                        // Get a session.
                        session, err := FlashStore.Get(r, "errors")
                        if err != nil {
                            http.Error(w, err.Error(), http.StatusInternalServerError)
                            logrus.Error(err)
                            return
                        }
                        session.AddFlash(v.Error())
                        session.Save(r, w)
                        redirectUrl := fmt.Sprintf("%s://%s/ui/login", Config.Scheme, Config.Host)
                        http.Redirect(w, r, redirectUrl, 301)
                        return
                    }
                
                default:
                    w.WriteHeader(500)
                    errMsg := `{"success":false, "message": "Unkown error (please contact your admin)"}`
                    w.Write([]byte(errMsg))
                    return
                }
            }
        }()
        next.ServeHTTP(w, r)
    })
}
