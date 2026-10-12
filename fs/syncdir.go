package fs

// SyncDir flushes a directory so that a rename or a file creation within it is
// durable across a crash.
//
// On Windows, and on platforms where a directory cannot be opened for flushing,
// it does nothing and returns nil: there a rename is atomic, and the operating
// system gives no way to ask for more. Callers get the strongest guarantee the
// platform offers and need no platform check.
func SyncDir(dir string) error { return syncDir(dir) }
