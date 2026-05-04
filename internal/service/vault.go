package service

import (
	"context"
	"time"

	"passwordkeeper/internal/crypto"
	"passwordkeeper/internal/domain"
	"passwordkeeper/internal/pkerror"
	"passwordkeeper/internal/repo"
	"passwordkeeper/internal/repo/models"
)

type vaultService struct {
	vaultRepo  repo.VaultRepository
	keyDeriver crypto.KeyDeriver
	encryptor  crypto.Encryptor
	session    VaultSession
}

const vaultKeyCheckPlainText = "password_keeper:vault_key_check:v1"

func NewVaultService(
	vaultRepo repo.VaultRepository,
	keyDeriver crypto.KeyDeriver,
	encryptor crypto.Encryptor,
	session VaultSession,
) VaultService {
	return &vaultService{
		vaultRepo:  vaultRepo,
		keyDeriver: keyDeriver,
		encryptor:  encryptor,
		session:    session,
	}
}

func (s *vaultService) Init(ctx context.Context, input domain.InitVaultInput) (domain.VaultMeta, error) {
	exists, err := s.vaultRepo.VaultMetaExists(ctx)
	if err != nil {
		return domain.VaultMeta{}, err
	}
	if exists {
		return domain.VaultMeta{}, pkerror.ErrVaultAlreadyExists
	}

	salt, err := s.keyDeriver.GenerateSalt(16)
	if err != nil {
		return domain.VaultMeta{}, err
	}

	kdfParams := crypto.KDFParams{
		Algo:        crypto.KDFAlgoArgon2id,
		Salt:        salt,
		TimeCost:    3,
		MemoryCost:  64 * 1024,
		Parallelism: 4,
		Keylength:   crypto.DefaultKeyLen,
	}

	masterKey, err := s.keyDeriver.DeriveKey(input.MasterPassword, kdfParams)
	if err != nil {
		return domain.VaultMeta{}, err
	}

	keyCheck, err := s.encryptor.Encrypt([]byte(vaultKeyCheckPlainText), masterKey)
	if err != nil {
		return domain.VaultMeta{}, err
	}

	now := time.Now()
	metaModel := models.VaultMetaModel{
		VaultName:      input.VaultName,
		KdfAlgo:        kdfParams.Algo,
		KdfSalt:        kdfParams.Salt,
		KdfTimeCost:    kdfParams.TimeCost,
		KdfMemoryCost:  kdfParams.MemoryCost,
		KdfParallelism: kdfParams.Parallelism,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.vaultRepo.CreateMeta(ctx, metaModel); err != nil {
		return domain.VaultMeta{}, err
	}

	if err := s.vaultRepo.SaveKeyCheck(ctx, models.VaultKeyCheckModel{
		Nonce:      keyCheck.Nonce,
		CipherText: keyCheck.CipherText,
		CreatedAt:  now,
	}); err != nil {
		return domain.VaultMeta{}, err
	}

	s.session.SetMasterKey(masterKey)

	return domain.VaultMeta{
		Name: input.VaultName,
		KDF: domain.KDFParams{
			Algo:        kdfParams.Algo,
			Salt:        kdfParams.Salt,
			TimeCost:    metaModel.KdfTimeCost,
			MemoryCost:  metaModel.KdfMemoryCost,
			Parallelism: metaModel.KdfParallelism,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *vaultService) Unlock(ctx context.Context, input domain.UnlockVaultInput) error {
	meta, err := s.vaultRepo.Meta(ctx)
	if err != nil {
		return err
	}

	kdfParams := crypto.KDFParams{
		Algo:        meta.KdfAlgo,
		Salt:        meta.KdfSalt,
		TimeCost:    meta.KdfTimeCost,
		MemoryCost:  meta.KdfMemoryCost,
		Parallelism: meta.KdfParallelism,
		Keylength:   crypto.DefaultKeyLen,
	}

	masterKey, err := s.keyDeriver.DeriveKey(input.MasterPassword, kdfParams)
	if err != nil {
		return err
	}

	keyCheck, err := s.vaultRepo.KeyCheck(ctx)
	if err != nil {
		return err
	}

	plainText, err := s.encryptor.Decrypt(keyCheck.CipherText, keyCheck.Nonce, masterKey)
	if err != nil {
		return pkerror.ErrInvalidMasterPassword
	}
	if string(plainText) != vaultKeyCheckPlainText {
		return pkerror.ErrInvalidMasterPassword
	}

	s.session.SetMasterKey(masterKey)
	return nil
}

func (s *vaultService) IsInitialized(ctx context.Context) (bool, error) {
	return s.vaultRepo.VaultMetaExists(ctx)
}

func (s *vaultService) Lock(ctx context.Context) error {
	s.session.Clear()
	return nil
}

func (s *vaultService) GetMeta(ctx context.Context) (domain.VaultMeta, error) {
	meta, err := s.vaultRepo.Meta(ctx)
	if err != nil {
		return domain.VaultMeta{}, err
	}
	return toDomainVaultMeta(meta), err
}
