package main

import (
    "strings"
    "time"
    "fmt"
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
        chiselRemote = menshendRemote[0] + ":default.menshend:" + menshendRemote[1]
    } else if len(menshendRemote) == 3 {
        chiselRemote = menshendRemote[0] + ":" + menshendRemote[1] + ":default.menshend:" + menshendRemote[2]
    }
    remotes := []string{chiselRemote}
    return connect(flags.verbose, flags.keepAlive, flags.server, flags.token, remotes...)
}
