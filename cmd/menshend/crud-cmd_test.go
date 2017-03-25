package main

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
    "net/http"
    "github.com/nebtex/menshend/pkg/apis/menshend/v1"
    "time"
    testutils "github.com/nebtex/menshend/pkg/utils/test"
    mutils "github.com/nebtex/menshend/pkg/utils"
    "os"
    vault "github.com/hashicorp/vault/api"
    "fmt"
)

func TestCli(t *testing.T) {
    mutils.CheckPanic(os.Setenv(vault.EnvVaultAddress, "http://127.0.0.1:8200"))
    go http.ListenAndServe(":31800", v1.APIHandler())
    time.Sleep(1*time.Second)
    testutils.CleanVault()
    
    Convey("create from file", t, func() {
        flag := &APIFlags{}
        flag.api = "http://localhost:31800/v1"
        flag.subdomain = "matlab."
        flag.role = "nosecret"
        flag.filename = "service-example-1.yml"
        flag.method = "upsert"
        flag.token = "myroot"
        _, ok := cmdAPIHandler(flag)
        So(ok, ShouldBeTrue)
        
    })
    Convey("create from url", t, func() {
        flag := &APIFlags{}
        flag.api = "http://localhost:31800/v1"
        flag.filename = "https://gist.githubusercontent.com/criloz/709e727a192bc25acb882a3dd1fc4f8e/raw/05c543e8a57416022fa23cb53270552e40aacaec/menshend-frontend.yml"
        flag.method = "upsert"
        flag.token = "myroot"
        _, ok := cmdAPIHandler(flag)
        So(ok, ShouldBeTrue)
    })
    
    
    Convey("get service", t, func() {
        flag := &APIFlags{}
        flag.api = "http://localhost:31800/v1"
        flag.filename = "https://gist.githubusercontent.com/criloz/709e727a192bc25acb882a3dd1fc4f8e/raw/05c543e8a57416022fa23cb53270552e40aacaec/menshend-frontend.yml"
        flag.method = "get"
        flag.token = "myroot"
        response, ok := cmdAPIHandler(flag)
        fmt.Println(response)
        So(ok, ShouldBeTrue)
    })
    
    Convey("delete from url", t, func() {
        flag := &APIFlags{}
        flag.api = "http://localhost:31800/v1"
        flag.filename = "https://gist.githubusercontent.com/criloz/709e727a192bc25acb882a3dd1fc4f8e/raw/05c543e8a57416022fa23cb53270552e40aacaec/menshend-frontend.yml"
        flag.method = "delete"
        flag.token = "myroot"
        _, ok := cmdAPIHandler(flag)
        So(ok, ShouldBeTrue)
    })
}

