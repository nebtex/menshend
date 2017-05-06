package pfclient

import (
    "time"
    "github.com/nebtex/menshend/pkg/pfclient/chisel/client"
)

func PFConnect(verbose bool, keepalive time.Duration, server string, token string, role string, remotes ...string) error {
    c, err := chisel.NewClient(&chisel.Config{
        KeepAlive:   keepalive,
        Server:      server,
        Remotes:     remotes,
        Token: token,
        Role: role,
    })
    if err != nil {
        return err
    }
    c.Info = true
    c.Debug = verbose
    return c.Run()
}
