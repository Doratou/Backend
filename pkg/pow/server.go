package pow

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

// HMAC payload: "<salt>:<timestamp>:<difficulty>"
// PoW payload:  "<salt>:<timestamp>:<nonce>"   (same buildKey helper)
type Server struct {
	maxTTL    time.Duration
	secretKey []byte
}

// Challenge is the JSON-serialisable challenge issued to a client.
// All fields needed to solve and later verify the puzzle are embedded here.
type Challenge struct {
	// Salt is a 16-byte random hex string; prevents precomputed solutions.
	Salt string `json:"salt"`

	// Timestamp is the Unix second at which the challenge was issued.
	// Used to enforce the TTL window on the server side.
	Timestamp int64 `json:"timestamp"`

	// HMAC is HMAC-SHA256 over "<salt>:<timestamp>:<difficulty>" with the
	// server's secret key.  It proves the challenge was issued by this server
	// and that the difficulty value has not been tampered with.
	HMAC string `json:"hmac"`

	// Target is the upper bound (exclusive) the hash must be below.
	// Encoded as a 64-character lowercase hex string (256-bit big-endian).
	// Derived from Difficulty: target = 2^(256 - Difficulty).
	Target string `json:"target"`

	// Difficulty is the number of leading zero bits required in the solution hash.
	Difficulty int `json:"difficulty"`
}

// maxTTL: how long an issued challenge remains valid (0 = no limit).
// secretKey: HMAC signing key; keep it secret and rotate periodically.
func NewPOWServer(maxTTL time.Duration, secretKey string) *Server {
	return &Server{
		maxTTL:    maxTTL,
		secretKey: []byte(secretKey),
	}
}

// GenerateSalt returns a 16-byte cryptographically random hex string.
func (s *Server) GenerateSalt() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateChallenge issues a fresh stateless challenge.
func (s *Server) GenerateChallenge(difficulty int) Challenge {
	salt := s.GenerateSalt()
	ts := time.Now().Unix()
	return Challenge{
		Salt:       salt,
		Timestamp:  ts,
		HMAC:       s.computeHMAC(salt, ts, difficulty),
		Target:     computeTarget(difficulty),
		Difficulty: difficulty,
	}
}

// VerifyChallenge performs a fully stateless three-step verification:
//
//  1. TTL check   – challenge must not be older than maxTTL.
//  2. HMAC check  – ensures the challenge was issued by this server and that
//     salt, timestamp, and difficulty have not been tampered with.
//     Comparison is done in constant time to resist timing attacks.
//  3. PoW check   – SHA-256(salt:timestamp:nonce) must be strictly less than
//     the target derived from the HMAC-authenticated difficulty.
func (s *Server) VerifyChallenge(salt string, timestamp int64, hmacHex string, difficulty int, nonce string) bool {
	// 1. TTL
	if s.maxTTL != 0 && time.Since(time.Unix(timestamp, 0)) > s.maxTTL {
		return false
	}

	// 2. HMAC integrity (constant-time comparison)
	expected := s.computeHMAC(salt, timestamp, difficulty)
	if !hmac.Equal([]byte(hmacHex), []byte(expected)) {
		return false
	}

	// 3. PoW: re-derive the target from the HMAC-authenticated difficulty so the
	//    client cannot use a weaker target than the server originally issued.
	targetInt, ok := new(big.Int).SetString(computeTarget(difficulty), 16)
	if !ok {
		return false
	}

	hashBytes := sha256.Sum256([]byte(buildKey(salt, timestamp, nonce)))
	hashInt := new(big.Int).SetBytes(hashBytes[:])

	// Hash must be strictly less than target.
	return hashInt.Cmp(targetInt) < 0
}

// computeHMAC returns HMAC-SHA256("<salt>:<timestamp>:<difficulty>") as hex.
func (s *Server) computeHMAC(salt string, timestamp int64, difficulty int) string {
	h := hmac.New(sha256.New, s.secretKey)
	h.Write(fmt.Appendf(nil, "%s:%d:%d", salt, timestamp, difficulty))
	return hex.EncodeToString(h.Sum(nil))
}
