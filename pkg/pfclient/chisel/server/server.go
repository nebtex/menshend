package chserver

import (
    "log"
    "net/http"
    "time"
    "github.com/nebtex/menshend/pkg/pfclient/chisel/share"
    "golang.org/x/crypto/ssh"
    "golang.org/x/net/websocket"
    "io"
    "net"
)

type Config struct {
    KeySeed string
    Remote  string
}

type Server struct {
    *chshare.Logger
    Remote      string
    fingerprint string
    wsCount     int
    wsServer    websocket.Server
    httpServer  *chshare.HTTPServer
    sshConfig   *ssh.ServerConfig
}

func NewServer(config *Config) (*Server, error) {
    s := &Server{
        Logger:     chshare.NewLogger("server"),
        wsServer:   websocket.Server{},
        httpServer: chshare.NewHTTPServer(),
    }
    s.wsServer.Handler = websocket.Handler(s.handleWS)
    
    
    //generate private key (optionally using seed)
    key, _ := chshare.GenerateKey(config.KeySeed)
    //convert into ssh.PrivateKey
    private, err := ssh.ParsePrivateKey(key)
    if err != nil {
        log.Fatal("Failed to parse key")
    }
    //fingerprint this key
    s.fingerprint = chshare.FingerprintKey(private.PublicKey())
    //create ssh config
    s.sshConfig = &ssh.ServerConfig{
        ServerVersion:    chshare.ProtocolVersion + "-server",
        PasswordCallback: s.authUser,
        
    }
    s.sshConfig.AddHostKey(private)
    s.Remote = config.Remote
    return s, nil
}

func (s *Server) Close() error {
    //this should cause an error in the open websockets
    return s.httpServer.Close()
}

func (s *Server) HandleHTTP(w http.ResponseWriter, r *http.Request) {
    // websockets upgrade AND has chisel prefix
    if r.Header.Get("Upgrade") == "websocket" &&
        r.Header.Get("Sec-WebSocket-Protocol") == chshare.ProtocolVersion {
        s.wsServer.ServeHTTP(w, r)
        return
    }
    //missing :O
    w.WriteHeader(404)
}
func (s *Server) authUser(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
    return nil, nil
}

func (s *Server) handleWS(ws *websocket.Conn) {
    // Before use, a handshake must be performed on the incoming net.Conn.
    sshConn, chans, reqs, err := ssh.NewServerConn(ws, s.sshConfig)
    if err != nil {
        s.Debugf("Failed to handshake (%s)", err)
        return
    }
    
    //verify configuration
    s.Debugf("Verifying configuration")
    
    //wait for request, with timeout
    var r *ssh.Request
    select {
    case r = <-reqs:
    case <-time.After(10 * time.Second):
        sshConn.Close()
        return
    }
    
    failed := func(err error) {
        r.Reply(false, []byte(err.Error()))
    }
    
    if r.Type != "config" {
        failed(s.Errorf("expecting config request"))
        return
    }
    _, err = chshare.DecodeConfig(r.Payload)
    if err != nil {
        failed(s.Errorf("invalid config"))
        return
    }
    
    //success!
    r.Reply(true, nil)
    
    //prepare connection logger
    s.wsCount++
    id := s.wsCount
    l := s.Fork("session#%d", id)
    
    l.Debugf("Open")
    
    go func() {
        for r := range reqs {
            switch r.Type {
            case "ping":
                r.Reply(true, nil)
            default:
                l.Debugf("Unknown request: %s", r.Type)
            }
        }
    }()
    
    go ConnectStreams(l, chans, s.Remote)
    sshConn.Wait()
    l.Debugf("Close")
}

func ConnectStreams(l *chshare.Logger, chans <-chan ssh.NewChannel, remote string) {
    
    var streamCount int
    
    for ch := range chans {
        
        // string(ch.ExtraData())
        
        stream, reqs, err := ch.Accept()
        if err != nil {
            l.Debugf("Failed to accept stream: %s", err)
            continue
        }
        
        streamCount++
        id := streamCount
        
        go ssh.DiscardRequests(reqs)
        go handleStream(l.Fork("stream#%d", id), stream, remote)
    }
}

func handleStream(l *chshare.Logger, src io.ReadWriteCloser, remote string) {
    
    dst, err := net.Dial("tcp", remote)
    if err != nil {
        l.Debugf("%s", err)
        src.Close()
        return
    }
    
    l.Debugf("Open")
    s, r := chshare.Pipe(src, dst)
    l.Debugf("Close (sent %d received %d)", s, r)
}
