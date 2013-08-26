package sshgate

import (
    "code.google.com/p/jeansebtr-crypto-ssh/ssh"
    "encoding/binary"
    "github.com/kballard/go-shellquote"
    "io"
    "log"
)

type chanIn struct {
    c   ssh.Channel
    e   chan ssh.ChannelRequest
}

func (r chanIn) Read(p []byte) (n int, err error) {
    n, err = r.c.Read(p)
    if req, ok := err.(ssh.ChannelRequest); ok {
        r.e <- req
        err = nil
    } else if err == io.EOF {
        close(r.e)
    }
    return n, err
}

func (r chanIn) ReadByte() (byte, error) {
    data := make([]byte, 1)
    _, err := r.Read(data)
    return data[0], err
}

func setEnviron(m map[string]string, buf []byte) {
    var i, l uint32
    i, l = uint32(0), uint32(len(buf))
    for i+4 < l {
        length := binary.BigEndian.Uint32(buf[i : i+4])
        i += 4
        if i+length+4 > l {
            return
        }
        name := string(buf[i : i+length])
        i += length
        length = binary.BigEndian.Uint32(buf[i : i+4])
        i += 4
        if i+length > l {
            return
        }
        m[name] = string(buf[i : i+length])
    }
}

func handleSessionChannel(app Executable, channel ssh.Channel) {
    delegateReads := false
    env := make(map[string]string)
    var complete chan bool
    var ev chan ssh.ChannelRequest
    defer func() {
        if complete != nil {
            <-complete
        }
        channel.Close()
    }()
    for {
        if !delegateReads {
            var data []byte
            // don't actualy read here
            _, err := channel.Read(data)
            if req, ok := err.(ssh.ChannelRequest); ok {
                if req.Request == "env" {
                    setEnviron(env, req.Payload)
                } else if req.Request == "exec" {
                    l := binary.BigEndian.Uint32(req.Payload[:4])
                    args, err := shellquote.Split(string(req.Payload[4 : l+4]))
                    if err != nil {
                        channel.AckRequest(false)
                    } else if app.CanExec(args[0], args[1:], env) {
                        channel.AckRequest(true)
                        delegateReads = true
                        ev = make(chan ssh.ChannelRequest)
                        complete = make(chan bool)
                        go handleExec(app, args, env, channel, ev, complete)
                    } else {
                        channel.AckRequest(false)
                    }
                } else {
                    log.Printf("Unknown ChannelRequest “%s” <%#v>\n", req.Request, req.Payload)
                    if req.WantReply {
                        channel.AckRequest(false)
                    }
                }
            } else if err != nil {
                if err != io.EOF {
                    log.Printf("Error reading ssh.Channel: %#v\n", err)
                }
                break
            }
        } else {
            req, ok := <-ev
            if !ok {
                break
            }
            log.Println("Received subsequent ChannelRequest:", req.Request, req.Payload, req.WantReply)
            if req.WantReply {
                channel.AckRequest(false)
            }
        }
    }
}

func handleExec(app Executable, args []string, env map[string]string, ch ssh.Channel, ev chan ssh.ChannelRequest, end chan bool) {
    defer func() {
        end <- true
    }()
    in := chanIn{c: ch, e: ev}
    // the go ssh library doesn't support sending status code yet...
    // TODO: patch the lib and send ExitCode
    status := app.Exec(args[0], args[1:], env, in, ch, ch.Stderr())
    ch.SendExitStatus(status)
}
