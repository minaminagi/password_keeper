package service

import (
	"context"
	"testing"
	"time"

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

func (r *memoryVaultRepo) UpdateMeta(ctx context.Context, meta models.VaultMetaModel) error {
	if r.meta == nil {
		return pkerror.ErrVaultNotFound
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

func (r *memoryVaultRepo) UpdateKeyCheck(ctx context.Context, keyCheck models.VaultKeyCheckModel) error {
	if r.keyCheck == nil {
		return pkerror.ErrVaultKeyCheckNotFound
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

func (r *memoryVaultRepo) UpdateRecovery(ctx context.Context, recovery models.VaultRecoveryModel) error {
	if r.recovery == nil {
		return pkerror.ErrVaultRecoveryNotFound
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
		newMemoryItemRepo(),
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
		newMemoryItemRepo(),
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

func TestVaultServiceChangeMasterPasswordReencryptsItems(t *testing.T) {
	ctx := context.Background()
	session := NewMemoryVaultSession()
	vaultRepo := &memoryVaultRepo{}
	itemRepo := newMemoryItemRepo()
	encryptor := crypto.NewAESGCMEncryptor()
	svc := NewVaultService(
		vaultRepo,
		itemRepo,
		crypto.NewArgon2KeyDeriver(),
		encryptor,
		session,
	)

	meta, err := svc.Init(ctx, domain.InitVaultInput{
		VaultName:      "Personal",
		MasterPassword: "old-password",
	})
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	oldRecoveryCode := meta.RecoveryCode

	oldMasterKey, ok := session.GetMasterKey()
	if !ok {
		t.Fatalf("session is not unlocked after Init")
	}
	item := encryptedTestItem(t, encryptor, oldMasterKey)
	itemRepo.items[item.ID] = item

	changedMeta, err := svc.ChangeMasterPassword(ctx, domain.ChangeMasterPasswordInput{
		CurrentMasterPassword: "old-password",
		NewMasterPassword:     "new-password",
	})
	if err != nil {
		t.Fatalf("ChangeMasterPassword returned error: %v", err)
	}
	if changedMeta.RecoveryCode == "" || changedMeta.RecoveryCode == oldRecoveryCode {
		t.Fatalf("ChangeMasterPassword did not return a new recovery code")
	}

	if err := svc.Lock(ctx); err != nil {
		t.Fatalf("Lock returned error: %v", err)
	}
	if err := svc.Unlock(ctx, domain.UnlockVaultInput{MasterPassword: "old-password"}); err != pkerror.ErrInvalidMasterPassword {
		t.Fatalf("Unlock with old password returned %v, want ErrInvalidMasterPassword", err)
	}
	if err := svc.Unlock(ctx, domain.UnlockVaultInput{MasterPassword: "new-password"}); err != nil {
		t.Fatalf("Unlock with new password returned error: %v", err)
	}

	newMasterKey, ok := session.GetMasterKey()
	if !ok {
		t.Fatalf("session is not unlocked after new password unlock")
	}
	gotItem, err := itemRepo.GetByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	password, err := encryptor.Decrypt(gotItem.PasswordEnc, gotItem.NoncePassword, newMasterKey)
	if err != nil {
		t.Fatalf("decrypt reencrypted password returned error: %v", err)
	}
	if string(password) != "secret" {
		t.Fatalf("password = %q, want secret", password)
	}

	if err := svc.Lock(ctx); err != nil {
		t.Fatalf("Lock returned error: %v", err)
	}
	if err := svc.Recover(ctx, domain.RecoverVaultInput{RecoveryCode: oldRecoveryCode}); err != pkerror.ErrInvalidRecoveryCode {
		t.Fatalf("Recover with old recovery code returned %v, want ErrInvalidRecoveryCode", err)
	}
	if err := svc.Recover(ctx, domain.RecoverVaultInput{RecoveryCode: changedMeta.RecoveryCode}); err != nil {
		t.Fatalf("Recover with new recovery code returned error: %v", err)
	}
}

func encryptedTestItem(t *testing.T, encryptor crypto.Encryptor, masterKey []byte) models.ItemModel {
	t.Helper()
	username, err := encryptor.Encrypt([]byte("alice"), masterKey)
	if err != nil {
		t.Fatalf("encrypt username: %v", err)
	}
	password, err := encryptor.Encrypt([]byte("secret"), masterKey)
	if err != nil {
		t.Fatalf("encrypt password: %v", err)
	}
	website, err := encryptor.Encrypt([]byte("example.com"), masterKey)
	if err != nil {
		t.Fatalf("encrypt website: %v", err)
	}
	notes, err := encryptor.Encrypt([]byte("note"), masterKey)
	if err != nil {
		t.Fatalf("encrypt notes: %v", err)
	}

	now := time.Now()
	return models.ItemModel{
		ID:            "item-1",
		Title:         "Example",
		UsernameEnc:   username.CipherText,
		NonceUsername: username.Nonce,
		PasswordEnc:   password.CipherText,
		NoncePassword: password.Nonce,
		WebsiteEnc:    website.CipherText,
		NonceWebsite:  website.Nonce,
		NotesEnc:      notes.CipherText,
		NonceNotes:    notes.Nonce,
		Category:      "login",
		Favorite:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}
