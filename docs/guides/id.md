# id: identifiers

```go
import "github.com/grokify/oscompat/id"
```

Identifiers built from the clock collide on Windows, where `time.Now()` advances
in steps of about 15.6 ms and several values taken in a row are equal. `id`
draws from `crypto/rand` instead.

```go
spanID := id.Generate16()  // 16 hex characters (8 random bytes)
traceID := id.Generate32() // 32 hex characters (16 random bytes)
custom := id.Generate(4)   // 8 hex characters (N bytes is 2N characters)
```

`Generate` panics if `crypto/rand` fails, which does not happen on a working
system.

These are random, not ordered. If you need identifiers that sort by creation
time, use a time-ordered scheme such as UUIDv7 and keep the ordering in the
identifier, not in a timestamp column.
