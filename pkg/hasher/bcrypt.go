package hasher

import "golang.org/x/crypto/bcrypt"

type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost ...int) Hasher {
	c := bcrypt.DefaultCost
	if len(cost) > 0 && cost[0] > 0 {
		c = cost[0]
	}
	return &BcryptHasher{cost: c}
}

func (b *BcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (b *BcryptHasher) Compare(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
