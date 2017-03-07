package main

import (
    "flag"
    "fmt"
    "os"
    "github.com/Sirupsen/logrus"
    "strings"
    "time"
)

func portForward(args []string, connect func(v bool, k time.Duration, s string, r ...string)) {
    
    var clientHelp = `
    this command use the awesome chisel project https://github.com/jpillora/chisel
    
	Usage: menshend port-forward [options] <server> [local-host]:[local-port]:<remote-port>
	
	<server>        is the full http(s) url of the service under the Menshend
	                space wanted to be tunneled, ip addresses are not allowed
	[local-host]    defaults to 0.0.0.0 (all interfaces).
	[local-port]    defaults to remote-port.
	<remote-port>   required.
	
	● Example:
	    
	    tunneling a mongo database locate in the machine learning lab of the example.com company to the localhost
	    
		menshend client https://mongo.ml-lab.example.com 27017          mongo will  be available under localhost:27017
		menshend client	https://mongo.ml-lab.example.com 3000:27017     mongo ... localhost:3000
		menshend client	https://labs.example.com 192.168.0.5:3000:80    mongo ... 192.168.0.5:3000
		menshend client	https://labs.example.com 192.168.0.5:80         mongo ... 192.168.0.5:27017
			
	● Options:
	
	  --keepalive, An optional keepalive interval. Since the underlying
	  transport is HTTP, in many instances we'll be traversing through
	  proxies, often these proxies will close idle connections. You must
	  specify a time with a unit, for example '30s' or '2m'. Defaults
	  to '0s' (disabled).
	  
	  --v, verbose debug
	  
	● Environment variables:
	
	  MENSHEND_TOKEN, menshend auth token`
    
    flags := flag.NewFlagSet("client", flag.ContinueOnError)
    keepalive := flags.Duration("keepalive", 0, "")
    verbose := flags.Bool("v", false, "")
    
    flags.Usage = func() {
        fmt.Fprintf(os.Stderr, clientHelp)
        os.Exit(1)
    }
    
    flags.Parse(args)
    
    //pull out options, put back remaining args
    args = flags.Args()
    if len(args) < 2 {
        logrus.Fatal("A server and least one remote is required")
    }
    server := args[0]
    menshendRemote := strings.Split(args[1], ":")
    var chiselRemote string
    if len(menshendRemote) == 1 {
        chiselRemote = "default.menshend:" + menshendRemote[0]
    } else if len(menshendRemote) == 2 {
        chiselRemote = menshendRemote[0] + ":default.menshend:" + menshendRemote[1]
    } else if len(menshendRemote) == 3 {
        chiselRemote = menshendRemote[0] + ":" + menshendRemote[1] + ":default.menshend:" + menshendRemote[2]
    }
    remotes := []string{chiselRemote}
    connect(*verbose, *keepalive, server, remotes...)
}
