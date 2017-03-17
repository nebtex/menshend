package chisel
/*
Notice: this file has been taken and adapted from https://github.com/jpillora/chisel/blob/master/client/client.go

MIT License

Copyright © 2015 Jaime Pillora <dev@jpillora.com>

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the 'Software'), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

import (
    "fmt"
    "io"
    "net/url"
    "regexp"
    "strings"
    "time"
    "github.com/jpillora/backoff"
    "github.com/jpillora/chisel/share"
    "golang.org/x/crypto/ssh"
    "golang.org/x/net/websocket"
    "os"
    "github.com/Sirupsen/logrus"
)

type Config struct {
    shared      *chshare.Config
    Fingerprint string
    Auth        string
    KeepAlive   time.Duration
    Server      string
    Remotes     []string
}

type Client struct {
    *chshare.Logger
    config    *Config
    sshConfig *ssh.ClientConfig
    proxies   []*Proxy
    sshConn   ssh.Conn
    server    string
    running   bool
    runningc  chan error
}

func Dial(url_, protocol, origin string, port string) (ws *websocket.Conn, err error) {
    config, err := websocket.NewConfig(url_, origin)
    if err != nil {
        return nil, err
    }
    if protocol != "" {
        config.Protocol = []string{protocol}
    }
    
    mt := os.Getenv("MENSHEND_TOKEN")
    if mt == "" {
        logrus.Fatal("Please set the MENSHEND_TOKEN environment variable")
    }
    config.Header.Add("X-Vault-Token", mt)
    config.Header.Add("X-Menshend-Port-Forward", port)
    return websocket.DialConfig(config)
}

func NewClient(config *Config) (*Client, error) {
    
    //apply default scheme
    if !strings.HasPrefix(config.Server, "http") {
        config.Server = "https://" + config.Server
    }
    
    u, err := url.Parse(config.Server)
    if err != nil {
        return nil, err
    }
    
    //apply default port
    if !regexp.MustCompile(`:\d+$`).MatchString(u.Host) {
        if u.Scheme == "https" || u.Scheme == "wss" {
            u.Host = u.Host + ":443"
        } else {
            u.Host = u.Host + ":80"
        }
    }
    
    //swap to websockets scheme
    u.Scheme = strings.Replace(u.Scheme, "http", "ws", 1)
    
    shared := &chshare.Config{}
    
    for _, s := range config.Remotes {
        r, err := chshare.DecodeRemote(s)
        if err != nil {
            return nil, fmt.Errorf("Failed to decode remote '%s': %s", s, err)
        }
        shared.Remotes = append(shared.Remotes, r)
    }
    config.shared = shared
    
    client := &Client{
        Logger:   chshare.NewLogger("client"),
        config:   config,
        server:   u.String(),
        running:  true,
        runningc: make(chan error, 1),
    }
    
    user, pass := chshare.ParseAuth(config.Auth)
    
    client.sshConfig = &ssh.ClientConfig{
        User:            user,
        Auth:            []ssh.AuthMethod{ssh.Password(pass)},
        ClientVersion:   chshare.ProtocolVersion + "-client",
    }
    
    return client, nil
}

//Start then Wait
func (c *Client) Run() error {
    go c.start()
    return c.Wait()
}


//Starts the client
func (c *Client) Start() {
    go c.start()
}

func (c *Client) start() {
    c.Infof("Connecting to %s\n", c.server)
    
    //prepare proxies
    for id, r := range c.config.shared.Remotes {
        proxy := NewProxy(c, id, r)
        go proxy.start()
        c.proxies = append(c.proxies, proxy)
    }
    
    //optional keepalive loop
    if c.config.KeepAlive > 0 {
        go func() {
            for range time.Tick(c.config.KeepAlive) {
                if c.sshConn != nil {
                    c.sshConn.SendRequest("ping", true, nil)
                }
            }
        }()
    }
    
    //connection loop!
    var connerr error
    b := &backoff.Backoff{Max: 5 * time.Minute}
    
    for {
        if !c.running {
            break
        }
        if connerr != nil {
            d := b.Duration()
            c.Infof("Retrying in %s...\n", d)
            connerr = nil
            time.Sleep(d)
        }
        
        ws, err := Dial(c.server, chshare.ProtocolVersion, "http://menshend.io/", c.config.shared.Remotes[0].RemotePort)
        if err != nil {
            connerr = err
            continue
        }
        
        sshConn, chans, reqs, err := ssh.NewClientConn(ws, "", c.sshConfig)
        
        //NOTE: break == dont retry on handshake failures
        if err != nil {
            if strings.Contains(err.Error(), "unable to authenticate") {
                c.Infof("Authentication failed")
                c.Debugf(err.Error())
            } else {
                c.Infof(err.Error())
            }
            break
        }
        conf, _ := chshare.EncodeConfig(c.config.shared)
        c.Debugf("Sending configurating")
        t0 := time.Now()
        _, configerr, err := sshConn.SendRequest("config", true, conf)
        if err != nil {
            c.Infof("Config verification failed")
            break
        }
        if len(configerr) > 0 {
            c.Infof(string(configerr))
            break
        }
        c.Infof("Connected (Latency %s)", time.Now().Sub(t0))
        //connected
        b.Reset()
        
        c.sshConn = sshConn
        go ssh.DiscardRequests(reqs)
        go chshare.RejectStreams(chans) //TODO allow client to ConnectStreams
        err = sshConn.Wait()
        //disconnected
        c.sshConn = nil
        if err != nil && err != io.EOF {
            connerr = err
            c.Infof("Disconnection error: %s", err)
            continue
        }
        c.Infof("Disconnected\n")
    }
    close(c.runningc)
}

//Wait blocks while the client is running
func (c *Client) Wait() error {
    return <-c.runningc
}

//Close manual stops the client
func (c *Client) Close() error {
    c.running = false
    if c.sshConn == nil {
        return nil
    }
    return c.sshConn.Close()
}
