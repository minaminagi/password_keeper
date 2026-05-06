package repo

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"passwordkeeper/internal/config"
	"passwordkeeper/internal/repo/models"
)

func TestBackupRepositorySnapshotAndRestore(t *testing.T) {
	ctx := context.Background()
	db, err := config.SQLiteService(filepath.Join(t.TempDir(), "vault.db"))
	if err != nil {
		t.Fatalf("init sqlite: %v", err)
	}

	vaultRepo := NewVaultRepository(db)
	itemRepo := NewItemRepository(db)
	tagRepo := NewTagRepository(db)
	itemTagRepo := NewItemTagRepository(db)
	backupRepo := NewBackupRepository(db)
	now := time.Unix(1714970000, 0)

	if err := vaultRepo.CreateMeta(ctx, models.VaultMetaModel{
		VaultName:      "test vault",
		KdfAlgo:        "argon2id",
		KdfSalt:        []byte("salt"),
		KdfTimeCost:    3,
		KdfMemoryCost:  65536,
		KdfParallelism: 4,
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create meta: %v", err)
	}
	if err := vaultRepo.SaveKeyCheck(ctx, models.VaultKeyCheckModel{
		Nonce:      []byte("key nonce"),
		CipherText: []byte("key cipher"),
		CreatedAt:  now,
	}); err != nil {
		t.Fatalf("save key check: %v", err)
	}
	if err := vaultRepo.SaveRecovery(ctx, models.VaultRecoveryModel{
		KdfAlgo:        "argon2id",
		KdfSalt:        []byte("recovery salt"),
		KdfTimeCost:    3,
		KdfMemoryCost:  65536,
		KdfParallelism: 4,
		Nonce:          []byte("recovery nonce"),
		CipherText:     []byte("recovery cipher"),
		CreatedAt:      now,
	}); err != nil {
		t.Fatalf("save recovery: %v", err)
	}
	if err := tagRepo.Create(ctx, models.TagModel{
		ID:        "tag-1",
		Name:      "work",
		CreatedAt: now,
	}); err != nil {
		t.Fatalf("create tag: %v", err)
	}
	if err := itemRepo.Create(ctx, models.ItemModel{
		ID:            "item-1",
		Title:         "GitHub",
		UsernameEnc:   []byte("username"),
		PasswordEnc:   []byte("password"),
		WebsiteEnc:    []byte("website"),
		NotesEnc:      []byte("notes"),
		NonceUsername: []byte("n1"),
		NoncePassword: []byte("n2"),
		NonceWebsite:  []byte("n3"),
		NonceNotes:    []byte("n4"),
		Category:      "dev",
		Favorite:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}); err != nil {
		t.Fatalf("create item: %v", err)
	}
	if err := itemTagRepo.ReplaceItemTags(ctx, "item-1", []string{"tag-1"}); err != nil {
		t.Fatalf("replace item tags: %v", err)
	}

	backup, err := backupRepo.Snapshot(ctx)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(backup.Items) != 1 || len(backup.Tags) != 1 || len(backup.ItemTags) != 1 {
		t.Fatalf("unexpected snapshot sizes: items=%d tags=%d itemTags=%d", len(backup.Items), len(backup.Tags), len(backup.ItemTags))
	}

	backup.Items[0].Title = "Restored GitHub"
	if err := backupRepo.Restore(ctx, backup); err != nil {
		t.Fatalf("restore: %v", err)
	}
	restored, err := itemRepo.GetByID(ctx, "item-1")
	if err != nil {
		t.Fatalf("get restored item: %v", err)
	}
	if restored.Title != "Restored GitHub" || !restored.Favorite {
		t.Fatalf("unexpected restored item: %#v", restored)
	}
	tagIDs, err := itemTagRepo.GetTagsIDsByItemID(ctx, "item-1")
	if err != nil {
		t.Fatalf("get restored item tags: %v", err)
	}
	if len(tagIDs) != 1 || tagIDs[0] != "tag-1" {
		t.Fatalf("unexpected restored tags: %#v", tagIDs)
	}
}
