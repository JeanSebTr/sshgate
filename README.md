# sshgate

Utility to ease the creation of SSH based applications in go


## Example

See this [Demo][demo] ( [app.go][file] ) of sshgate which accept git push/pull operations.


## How to use?

The idea is to create an [authentication function](https://github.com/xpensia/git-demo/blob/master/app.go#L48) that will receive the user creadentials and return if the connection is allowed and an object that implement one of sshgate interfaces.
`type Authenticate func(c sshgate.Connection, user, algo string, pubkey []byte) (bool, sshgate.App)`


The [Demo][demo] implement `sshgate.Executable` to allow the clients to execute git-upload-pack and git-receive-pack.

- First, sshgate query the app to check if the execution should be allowed with [Executable.CanExec](https://github.com/xpensia/git-demo/blob/master/app.go#L63)
- If the response is yes, the command is executed by [Executable.Exec](https://github.com/xpensia/git-demo/blob/master/app.go#L69)


## TODOs

- Add support for password based authentication
- Add more interfaces to expose more features of SSH


## Caveats

In GO, SSH is a first class citizen with support for the protocol from [go.crypto/ssh](http://godoc.org/code.google.com/p/go.crypto/ssh), but a lot of features are not exposed in the server API.
For example, the package doesn't allow to send program's exit code to the client and this makes git complain that `The remote end hung up unexpectedly`...
There is [an issue about that](https://code.google.com/p/go/issues/detail?id=6235) opened at the same time I was working on the project.

Furthermore if the server hasn't read all the data before io.EOF is received, ssh won't let us read the data.
I've opened [an issue about that](https://code.google.com/p/go/issues/detail?id=6250).

Please star those issues :)

For the meantime, sshgate use [my fork of go.crypto/ssh](https://code.google.com/r/jeansebtr-crypto-ssh/source/list)


[demo]: https://github.com/xpensia/git-demo
[file]: https://github.com/xpensia/git-demo/blob/master/app.go
