package main

import (
    mclient "github.com/nebtex/menshend/pkg/apis/client"
    mutils "github.com/nebtex/menshend/pkg/utils"
    "encoding/json"
    "fmt"
    "github.com/ghodss/yaml"
    "github.com/urfave/cli"
    vault "github.com/hashicorp/vault/api"
    "os"
    "net/url"
    "io/ioutil"
    "net/http"
    "strings"
)

type APICall struct {
    API  string `json:"api,omitempty"`
    Kind string `json:"kind,omitempty"`
    Spec *mclient.AdminService `json:"spec,omitempty"`
}

type APIRequest struct {
    APICall      *APICall
    VaultToken   string
    OutputFormat string
}

type badResponse struct {
    Message string `json:"message,omitempty"`
}
type APIFlags struct {
    role      string
    subdomain string
    token     string
    filename  string
    output    string
    method    string
    api       string
}

func apiClientGetFlags() []cli.Flag {
    return []cli.Flag{
        cli.StringFlag{
            Name: "role, r",
            Value: "",
            Usage: "role/namespace/group of services",
            EnvVar: "MD_ROLE",
        },
        cli.StringFlag{
            Name: "subdomain, s",
            Value: "",
            Usage: "service subdomain",
        },
        cli.StringFlag{
            Name: "token, t",
            Value: "",
            Usage: "vault token",
            EnvVar: vault.EnvVaultToken,
        },
        cli.StringFlag{
            Name: "filename, f",
            Value: "",
            Usage: "Filename, or URL to files that contains the configuration to apply",
        },
        cli.StringFlag{
            Name: "output, o",
            Value: "",
            Usage: "output format json or yaml",
            EnvVar: "MD_OUTPUT",
        },
        cli.StringFlag{
            Name: "api, a",
            Value: "",
            Usage: "baseurl of the menshend api",
            EnvVar: "MD_ADDRESS",
        },
    }
}

func adminCMDHandler(method string) func(c *cli.Context) error {
    return func(c *cli.Context) error {
        flags := &APIFlags{
            role: c.String("role"),
            subdomain:  c.String("subdomain"),
            token:  c.String("token"),
            filename:  c.String("filename"),
            output:  c.String("output"),
            method: method,
            api:c.String("api"),
        }
        response, ok := cmdAPIHandler(flags)
        if ok {
            fmt.Println(printToCmd(response, flags.output))
            return nil
        }
        return fmt.Errorf("%v", printToCmd(response, flags.output))
    }
    
}

func printToCmd(object interface{}, outputFormat string) string {
    val, ok := object.(error)
    if ok {
        return val.Error()
    }
    switch outputFormat {
    case "json":
        data, err := json.Marshal(object)
        mutils.CheckPanic(err)
        return string(data)
    default:
        data, err := yaml.Marshal(object)
        mutils.CheckPanic(err)
        return string(data)
    }
    return ""
}

func cmdAPIHandler(flags *APIFlags) (response interface{}, ok bool) {
    
    //check if filename is a file
    fileInfo, err := os.Stat(flags.filename)
    if err != nil {
        //check if is a url
        URL, err := url.Parse(flags.filename)
        if err == nil {
            resp, err := http.Get(URL.String())
            if err == nil {
                defer resp.Body.Close()
                payload, err := ioutil.ReadAll(resp.Body)
                if err == nil {
                    return parsePayload(payload, flags)
                }
            }
        }
    } else {
        if !fileInfo.IsDir() {
            //list all the files
            payload, err := ioutil.ReadFile(flags.filename)
            for err == nil {
                return parsePayload(payload, flags)
            }
        }
    }
    if flags.method == "get" {
        if flags.role != "" && flags.subdomain != "" {
            return makeRawRequest(flags, &APICall{Spec: &mclient.AdminService{}})
        } else {
            return "Please define a subdomain and a role", false
        }
    }
    return "Could not read or understand the resource", false
}

func parsePayload(payload []byte, flags *APIFlags) (interface{}, bool) {
    ac := &APICall{}
    err := json.Unmarshal(payload, ac)
    if err != nil {
        err = yaml.Unmarshal(payload, ac)
        if err != nil {
            return "File has not a valid json or yaml content: \n" + err.Error(), false
        }
    }
    
    return makeRawRequest(flags, ac)
}

func makeRawRequest(flags *APIFlags, ac *APICall) (response interface{}, ok bool) {
    if (ac.Spec == nil ) {
        return "please define the spec and/or meta section", false
    }
    if ac.Spec.Meta.Id == "" {
        ps := ac.Spec.Meta.SubDomain
        pr := ac.Spec.Meta.RoleId
        if flags.subdomain != "" {
            ps = flags.subdomain
        }
        if flags.role != "" {
            pr = flags.role
        }
        ac.Spec.Meta.Id = fmt.Sprintf("roles/%s/%s", pr, ps)
    }
    
    if flags.api != "" {
        ac.API = flags.api
    }
    
    if ac.API == "" {
        return fmt.Errorf("%v", "Please define the api endpoint with --api or the environment variable MD_ADDRESS"), false
    }
    if flags.token == "" {
        return fmt.Errorf("%v", "Please define the vault token  with --token or the environment variable VAULT_TOKEN"), false
    }
    
    ar := &APIRequest{APICall:ac, VaultToken:flags.token, OutputFormat:flags.output}
    
    switch flags.method{
    case "upsert":
        return Upsert(ar)
    case "get":
        return Get(ar)
    case "delete":
        return Delete(ar)
    default:
        return ":O, why am I here?", false
    }
}

func getAdminApi(ar *APIRequest) *mclient.AdminApi {
    adminApi := mclient.NewAdminApi()
    adminApi.Configuration.BasePath = strings.TrimSuffix(ar.APICall.API, "/v1")
    adminApi.Configuration.DefaultHeader["X-Vault-Token"] = ar.VaultToken
    return adminApi
}

func Upsert(ar *APIRequest) (response interface{}, ok bool) {
    adminApi := getAdminApi(ar)
    as, res, err := adminApi.AdminSaveService(ar.APICall.Spec.Meta.Id, *ar.APICall.Spec)
    if err != nil {
        return err, false
    }
    ok = res.StatusCode == 200
    if ok {
        return fmt.Sprintf("Service %s created/updated --> %s", as.Meta.Id, as.FullUrl), ok
    }
    br := &badResponse{}
    err = json.Unmarshal(res.Payload, br)
    mutils.CheckPanic(err)
    return br, ok
}

func Get(ar *APIRequest) (response interface{}, ok bool) {
    adminApi := getAdminApi(ar)
    as, res, err := adminApi.AdminGetService(ar.APICall.Spec.Meta.Id)
    if err != nil {
        return err, false
    }
    ok = res.StatusCode == 200
    if ok {
        return as, ok
    }
    br := &badResponse{}
    err = json.Unmarshal(res.Payload, br)
    mutils.CheckPanic(err)
    return br, ok
}

func Delete(ar *APIRequest) (response interface{}, ok bool) {
    adminApi := getAdminApi(ar)
    _, res, err := adminApi.AdminDeleteService(ar.APICall.Spec.Meta.Id)
    if err != nil {
        return err, false
    }
    ok = res.StatusCode == 200
    if ok {
        return "Service successfully deleted: " + ar.APICall.Spec.Meta.Id, ok
    }
    br := &badResponse{}
    err = json.Unmarshal(res.Payload, br)
    mutils.CheckPanic(err)
    return br, ok
}
