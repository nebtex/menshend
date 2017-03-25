package main

import (
    "os"
    "github.com/urfave/cli"
)

func main() {
    app := cli.NewApp()
    
    app.Commands = []cli.Command{
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
    }
    
    app.Run(os.Args)
}
