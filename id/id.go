// Package id provides cross-platform unique identifier generation.
//
// This package exists because time-based ID generation is unreliable on Windows
// due to the system clock's coarse resolution (~15.6ms). Using crypto/rand
// ensures reliable unique ID generation across all platforms.
package id

import (
	"crypto/rand"
	"encoding/hex"
)

// Generate returns a cryptographically random ID encoded as a hex string.
// The byteLen parameter specifies the number of random bytes to generate;
// the resulting string will be twice this length (2 hex chars per byte).
//
// This function panics if crypto/rand fails, which should never happen
// on a properly functioning system.
//
// Example:
//
//	id.Generate(8)  // returns 16-character hex string like "a1b2c3d4e5f67890"
//	id.Generate(16) // returns 32-character hex string
func Generate(byteLen int) string {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		panic("oscompat/id: crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// Generate16 returns a 16-character hex string (8 random bytes).
// This is a convenience function for the common case of generating
// short unique identifiers suitable for span IDs, request IDs, etc.
func Generate16() string {
	return Generate(8)
}

// Generate32 returns a 32-character hex string (16 random bytes).
// This is a convenience function for generating longer identifiers
// suitable for trace IDs or other cases requiring more uniqueness.
func Generate32() string {
	return Generate(16)
}
