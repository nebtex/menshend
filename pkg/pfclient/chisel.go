package pfclient

import (
    "time"
    "github.com/Sirupsen/logrus"
    "github.com/nebtex/menshend/pkg/pfclient/chisel/client"
)

func PFConnect(verbose bool, keepalive time.Duration, server string, remotes ...string) {
    
    c, err := chisel.NewClient(&chisel.Config{
        KeepAlive:   keepalive,
        Server:      server,
        Remotes:    remotes,
    
    })
    
    if err != nil {
        logrus.Fatal(err)
    }
    
    c.Info = true
    
    c.Debug = verbose
    
    if err = c.Run(); err != nil {
        logrus.Fatal(err)
    }
}
