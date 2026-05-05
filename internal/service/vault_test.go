package service

import (
	"context"
	"testing"

	"passwordkeeper/internal/crypto"
	"passwordkeeper/internal/domain"
	"passwordkeeper/internal/pkerror"
	"passwordkeeper/internal/repo/models"
)

type memoryVaultRepo struct {
	meta     *models.VaultMetaModel
	keyCheck *models.VaultKeyCheckModel
	recovery *models.VaultRecoveryModel
}

func (r *memoryVaultRepo) CreateMeta(ctx context.Context, meta models.VaultMetaModel) error {
	if r.meta != nil {
		return pkerror.ErrVaultAlreadyExists
	}
	r.meta = &meta
	return nil
}

func (r *memoryVaultRepo) Meta(ctx context.Context) (models.VaultMetaModel, error) {
	if r.meta == nil {
		return models.VaultMetaModel{}, pkerror.ErrVaultNotFound
	}
	return *r.meta, nil
}

func (r *memoryVaultRepo) VaultMetaExists(ctx context.Context) (bool, error) {
	return r.meta != nil, nil
}

func (r *memoryVaultRepo) SaveKeyCheck(ctx context.Context, keyCheck models.VaultKeyCheckModel) error {
	if r.keyCheck != nil {
		return pkerror.ErrVaultKeyCheckAlreadyExists
	}
	r.keyCheck = &keyCheck
	return nil
}

func (r *memoryVaultRepo) KeyCheck(ctx context.Context) (models.VaultKeyCheckModel, error) {
	if r.keyCheck == nil {
		return models.VaultKeyCheckModel{}, pkerror.ErrVaultKeyCheckNotFound
	}
	return *r.keyCheck, nil
}

func (r *memoryVaultRepo) SaveRecovery(ctx context.Context, recovery models.VaultRecoveryModel) error {
	if r.recovery != nil {
		return pkerror.ErrVaultRecoveryAlreadyExists
	}
	r.recovery = &recovery
	return nil
}

func (r *memoryVaultRepo) Recovery(ctx context.Context) (models.VaultRecoveryModel, error) {
	if r.recovery == nil {
		return models.VaultRecoveryModel{}, pkerror.ErrVaultRecoveryNotFound
	}
	return *r.recovery, nil
}

func TestVaultServiceRecoverUnlocksWithRecoveryCode(t *testing.T) {
	ctx := context.Background()
	session := NewMemoryVaultSession()
	svc := NewVaultService(
		&memoryVaultRepo{},
		crypto.NewArgon2KeyDeriver(),
		crypto.NewAESGCMEncryptor(),
		session,
	)

	meta, err := svc.Init(ctx, domain.InitVaultInput{
		VaultName:      "Personal",
		MasterPassword: "master-password",
	})
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	if meta.RecoveryCode == "" {
		t.Fatalf("Init returned empty recovery code")
	}

	if err := svc.Lock(ctx); err != nil {
		t.Fatalf("Lock returned error: %v", err)
	}
	if session.IsUnlocked() {
		t.Fatalf("session is unlocked after Lock")
	}

	if err := svc.Recover(ctx, domain.RecoverVaultInput{
		RecoveryCode: meta.RecoveryCode,
	}); err != nil {
		t.Fatalf("Recover returned error: %v", err)
	}
	if !session.IsUnlocked() {
		t.Fatalf("session is not unlocked after Recover")
	}
}

func TestVaultServiceRecoverRejectsWrongRecoveryCode(t *testing.T) {
	ctx := context.Background()
	session := NewMemoryVaultSession()
	svc := NewVaultService(
		&memoryVaultRepo{},
		crypto.NewArgon2KeyDeriver(),
		crypto.NewAESGCMEncryptor(),
		session,
	)

	if _, err := svc.Init(ctx, domain.InitVaultInput{
		VaultName:      "Personal",
		MasterPassword: "master-password",
	}); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	if err := svc.Lock(ctx); err != nil {
		t.Fatalf("Lock returned error: %v", err)
	}

	if err := svc.Recover(ctx, domain.RecoverVaultInput{
		RecoveryCode: "PKR-wrong-code",
	}); err != pkerror.ErrInvalidRecoveryCode {
		t.Fatalf("Recover returned %v, want ErrInvalidRecoveryCode", err)
	}
}
