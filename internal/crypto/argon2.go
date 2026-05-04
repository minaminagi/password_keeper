package crypto

import (
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	KDFAlgoArgon2id = "argon2id"
	DefaultKeyLen   = 32
)

type Argon2KeyDeriver struct{}

func NewArgon2KeyDeriver() *Argon2KeyDeriver {
	return &Argon2KeyDeriver{}
}

func (d *Argon2KeyDeriver) GenerateSalt(size int) ([]byte, error) {
	if size <= 0 {
		return nil, fmt.Errorf("salt size must be positive")
	}

	salt := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}
	return salt, nil
}

func (d *Argon2KeyDeriver) DeriveKey(password string, params KDFParams) ([]byte, error) {
	if params.Algo != "" && params.Algo != KDFAlgoArgon2id {
		return nil, fmt.Errorf("unsupported kdf algorithm: %s", params.Algo)
	}
	if len(params.Salt) == 0 {
		return nil, fmt.Errorf("kdf salt is empty")
	}

	keyLen := params.Keylength
	if keyLen == 0 {
		keyLen = DefaultKeyLen
	}

	return argon2.IDKey(
		[]byte(password),
		params.Salt,
		uint32(params.TimeCost),
		uint32(params.MemoryCost),
		uint8(params.Parallelism),
		keyLen,
	), nil
}
