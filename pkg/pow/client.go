package pow

import (
	"crypto/sha256"
	"fmt"
	"math/big"
)

// Client provides a pure-Go helper for solving PoW challenges.
// It is intended for testing and CLI tooling; production clients
// will typically solve challenges in a browser or mobile app.
type Client struct{}

func NewPOWClient() *Client { return &Client{} }

// SolveChallenge finds the smallest non-negative integer nonce whose
// SHA-256 hash of "<salt>:<timestamp>:<nonce>" is strictly less than
// the challenge Target (which encodes the required difficulty).
//
// This is a CPU-bound brute-force search; it blocks until a solution
// is found.  For high difficulty values, run it in a goroutine.
func SolveChallenge(c Challenge) string {
	targetInt, _ := new(big.Int).SetString(c.Target, 16)

	for nonce := 0; ; nonce++ {
		nonceStr := fmt.Sprintf("%d", nonce)
		rawInput := buildKey(c.Salt, c.Timestamp, nonceStr)

		hashBytes := sha256.Sum256([]byte(rawInput))
		hashInt := new(big.Int).SetBytes(hashBytes[:])

		// Hash < Target → valid solution found.
		if hashInt.Cmp(targetInt) < 0 {
			return nonceStr
		}
	}
}
