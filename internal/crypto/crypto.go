package crypto

type KDFParams struct {
	Algo        string
	Salt        []byte
	TimeCost    int
	MemoryCost  int
	Parallelism int
	Keylength   uint32
}

type EncryptResult struct {
	CipherText []byte
	Nonce      []byte
}

type KeyDeriver interface {
	GenerateSalt(size int) ([]byte, error)
	DeriveKey(password string, params KDFParams) ([]byte, error)
}

type Encryptor interface {
	Encrypt(plainText []byte, key []byte) (EncryptResult, error)
	Decrypt(cipherText []byte, nonce []byte, key []byte) ([]byte, error)
}
