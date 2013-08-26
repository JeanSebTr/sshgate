package sshgate

import (
    "io"
    "log"
)

type App interface {
    // allow the app to make some cleanup
    // disconnect from database, close files...
    Terminate()
}

type Authenticate func(c Connection, user, algo string, pubkey []byte) (bool, App)

type Executable interface {
    App
    // check if the command can be executed
    CanExec(cmd string, args []string, env map[string]string) bool
    // execute the command
    Exec(cmd string, args []string, env map[string]string, stdin io.Reader, stdout, stderr io.Writer) int
}

// all apps must include sshgate.BaseApp
// is a trick for gracefull upgrade of the interface without braking compatibility
// useful because go doesn't have version management of it's dependencies
type BaseApp struct {
}

func (a BaseApp) Terminate() {
    log.Println("For sshgate v0.2.x it will be mandatary to implement App.Terminate. Please upgrade your code.")
}

func MapToEnviron(m map[string]string) (env []string) {
    for k, v := range m {
        env = append(env, k+"="+v)
    }
    return env
}
