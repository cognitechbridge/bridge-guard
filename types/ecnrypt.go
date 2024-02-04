package types

import "golang.org/x/crypto/chacha20poly1305"

// Key represents a 256-bit key used for ChaCha20-Poly1305.
type Key [chacha20poly1305.KeySize]byte
