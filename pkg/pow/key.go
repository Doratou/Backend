package pow

import (
	"fmt"
	"math/big"
)

// buildKey assembles the canonical string that is both hashed for the PoW
// puzzle and embedded in the HMAC-signed challenge.
//
// Format: "<salt>:<timestamp>:<nonce>"
func buildKey(salt string, timestamp int64, nonce string) string {
	return fmt.Sprintf("%s:%d:%s", salt, timestamp, nonce)
}

// computeTarget converts a difficulty (number of required leading zero bits)
// into a 64-character hex target string representing the 256-bit upper bound
// (exclusive) that a valid solution hash must be below.
//
// Formula: target = 2^(256 - difficulty)
//
// Examples:
//
//	difficulty=16 → first 16 bits are zero → target ≈ 0001000...000 (hex)
//	difficulty=24 → first 24 bits are zero → target ≈ 0000010...000 (hex)
func computeTarget(difficulty int) string {
	target := new(big.Int).Lsh(big.NewInt(1), uint(256-difficulty))
	return fmt.Sprintf("%064x", target)
}
