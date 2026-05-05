package service

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"passwordkeeper/internal/crypto"
	"passwordkeeper/internal/domain"
	"passwordkeeper/internal/pkerror"
	"passwordkeeper/internal/repo"
	"passwordkeeper/internal/repo/models"
)

type vaultService struct {
	vaultRepo  repo.VaultRepository
	itemRepo   repo.ItemRepository
	keyDeriver crypto.KeyDeriver
	encryptor  crypto.Encryptor
	session    VaultSession
}

const vaultKeyCheckPlainText = "password_keeper:vault_key_check:v1"
const recoveryCodePrefix = "PKR"

func NewVaultService(
	vaultRepo repo.VaultRepository,
	itemRepo repo.ItemRepository,
	keyDeriver crypto.KeyDeriver,
	encryptor crypto.Encryptor,
	session VaultSession,
) VaultService {
	return &vaultService{
		vaultRepo:  vaultRepo,
		itemRepo:   itemRepo,
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
	recoveryCode, recoveryModel, err := s.buildRecoveryModel(masterKey, now)
	if err != nil {
		return domain.VaultMeta{}, err
	}

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
	if err := s.vaultRepo.SaveRecovery(ctx, recoveryModel); err != nil {
		return domain.VaultMeta{}, err
	}

	s.session.SetMasterKey(masterKey)

	return domain.VaultMeta{
		Name:         input.VaultName,
		RecoveryCode: recoveryCode,
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

func (s *vaultService) Recover(ctx context.Context, input domain.RecoverVaultInput) error {
	masterKey, err := s.masterKeyFromRecovery(ctx, input.RecoveryCode)
	if err != nil {
		return err
	}

	s.session.SetMasterKey(masterKey)
	return nil
}

func (s *vaultService) ChangeMasterPassword(ctx context.Context, input domain.ChangeMasterPasswordInput) (domain.VaultMeta, error) {
	if strings.TrimSpace(input.NewMasterPassword) == "" {
		return domain.VaultMeta{}, pkerror.ErrInvalidMasterPassword
	}

	oldMasterKey, err := s.masterKeyForPasswordChange(ctx, input)
	if err != nil {
		return domain.VaultMeta{}, err
	}

	meta, err := s.vaultRepo.Meta(ctx)
	if err != nil {
		return domain.VaultMeta{}, err
	}

	newKDFParams, err := s.newPasswordKDFParams()
	if err != nil {
		return domain.VaultMeta{}, err
	}
	newMasterKey, err := s.keyDeriver.DeriveKey(input.NewMasterPassword, newKDFParams)
	if err != nil {
		return domain.VaultMeta{}, err
	}

	rekeyedItems, err := s.reencryptItems(ctx, oldMasterKey, newMasterKey)
	if err != nil {
		return domain.VaultMeta{}, err
	}

	keyCheck, err := s.encryptor.Encrypt([]byte(vaultKeyCheckPlainText), newMasterKey)
	if err != nil {
		return domain.VaultMeta{}, err
	}

	now := time.Now()
	recoveryCode, recoveryModel, err := s.buildRecoveryModel(newMasterKey, now)
	if err != nil {
		return domain.VaultMeta{}, err
	}

	meta.KdfAlgo = newKDFParams.Algo
	meta.KdfSalt = newKDFParams.Salt
	meta.KdfTimeCost = newKDFParams.TimeCost
	meta.KdfMemoryCost = newKDFParams.MemoryCost
	meta.KdfParallelism = newKDFParams.Parallelism
	meta.UpdatedAt = now

	for _, item := range rekeyedItems {
		if err := s.itemRepo.Update(ctx, item); err != nil {
			return domain.VaultMeta{}, err
		}
	}
	if err := s.vaultRepo.UpdateMeta(ctx, meta); err != nil {
		return domain.VaultMeta{}, err
	}
	if err := s.vaultRepo.UpdateKeyCheck(ctx, models.VaultKeyCheckModel{
		Nonce:      keyCheck.Nonce,
		CipherText: keyCheck.CipherText,
		CreatedAt:  now,
	}); err != nil {
		return domain.VaultMeta{}, err
	}
	if err := s.vaultRepo.UpdateRecovery(ctx, recoveryModel); err != nil {
		return domain.VaultMeta{}, err
	}

	s.session.SetMasterKey(newMasterKey)
	result := toDomainVaultMeta(meta)
	result.RecoveryCode = recoveryCode
	return result, nil
}

func (s *vaultService) masterKeyForPasswordChange(ctx context.Context, input domain.ChangeMasterPasswordInput) ([]byte, error) {
	if strings.TrimSpace(input.RecoveryCode) != "" {
		return s.masterKeyFromRecovery(ctx, input.RecoveryCode)
	}
	if input.CurrentMasterPassword != "" {
		return s.masterKeyFromPassword(ctx, input.CurrentMasterPassword)
	}
	return nil, pkerror.ErrInvalidMasterPassword
}

func (s *vaultService) masterKeyFromPassword(ctx context.Context, masterPassword string) ([]byte, error) {
	meta, err := s.vaultRepo.Meta(ctx)
	if err != nil {
		return nil, err
	}
	kdfParams := crypto.KDFParams{
		Algo:        meta.KdfAlgo,
		Salt:        meta.KdfSalt,
		TimeCost:    meta.KdfTimeCost,
		MemoryCost:  meta.KdfMemoryCost,
		Parallelism: meta.KdfParallelism,
		Keylength:   crypto.DefaultKeyLen,
	}
	masterKey, err := s.keyDeriver.DeriveKey(masterPassword, kdfParams)
	if err != nil {
		return nil, err
	}
	if err := s.validateMasterKey(ctx, masterKey, pkerror.ErrInvalidMasterPassword); err != nil {
		return nil, err
	}
	return masterKey, nil
}

func (s *vaultService) masterKeyFromRecovery(ctx context.Context, recoveryCode string) ([]byte, error) {
	recovery, err := s.vaultRepo.Recovery(ctx)
	if err != nil {
		return nil, err
	}

	recoveryKey, err := s.keyDeriver.DeriveKey(normalizeRecoveryCode(recoveryCode), crypto.KDFParams{
		Algo:        recovery.KdfAlgo,
		Salt:        recovery.KdfSalt,
		TimeCost:    recovery.KdfTimeCost,
		MemoryCost:  recovery.KdfMemoryCost,
		Parallelism: recovery.KdfParallelism,
		Keylength:   crypto.DefaultKeyLen,
	})
	if err != nil {
		return nil, err
	}

	masterKey, err := s.encryptor.Decrypt(recovery.CipherText, recovery.Nonce, recoveryKey)
	if err != nil {
		return nil, pkerror.ErrInvalidRecoveryCode
	}

	if err := s.validateMasterKey(ctx, masterKey, pkerror.ErrInvalidRecoveryCode); err != nil {
		return nil, err
	}
	return masterKey, nil
}

func (s *vaultService) validateMasterKey(ctx context.Context, masterKey []byte, invalidErr error) error {
	keyCheck, err := s.vaultRepo.KeyCheck(ctx)
	if err != nil {
		return err
	}
	plainText, err := s.encryptor.Decrypt(keyCheck.CipherText, keyCheck.Nonce, masterKey)
	if err != nil || string(plainText) != vaultKeyCheckPlainText {
		return invalidErr
	}
	return nil
}

func (s *vaultService) newPasswordKDFParams() (crypto.KDFParams, error) {
	salt, err := s.keyDeriver.GenerateSalt(16)
	if err != nil {
		return crypto.KDFParams{}, err
	}
	return crypto.KDFParams{
		Algo:        crypto.KDFAlgoArgon2id,
		Salt:        salt,
		TimeCost:    3,
		MemoryCost:  64 * 1024,
		Parallelism: 4,
		Keylength:   crypto.DefaultKeyLen,
	}, nil
}

func (s *vaultService) reencryptItems(ctx context.Context, oldMasterKey []byte, newMasterKey []byte) ([]models.ItemModel, error) {
	items, err := s.itemRepo.List(ctx, models.ItemListFilter{})
	if err != nil {
		return nil, err
	}

	rekeyedItems := make([]models.ItemModel, 0, len(items))
	for _, item := range items {
		rekeyed, err := s.reencryptItem(item, oldMasterKey, newMasterKey)
		if err != nil {
			return nil, err
		}
		rekeyedItems = append(rekeyedItems, rekeyed)
	}
	return rekeyedItems, nil
}

func (s *vaultService) reencryptItem(item models.ItemModel, oldMasterKey []byte, newMasterKey []byte) (models.ItemModel, error) {
	username, err := s.reencryptField(item.UsernameEnc, item.NonceUsername, oldMasterKey, newMasterKey)
	if err != nil {
		return models.ItemModel{}, err
	}
	password, err := s.reencryptField(item.PasswordEnc, item.NoncePassword, oldMasterKey, newMasterKey)
	if err != nil {
		return models.ItemModel{}, err
	}
	website, err := s.reencryptField(item.WebsiteEnc, item.NonceWebsite, oldMasterKey, newMasterKey)
	if err != nil {
		return models.ItemModel{}, err
	}
	notes, err := s.reencryptField(item.NotesEnc, item.NonceNotes, oldMasterKey, newMasterKey)
	if err != nil {
		return models.ItemModel{}, err
	}

	item.UsernameEnc = username.CipherText
	item.NonceUsername = username.Nonce
	item.PasswordEnc = password.CipherText
	item.NoncePassword = password.Nonce
	item.WebsiteEnc = website.CipherText
	item.NonceWebsite = website.Nonce
	item.NotesEnc = notes.CipherText
	item.NonceNotes = notes.Nonce
	item.UpdatedAt = time.Now()
	return item, nil
}

func (s *vaultService) reencryptField(cipherText []byte, nonce []byte, oldMasterKey []byte, newMasterKey []byte) (crypto.EncryptResult, error) {
	plainText, err := s.encryptor.Decrypt(cipherText, nonce, oldMasterKey)
	if err != nil {
		return crypto.EncryptResult{}, err
	}
	return s.encryptor.Encrypt(plainText, newMasterKey)
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

func (s *vaultService) buildRecoveryModel(masterKey []byte, createdAt time.Time) (string, models.VaultRecoveryModel, error) {
	codeBytes, err := s.keyDeriver.GenerateSalt(18)
	if err != nil {
		return "", models.VaultRecoveryModel{}, err
	}
	recoveryCode := recoveryCodePrefix + "-" + base64.RawURLEncoding.EncodeToString(codeBytes)

	salt, err := s.keyDeriver.GenerateSalt(16)
	if err != nil {
		return "", models.VaultRecoveryModel{}, err
	}
	kdfParams := crypto.KDFParams{
		Algo:        crypto.KDFAlgoArgon2id,
		Salt:        salt,
		TimeCost:    3,
		MemoryCost:  64 * 1024,
		Parallelism: 4,
		Keylength:   crypto.DefaultKeyLen,
	}
	recoveryKey, err := s.keyDeriver.DeriveKey(normalizeRecoveryCode(recoveryCode), kdfParams)
	if err != nil {
		return "", models.VaultRecoveryModel{}, err
	}
	encryptedMasterKey, err := s.encryptor.Encrypt(masterKey, recoveryKey)
	if err != nil {
		return "", models.VaultRecoveryModel{}, err
	}

	return recoveryCode, models.VaultRecoveryModel{
		KdfAlgo:        kdfParams.Algo,
		KdfSalt:        kdfParams.Salt,
		KdfTimeCost:    kdfParams.TimeCost,
		KdfMemoryCost:  kdfParams.MemoryCost,
		KdfParallelism: kdfParams.Parallelism,
		Nonce:          encryptedMasterKey.Nonce,
		CipherText:     encryptedMasterKey.CipherText,
		CreatedAt:      createdAt,
	}, nil
}

func normalizeRecoveryCode(code string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(code), " ", ""))
}
