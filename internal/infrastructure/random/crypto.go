package random

import (
	"crypto/rand"
	"errors"
	"math/big"
)

type CryptoGenerator struct{}

func NewCryptoGenerator() *CryptoGenerator { return &CryptoGenerator{} }

func (*CryptoGenerator) Int64n(max int64) (int64, error) {
	if max <= 0 {
		return 0, errors.New("random upper bound must be positive")
	}
	value, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, err
	}
	return value.Int64(), nil
}
