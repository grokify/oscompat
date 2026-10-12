# tsync: comparing timestamps

```go
import "github.com/grokify/oscompat/tsync"
```

Filesystems keep modification times at different precision: NTFS about 100 ns
(sometimes 2 s), ext4 and APFS nanoseconds, FAT32 2 s, network drives anything.
Comparing times exactly across them reports files as changed when they are not.

```go
if tsync.Equal(a.ModTime(), b.ModTime()) { /* same, within a second */ }
if tsync.After(a.ModTime(), b.ModTime()) { /* definitely newer */ }

tsync.EqualWithTolerance(t1, t2, tsync.FAT32Tolerance)
newest := tsync.Newer(t1, t2)
stored := tsync.TruncateToSecond(time.Now())
```

| Constant | Value | For |
|----------|-------|-----|
| `DefaultTolerance` | 1 s | Mixed platforms and network drives |
| `FAT32Tolerance` | 2 s | FAT32 |
| `HighPrecisionTolerance` | 100 ms | NTFS, ext4, APFS |

`Compare` returns -1, 0 or 1 with the tolerance applied, and the `Before`,
`After`, `Equal` and `Compare` families have a `WithTolerance` form.
`Truncate(t, precision)` normalizes a time before storing or comparing it.
