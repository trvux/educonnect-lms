package security

import "golang.org/x/crypto/bcrypt"

// BcryptHasher hiện thực service/auth.PasswordHasher.
type BcryptHasher struct {
	Cost int
}

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{Cost: bcrypt.DefaultCost}
}

func (h *BcryptHasher) Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), h.Cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (h *BcryptHasher) Compare(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
