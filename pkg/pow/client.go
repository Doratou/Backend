package pow

import (
	"crypto/sha256"
	"fmt"
	"math/big"
)

// Client provides a pure-Go helper for solving PoW challenges.
type Client struct{}

func NewPOWClient() *Client { return &Client{} }

// SolveChallenge finds the smallest non-negative integer nonce whose
// SHA-256 hash of "<salt>:<timestamp>:<nonce>" is strictly less than
// the challenge Target (which encodes the required difficulty).
func SolveChallenge(c Challenge) string {
	targetInt, _ := new(big.Int).SetString(c.Target, 16)

	for nonce := 0; ; nonce++ {
		nonceStr := fmt.Sprintf("%d", nonce)
		rawInput := buildKey(c.Salt, c.Timestamp, nonceStr)

		hashBytes := sha256.Sum256([]byte(rawInput))
		hashInt := new(big.Int).SetBytes(hashBytes[:])

		// Hash < Target --> valid solution found.
		if hashInt.Cmp(targetInt) < 0 {
			return nonceStr
		}
	}
}
