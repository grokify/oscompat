# localnet: local sockets

```go
import "github.com/grokify/oscompat/localnet"
```

Unix uses domain sockets; Windows has them only in limited form. `localnet`
gives both the same calls.

```go
// Server
l, err := localnet.Listen("myapp")
if err != nil {
    return err
}
defer l.Close() // also removes the socket or port file

for {
    conn, err := l.Accept()
    // ...
}

// Client
conn, err := localnet.Dial("myapp")
```

| Platform | Mechanism |
|----------|-----------|
| Unix | A domain socket in `$XDG_RUNTIME_DIR` or `/tmp` |
| Windows | A TCP listener on localhost on an ephemeral port, with the port recorded in a file for clients |

- `Cleanup(name)` removes a stale socket or port file left by a crash. It is safe
  to call when none exists.
- `SocketPath(name)` reports the path or address, for logs.
- An empty name is `ErrInvalidName`; a socket that already exists is
  `ErrSocketExists`.
- Pick a name that is yours. Two applications that choose the same one will
  collide.
