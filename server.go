package sshgate

import (
    "code.google.com/p/jeansebtr-crypto-ssh/ssh"
    "log"
    "strconv"
    "sync"
)

type Server interface {
    Listen(addr string, port int) error
}

type connections struct {
    sync.RWMutex
    m   map[string]*tConnection
}

type tServer struct {
    config       *ssh.ServerConfig
    connections  connections
    onConnection Authenticate
}

func NewServer(key []byte, cb Authenticate) (Server, error) {
    server := &tServer{onConnection: cb}
    server.connections.m = make(map[string]*tConnection)
    server.config = &ssh.ServerConfig{
        PublicKeyCallback: func(conn *ssh.ServerConn, user, algo string, pubkey []byte) bool {
            return server.handleAuthKey(conn, user, algo, pubkey)
        },
    }
    if err := server.config.SetRSAPrivateKey(key); err != nil {
        return nil, err
    }
    return server, nil
}

func (s *tServer) Listen(addr string, port int) error {
    if addr == "" {
        addr = "0.0.0.0"
    }
    listener, err := ssh.Listen("tcp", addr+":"+strconv.Itoa(port), s.config)
    if err != nil {
        return err
    }
    for {
        sConn, err := listener.Accept()
        if err != nil {
            log.Printf("Failed to accept incoming connection: %#v\n", err)
        } else {
            conn := tConnection{ServerConn: sConn, server: s}
            s.addConnection(&conn)
            go handleConnection(s, &conn)
        }
    }
}

func (s *tServer) handleAuthKey(sConn *ssh.ServerConn, user, algo string, pubkey []byte) bool {
    conn := s.getConnection(sConn)
    if ok, app := s.onConnection(conn, user, algo, pubkey); ok {
        conn.app = app
        return true
    } else {
        return false
    }
}

func (s *tServer) addConnection(c *tConnection) {
    addr := c.RemoteAddr().String()
    s.connections.Lock()
    s.connections.m[addr] = c
    s.connections.Unlock()
}

func (s tServer) getConnection(c *ssh.ServerConn) *tConnection {
    addr := c.RemoteAddr().String()
    s.connections.RLock()
    defer s.connections.RUnlock()
    return s.connections.m[addr]
}

func (s *tServer) closeConnection(c *tConnection) {
    addr := c.RemoteAddr().String()
    s.connections.Lock()
    delete(s.connections.m, addr)
    s.connections.Unlock()
    c.Close()
    c.app.Terminate()
}
