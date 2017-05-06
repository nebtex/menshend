package main

import (
    "os"
    "github.com/urfave/cli"
    vault "github.com/hashicorp/vault/api"
    "github.com/nebtex/menshend/pkg/pfclient"
    mconfig "github.com/nebtex/menshend/pkg/config"
    "fmt"
    "net/url"
    "net/http"
)

func main() {
    app := cli.NewApp()
    app.Name = "menshend"
    app.Usage = ""
    app.Version = Version
    app.Authors = []cli.Author{{"nebtex", "publicdev@nebtex.com"}}
    
    app.Commands = []cli.Command{
        {
            Action: func(c *cli.Context) error {
                fl := &pfFlags{}
                fl.token = c.String("token")
                fl.keepAlive = c.Duration("keepalive")
                fl.port = c.String("port")
                fl.verbose = c.Bool("verbose")
                fl.server = c.String("server")
                //calculate role
                URL, err := url.Parse(fl.server)
                
                if err != nil {
                    return err
                }
                query := URL.Query()
                fl.role = query.Get("md-role")
                query.Del("md-role")
                URL.RawQuery = query.Encode()
                fl.server = URL.String()
                if c.String("role") != "" {
                    fl.role = c.String("role")
                }
                if fl.token == "" {
                    return fmt.Errorf("%v", "Please set token flag, or the VAULT_TOKEN environment variable")
                }
                
                if fl.role == "" {
                    return fmt.Errorf("%v", "Please set the role flag, or the MD-ROLE environment variable")
                }
                
                //test that the server is returning a ok response
                nr, err := http.NewRequest("GET", fl.server, nil)
                if err != nil {
                    return err
                }
                nr.Header.Set("X-Vault-Token", fl.token)
                nr.Header.Set("md-role", fl.role)
                
                client := &http.Client{}
                resp, err := client.Do(nr)
                if err != nil {
                    return err
                }
                
                if resp.StatusCode != 200 {
                    return fmt.Errorf("%v", resp.Status)
                }
                
                return portForward(fl, pfclient.PFConnect)
            },
            Name:    "port-forward",
            Flags: []cli.Flag{
                cli.StringFlag{
                    Name: "server, s",
                    Value: "",
                    Usage: "full http(s) url of the service under the Menshend space wanted to be tunneled, ip addresses are not allowed",
                    EnvVar: "PORT_FORWARD_ENDPOINT",
                },
                cli.StringFlag{
                    Name: "port, p",
                    Value: "",
                    Usage: "[local-host]:<local-port>",
                },
                cli.StringFlag{
                    Name: "token, t",
                    Value: "",
                    Usage: "vault token",
                    EnvVar: vault.EnvVaultToken,
                },
                cli.StringFlag{
                    Name: "role, r",
                    Value: "",
                    Usage: "service role",
                    EnvVar: "MD-ROLE",
                },
                cli.DurationFlag{
                    Name: "keepalive, k",
                    Usage: "An optional keepalive interval. Since the underlying \n transport is HTTP, in many instances we'll be traversing through " +
                        "proxies, often these proxies will close idle connections. \n You must" +
                        "specify a time with a unit, for example '30s' or '2m'. Defaults" +
                        "to '0s' (disabled)",
                },
                cli.BoolFlag{
                    Name: "verbose, v",
                    Usage: "verbose debug",
                },
            },
            Usage:  "Create secure tunnels",
            Description:`this command is adapted from the chisel project https://github.com/jpillora/chisel
							 
 ● Examples:
     
 tunneling a mongo database, locate in some of the laboratories of the example.com organization to the localhost
 
 ♦ make mongo available on localhost:27017
     menshend port-forward   --server https://mongo.ml-lab.example.com  --port 27017
 ♦ ... mongo ... localhost:3000
     menshend port-forward	--server https://mongo.ml-lab.example.com  --port 3000
 ♦ ... mongo ... 192.168.0.5:3000
     menshend port-forward	--server https://labs.example.com  --port 192.168.0.5:3000`,
        },
        {
            Name:    "admin",
            Aliases: []string{"adminServices"},
            Usage:   "admin api - add/update/delete services",
            Subcommands: []cli.Command{
                {
                    Name:  "get",
                    Aliases:[]string{"read"},
                    Usage: "return service definition",
                    Flags: apiClientGetFlags(),
                    Action: adminCMDHandler("get"),
                },
                {
                    Name:  "delete",
                    Aliases:[]string{"remove", "del", "eliminate", "erase"},
                    Usage: "delete a service",
                    Flags: apiClientGetFlags(),
                    Action: adminCMDHandler("delete"),
                },
                {
                    Name:  "upsert",
                    Aliases:[]string{"save", "apply", "update", "put", "write", "upload", "add", "replace", "create", "post"},
                    Usage: "create or update a service",
                    Flags: apiClientGetFlags(),
                    Action: adminCMDHandler("upsert"),
                },
            },
        },
        {
            Name:    "server",
            Aliases: []string{"run", "start"},
            Usage:   "run menshend server",
            Flags: []cli.Flag{
                cli.StringFlag{
                    Name: "port, p",
                    Value: "8787",
                    Usage: "bind port",
                },
                cli.StringFlag{
                    Name: "config, c",
                    Value: "",
                    Usage: "config file",
                    EnvVar: "MENSHEND_CONFIG_FILE",
                },
                cli.StringFlag{
                    Name: "address, a",
                    Value: "0.0.0.0",
                    Usage: "bind address",
                },
            },
            Action: func(c *cli.Context) error {
                config := c.String("config")
                mconfig.ConfigFile = &config
                err := mconfig.LoadConfig()
                if err != nil {
                    return err
                }
                return menshendServer(c.String("address"), c.String("port"))
            },
        },
        {
            Name:    "version",
            Aliases: []string{"release"},
            Usage:   "get binary version",
            Action: func(c *cli.Context) error {
                fmt.Println(Version)
                return nil
            },
        },
    }
    
    app.Run(os.Args)
}
