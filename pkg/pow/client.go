package pow

import (
	"crypto/sha256"
	"fmt"
	"math/big"
)

type Client struct {
}

func NewPOWClient() *Client {
	return &Client{}
}

func SolvePoW(salt string, timestamp int64, targetHex string) string {
	targetInt, _ := new(big.Int).SetString(targetHex, 16)
	nonceVal := 0

	for {
		nonceStr := fmt.Sprintf("%d", nonceVal)
		rawInput := buildKey(salt, timestamp, nonceStr)

		hashBytes := sha256.Sum256([]byte(rawInput))
		hashInt := new(big.Int).SetBytes(hashBytes[:])

		//  Hash < Target
		if hashInt.Cmp(targetInt) < 0 {
			return nonceStr
		}
		nonceVal++
	}
}
