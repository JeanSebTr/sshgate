package sshgate

import (
    "io"
)

type App interface {
}

type Authenticate func(c Connection, user, algo string, pubkey []byte) (bool, App)

type Executable interface {
    App
    CanExec(cmd string, args []string, env map[string]string) bool
    Exec(cmd string, args []string, env map[string]string, stdin io.Reader, stdout, stderr io.Writer) int
}

type BaseApp struct {
}

func MapToEnviron(m map[string]string) (env []string) {
    for k, v := range m {
        env = append(env, k+"="+v)
    }
    return env
}
