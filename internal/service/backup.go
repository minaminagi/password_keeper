package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"passwordkeeper/internal/crypto"
	"passwordkeeper/internal/domain"
	"passwordkeeper/internal/pkerror"
	"passwordkeeper/internal/repo"
	"passwordkeeper/internal/repo/models"
)

const (
	backupTextPrefix  = "PKB1."
	backupDataVersion = 1
)

type backupService struct {
	backupRepo repo.BackupRepository
	keyDeriver crypto.KeyDeriver
	encryptor  crypto.Encryptor
	session    VaultSession
}

type backupEnvelope struct {
	Version    int       `json:"version"`
	KDF        backupKDF `json:"kdf"`
	Nonce      []byte    `json:"nonce"`
	CipherText []byte    `json:"cipher_text"`
}

type backupKDF struct {
	Algo        string `json:"algo"`
	Salt        []byte `json:"salt"`
	TimeCost    int    `json:"time_cost"`
	MemoryCost  int    `json:"memory_cost"`
	Parallelism int    `json:"parallelism"`
}

func NewBackupService(
	backupRepo repo.BackupRepository,
	keyDeriver crypto.KeyDeriver,
	encryptor crypto.Encryptor,
	session VaultSession,
) BackupService {
	return &backupService{
		backupRepo: backupRepo,
		keyDeriver: keyDeriver,
		encryptor:  encryptor,
		session:    session,
	}
}

func (s *backupService) Export(ctx context.Context, input domain.ExportBackupInput) (string, error) {
	if strings.TrimSpace(input.ExportPassword) == "" {
		return "", pkerror.ErrInvalidBackupPassword
	}
	password := input.ExportPassword

	backup, err := s.backupRepo.Snapshot(ctx)
	if err != nil {
		return "", err
	}
	backup.Version = backupDataVersion

	plainText, err := json.Marshal(backup)
	if err != nil {
		return "", fmt.Errorf("marshal backup: %w", err)
	}

	kdfParams, err := s.newBackupKDFParams()
	if err != nil {
		return "", err
	}
	key, err := s.keyDeriver.DeriveKey(password, kdfParams)
	if err != nil {
		return "", err
	}
	encrypted, err := s.encryptor.Encrypt(plainText, key)
	if err != nil {
		return "", err
	}

	envelope := backupEnvelope{
		Version: backupDataVersion,
		KDF: backupKDF{
			Algo:        kdfParams.Algo,
			Salt:        kdfParams.Salt,
			TimeCost:    kdfParams.TimeCost,
			MemoryCost:  kdfParams.MemoryCost,
			Parallelism: kdfParams.Parallelism,
		},
		Nonce:      encrypted.Nonce,
		CipherText: encrypted.CipherText,
	}
	envelopeJSON, err := json.Marshal(envelope)
	if err != nil {
		return "", fmt.Errorf("marshal backup envelope: %w", err)
	}

	return backupTextPrefix + base64.RawURLEncoding.EncodeToString(envelopeJSON), nil
}

func (s *backupService) Import(ctx context.Context, input domain.ImportBackupInput) error {
	if strings.TrimSpace(input.ExportPassword) == "" {
		return pkerror.ErrInvalidBackupPassword
	}
	password := input.ExportPassword

	backup, err := s.decryptBackup(input.CipherText, password)
	if err != nil {
		return err
	}
	if backup.Version != backupDataVersion {
		return pkerror.ErrInvalidBackupFormat
	}

	if err := s.backupRepo.Restore(ctx, backup); err != nil {
		return err
	}
	s.session.Clear()
	return nil
}

func (s *backupService) decryptBackup(cipherText string, password string) (models.VaultBackupModel, error) {
	encoded := strings.TrimSpace(cipherText)
	if !strings.HasPrefix(encoded, backupTextPrefix) {
		return models.VaultBackupModel{}, pkerror.ErrInvalidBackupFormat
	}
	envelopeJSON, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(encoded, backupTextPrefix))
	if err != nil {
		return models.VaultBackupModel{}, pkerror.ErrInvalidBackupFormat
	}

	var envelope backupEnvelope
	if err := json.Unmarshal(envelopeJSON, &envelope); err != nil {
		return models.VaultBackupModel{}, pkerror.ErrInvalidBackupFormat
	}
	if envelope.Version != backupDataVersion || len(envelope.Nonce) == 0 || len(envelope.CipherText) == 0 {
		return models.VaultBackupModel{}, pkerror.ErrInvalidBackupFormat
	}

	key, err := s.keyDeriver.DeriveKey(password, crypto.KDFParams{
		Algo:        envelope.KDF.Algo,
		Salt:        envelope.KDF.Salt,
		TimeCost:    envelope.KDF.TimeCost,
		MemoryCost:  envelope.KDF.MemoryCost,
		Parallelism: envelope.KDF.Parallelism,
		Keylength:   crypto.DefaultKeyLen,
	})
	if err != nil {
		return models.VaultBackupModel{}, pkerror.ErrInvalidBackupFormat
	}

	plainText, err := s.encryptor.Decrypt(envelope.CipherText, envelope.Nonce, key)
	if err != nil {
		return models.VaultBackupModel{}, pkerror.ErrInvalidBackupPassword
	}

	var backup models.VaultBackupModel
	if err := json.Unmarshal(plainText, &backup); err != nil {
		return models.VaultBackupModel{}, pkerror.ErrInvalidBackupFormat
	}
	return backup, nil
}

func (s *backupService) newBackupKDFParams() (crypto.KDFParams, error) {
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
