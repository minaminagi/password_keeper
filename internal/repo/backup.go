package repo

import (
	"context"
	"fmt"
	"time"

	"passwordkeeper/internal/repo/models"

	"github.com/wailsapp/wails/v3/pkg/services/sqlite"
)

type BackupRepositoryImpl struct {
	db *sqlite.SQLiteService
}

func NewBackupRepository(db *sqlite.SQLiteService) *BackupRepositoryImpl {
	return &BackupRepositoryImpl{
		db: db,
	}
}

func (b *BackupRepositoryImpl) Snapshot(ctx context.Context) (models.VaultBackupModel, error) {
	vaultRepo := NewVaultRepository(b.db)
	meta, err := vaultRepo.Meta(ctx)
	if err != nil {
		return models.VaultBackupModel{}, err
	}
	keyCheck, err := vaultRepo.KeyCheck(ctx)
	if err != nil {
		return models.VaultBackupModel{}, err
	}
	recovery, err := vaultRepo.Recovery(ctx)
	if err != nil {
		return models.VaultBackupModel{}, err
	}
	items, err := b.listAllItems(ctx)
	if err != nil {
		return models.VaultBackupModel{}, err
	}
	tags, err := b.listAllTags(ctx)
	if err != nil {
		return models.VaultBackupModel{}, err
	}
	itemTags, err := b.listAllItemTags(ctx)
	if err != nil {
		return models.VaultBackupModel{}, err
	}

	return models.VaultBackupModel{
		Version:    1,
		ExportedAt: time.Now(),
		Meta:       meta,
		KeyCheck:   keyCheck,
		Recovery:   recovery,
		Items:      items,
		Tags:       tags,
		ItemTags:   itemTags,
	}, nil
}

func (b *BackupRepositoryImpl) Restore(ctx context.Context, backup models.VaultBackupModel) error {
	if err := b.db.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return fmt.Errorf("begin restore transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = b.db.ExecContext(ctx, "ROLLBACK")
		}
	}()

	for _, query := range []string{
		"DELETE FROM item_tags",
		"DELETE FROM items",
		"DELETE FROM tags",
		"DELETE FROM vault_recovery",
		"DELETE FROM vault_key_check",
		"DELETE FROM vault_meta",
	} {
		if err := b.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("clear backup table: %w", err)
		}
	}

	if err := b.insertMeta(ctx, backup.Meta); err != nil {
		return err
	}
	if err := b.insertKeyCheck(ctx, backup.KeyCheck); err != nil {
		return err
	}
	if err := b.insertRecovery(ctx, backup.Recovery); err != nil {
		return err
	}
	for _, tag := range backup.Tags {
		if err := b.insertTag(ctx, tag); err != nil {
			return err
		}
	}
	for _, item := range backup.Items {
		if err := b.insertItem(ctx, item); err != nil {
			return err
		}
	}
	for _, itemTag := range backup.ItemTags {
		if err := b.insertItemTag(ctx, itemTag); err != nil {
			return err
		}
	}

	if err := b.db.ExecContext(ctx, "COMMIT"); err != nil {
		return fmt.Errorf("commit restore transaction: %w", err)
	}
	committed = true
	return nil
}

func (b *BackupRepositoryImpl) listAllItems(ctx context.Context) ([]models.ItemModel, error) {
	rows, err := b.db.QueryContext(ctx, "SELECT * FROM items ORDER BY updated_at DESC")
	if err != nil {
		return nil, fmt.Errorf("list backup items: %w", err)
	}
	items := make([]models.ItemModel, 0, len(rows))
	for _, row := range rows {
		item, err := mapItemModel(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (b *BackupRepositoryImpl) listAllTags(ctx context.Context) ([]models.TagModel, error) {
	rows, err := b.db.QueryContext(ctx, "SELECT * FROM tags ORDER BY name ASC")
	if err != nil {
		return nil, fmt.Errorf("list backup tags: %w", err)
	}
	tags := make([]models.TagModel, 0, len(rows))
	for _, row := range rows {
		tags = append(tags, models.TagModel{
			ID:        row["id"].(string),
			Name:      row["name"].(string),
			CreatedAt: row["created_at"].(time.Time),
		})
	}
	return tags, nil
}

func (b *BackupRepositoryImpl) listAllItemTags(ctx context.Context) ([]models.ItemTagModel, error) {
	rows, err := b.db.QueryContext(ctx, "SELECT * FROM item_tags ORDER BY item_id ASC, tag_id ASC")
	if err != nil {
		return nil, fmt.Errorf("list backup item tags: %w", err)
	}
	itemTags := make([]models.ItemTagModel, 0, len(rows))
	for _, row := range rows {
		itemTags = append(itemTags, models.ItemTagModel{
			ItemID: row["item_id"].(string),
			TagID:  row["tag_id"].(string),
		})
	}
	return itemTags, nil
}

func (b *BackupRepositoryImpl) insertMeta(ctx context.Context, meta models.VaultMetaModel) error {
	if err := b.db.ExecContext(ctx, `
		INSERT INTO vault_meta (
			id,
			vault_name,
			kdf_algo,
			kdf_salt,
			kdf_time_cost,
			kdf_memory_cost,
			kdf_parallelism,
			created_at,
			updated_at
		)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?)
	`, meta.VaultName,
		meta.KdfAlgo,
		meta.KdfSalt,
		meta.KdfTimeCost,
		meta.KdfMemoryCost,
		meta.KdfParallelism,
		meta.CreatedAt,
		meta.UpdatedAt,
	); err != nil {
		return fmt.Errorf("restore vault meta: %w", err)
	}
	return nil
}

func (b *BackupRepositoryImpl) insertKeyCheck(ctx context.Context, keyCheck models.VaultKeyCheckModel) error {
	if err := b.db.ExecContext(ctx, `
		INSERT INTO vault_key_check (id, nonce, cipher_text, created_at)
		VALUES (1, ?, ?, ?)
	`, keyCheck.Nonce, keyCheck.CipherText, keyCheck.CreatedAt); err != nil {
		return fmt.Errorf("restore vault key check: %w", err)
	}
	return nil
}

func (b *BackupRepositoryImpl) insertRecovery(ctx context.Context, recovery models.VaultRecoveryModel) error {
	if err := b.db.ExecContext(ctx, `
		INSERT INTO vault_recovery (
			id,
			kdf_algo,
			kdf_salt,
			kdf_time_cost,
			kdf_memory_cost,
			kdf_parallelism,
			nonce,
			cipher_text,
			created_at
		)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?)
	`, recovery.KdfAlgo,
		recovery.KdfSalt,
		recovery.KdfTimeCost,
		recovery.KdfMemoryCost,
		recovery.KdfParallelism,
		recovery.Nonce,
		recovery.CipherText,
		recovery.CreatedAt,
	); err != nil {
		return fmt.Errorf("restore vault recovery: %w", err)
	}
	return nil
}

func (b *BackupRepositoryImpl) insertTag(ctx context.Context, tag models.TagModel) error {
	if err := b.db.ExecContext(ctx, `
		INSERT INTO tags (id, name, created_at)
		VALUES (?, ?, ?)
	`, tag.ID, tag.Name, tag.CreatedAt); err != nil {
		return fmt.Errorf("restore tag: %w", err)
	}
	return nil
}

func (b *BackupRepositoryImpl) insertItem(ctx context.Context, item models.ItemModel) error {
	if err := b.db.ExecContext(ctx, `
		INSERT INTO items (
			id,
			title,
			username_enc,
			password_enc,
			website_enc,
			notes_enc,
			nonce_username,
			nonce_password,
			nonce_website,
			nonce_notes,
			category,
			favorite,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, item.ID,
		item.Title,
		item.UsernameEnc,
		item.PasswordEnc,
		item.WebsiteEnc,
		item.NotesEnc,
		item.NonceUsername,
		item.NoncePassword,
		item.NonceWebsite,
		item.NonceNotes,
		item.Category,
		sqliteBoolInt(item.Favorite),
		item.CreatedAt,
		item.UpdatedAt,
	); err != nil {
		return fmt.Errorf("restore item: %w", err)
	}
	return nil
}

func (b *BackupRepositoryImpl) insertItemTag(ctx context.Context, itemTag models.ItemTagModel) error {
	if err := b.db.ExecContext(ctx, `
		INSERT INTO item_tags (item_id, tag_id)
		VALUES (?, ?)
	`, itemTag.ItemID, itemTag.TagID); err != nil {
		return fmt.Errorf("restore item tag: %w", err)
	}
	return nil
}
