package sshgate

import (
    "code.google.com/p/jeansebtr-crypto-ssh/ssh"
    "io"
    "log"
    "net"
)

type Connection interface {
    net.Conn
}

type tConnection struct {
    *ssh.ServerConn
    server *tServer
    app    App
}

func handleConnection(s *tServer, c *tConnection) {
    defer s.closeConnection(c)
    if err := c.Handshake(); err != nil {
        log.Printf("Handshake error: %#v\n", err)
        return
    }

    for {
        channel, err := c.Accept()
        if err != nil {
            if err == io.EOF {
                return
            }
            log.Panicf("Error: %#v\n", err)
            return
        }
        if channel.ChannelType() == "session" {
            if app, ok := c.app.(Executable); ok {
                channel.Accept()
                go handleSessionChannel(app, channel)
            } else {
                channel.Reject(ssh.UnknownChannelType, "no session support")
            }
        } else {
            channel.Reject(ssh.UnknownChannelType, "unknown channel type")
        }
    }
}
