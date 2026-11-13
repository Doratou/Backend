package pow

import (
	"fmt"
	"math/big"
)

// Format: "<salt>:<timestamp>:<nonce>"
func buildKey(salt string, timestamp int64, nonce string) string {
	return fmt.Sprintf("%s:%d:%s", salt, timestamp, nonce)
}

// Formula: target = 2^(256 - difficulty)
func computeTarget(difficulty int) string {
	target := new(big.Int).Lsh(big.NewInt(1), uint(256-difficulty))
	return fmt.Sprintf("%064x", target)
}
