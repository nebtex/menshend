package main

import (
    "strings"
    "time"
    "fmt"
    "net"
)

type pfFlags struct {
    verbose   bool
    keepAlive time.Duration
    token     string
    server    string
    port      string
}

func portForward(flags *pfFlags, connect func(v bool, k time.Duration, s string, token string, r ...string) error) error {
    
    if flags.server == "" || flags.port == "" {
        return fmt.Errorf("%v", "A server and one port is required")
    }
    menshendRemote := strings.Split(flags.port, ":")
    var chiselRemote string
    if len(menshendRemote) == 1 {
        chiselRemote = "default.menshend:" + menshendRemote[0]
    } else if len(menshendRemote) == 2 {
        ip:=net.ParseIP(menshendRemote[0])
        if ip==nil{
            return fmt.Errorf("%v", "Unsoported port format, example 192.168.0.5:3000, 3000")
        }
        chiselRemote = menshendRemote[0] + ":default.menshend:" + menshendRemote[1]
    } else {
        return fmt.Errorf("%v", "Unsoported port format,  example 192.168.0.5:3000, 3000")
    }
    remotes := []string{chiselRemote}
    return connect(flags.verbose, flags.keepAlive, flags.server, flags.token, remotes...)
}
