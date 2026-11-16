package pow

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"time"
)

type Server struct {
	maxTTL time.Duration
}

func NewPOWServer(
	maxTTL time.Duration,
) *Server {
	return &Server{
		maxTTL: maxTTL,
	}
}

func (s *Server) GenerateSalt() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *Server) VerifyPoW(salt string, timestamp int64, nonce string, targetHex string) bool {
	challengeTime := time.Unix(timestamp, 0)
	if s.maxTTL != 0 && time.Since(challengeTime) > s.maxTTL {
		return false
	}

	rawInput := buildKey(salt, timestamp, nonce)

	hashBytes := sha256.Sum256([]byte(rawInput))

	hashInt := new(big.Int).SetBytes(hashBytes[:])

	targetInt, success := new(big.Int).SetString(targetHex, 16)
	if !success {
		return false
	}

	//  Hash < Target
	return hashInt.Cmp(targetInt) < 0
}
