package chisel

/*
Notice: this file has been taken and adapted from https://github.com/jpillora/chisel/blob/master/client/proxy.go

MIT License

Copyright © 2015 Jaime Pillora <dev@jpillora.com>

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the 'Software'), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

import (
    "io"
    "net"
    "github.com/jpillora/chisel/share"
)

type Proxy struct {
    *chshare.Logger
    client *Client
    id     int
    count  int
    remote *chshare.Remote
}

func NewProxy(c *Client, id int, remote *chshare.Remote) *Proxy {
    return &Proxy{
        Logger: c.Logger.Fork("%s:%s#%d", remote.RemoteHost, remote.RemotePort, id + 1),
        client: c,
        id:     id,
        remote: remote,
    }
}

func (p *Proxy) start() {
    
    l, err := net.Listen("tcp4", p.remote.LocalHost + ":" + p.remote.LocalPort)
    if err != nil {
        p.Infof("%s", err)
        return
    }
    
    p.Debugf("Enabled")
    for {
        src, err := l.Accept()
        if err != nil {
            p.Infof("Accept error: %s", err)
            return
        }
        go p.accept(src)
    }
}

func (p *Proxy) accept(src io.ReadWriteCloser) {
    p.count++
    cid := p.count
    l := p.Fork("conn#%d", cid)
    
    l.Debugf("Open")
    
    if p.client.sshConn == nil {
        l.Debugf("No server connection")
        src.Close()
        return
    }
    
    remoteAddr := p.remote.RemoteHost + ":" + p.remote.RemotePort
    dst, err := chshare.OpenStream(p.client.sshConn, remoteAddr)
    if err != nil {
        l.Infof("Stream error: %s", err)
        src.Close()
        return
    }
    
    //then pipe
    s, r := chshare.Pipe(src, dst)
    l.Debugf("Close (sent %d received %d)", s, r)
}
