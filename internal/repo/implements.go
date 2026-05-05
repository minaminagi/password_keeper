package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"passwordkeeper/internal/pkerror"
	"passwordkeeper/internal/repo/models"

	"github.com/wailsapp/wails/v3/pkg/services/sqlite"
)

type VaultRepositoryImpl struct {
	db *sqlite.SQLiteService
}

type ItemRepositoryImpl struct {
	db *sqlite.SQLiteService
}

type TagRepositoryImpl struct {
	db *sqlite.SQLiteService
}

type TxManagerImpl struct {
	db *sqlite.SQLiteService
}

type ItemTagRepositoryImpl struct {
	db *sqlite.SQLiteService
}

func NewVaultRepository(db *sqlite.SQLiteService) *VaultRepositoryImpl {
	return &VaultRepositoryImpl{
		db: db,
	}
}

func NewItemRepository(db *sqlite.SQLiteService) *ItemRepositoryImpl {
	return &ItemRepositoryImpl{
		db: db,
	}
}

func NewTagRepository(db *sqlite.SQLiteService) *TagRepositoryImpl {
	return &TagRepositoryImpl{
		db: db,
	}
}

func NewItemTagRepository(db *sqlite.SQLiteService) *ItemTagRepositoryImpl {
	return &ItemTagRepositoryImpl{
		db: db,
	}
}

func NewTxManager(db *sqlite.SQLiteService) *TxManagerImpl {
	return &TxManagerImpl{
		db: db,
	}
}

func (v *VaultRepositoryImpl) CreateMeta(ctx context.Context, meta models.VaultMetaModel) error {
	exists, err := v.VaultMetaExists(ctx)
	if err != nil {
		return err
	}
	if exists {
		return pkerror.ErrVaultAlreadyExists
	}
	if err := v.db.ExecContext(ctx, `
		INSERT INTO vault_meta (vault_name, kdf_algo,kdf_salt,kdf_time_cost,kdf_memory_cost,kdf_parallelism,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?)
	`, meta.VaultName, meta.KdfAlgo, meta.KdfSalt, meta.KdfTimeCost, meta.KdfMemoryCost, meta.KdfParallelism, meta.CreatedAt, meta.UpdatedAt); err != nil {
		return fmt.Errorf("insert vault meta: %w", err)
	}

	return nil
}

func (v *VaultRepositoryImpl) Meta(ctx context.Context) (models.VaultMetaModel, error) {
	rows, err := v.db.QueryContext(ctx, `
		SELECT * FROM vault_meta WHERE id = 1
	`)
	if err != nil {
		return models.VaultMetaModel{}, fmt.Errorf("query vault meta: %w", err)
	}
	if len(rows) == 0 {
		return models.VaultMetaModel{}, pkerror.ErrVaultNotFound
	}
	row := rows[0]
	kdfTimeCost, err := sqliteInt(row["kdf_time_cost"])
	if err != nil {
		return models.VaultMetaModel{}, fmt.Errorf("map vault meta kdf_time_cost: %w", err)
	}
	kdfMemoryCost, err := sqliteInt(row["kdf_memory_cost"])
	if err != nil {
		return models.VaultMetaModel{}, fmt.Errorf("map vault meta kdf_memory_cost: %w", err)
	}
	kdfParallelism, err := sqliteInt(row["kdf_parallelism"])
	if err != nil {
		return models.VaultMetaModel{}, fmt.Errorf("map vault meta kdf_parallelism: %w", err)
	}

	return models.VaultMetaModel{
		ID:             1,
		VaultName:      row["vault_name"].(string),
		KdfAlgo:        row["kdf_algo"].(string),
		KdfSalt:        row["kdf_salt"].([]byte),
		KdfTimeCost:    kdfTimeCost,
		KdfMemoryCost:  kdfMemoryCost,
		KdfParallelism: kdfParallelism,
		CreatedAt:      row["created_at"].(time.Time),
		UpdatedAt:      row["updated_at"].(time.Time),
	}, nil
}

func (v *VaultRepositoryImpl) VaultMetaExists(ctx context.Context) (bool, error) {
	rows, err := v.db.QueryContext(ctx, `
		SELECT vault_name FROM vault_meta WHERE id = 1
	`)
	if err != nil {
		return false, fmt.Errorf("query vault meta exists: %w", err)
	}
	return len(rows) > 0, nil
}

func (v *VaultRepositoryImpl) SaveKeyCheck(ctx context.Context, keyCheck models.VaultKeyCheckModel) error {
	exists, err := v.vaultKeyCheckExists(ctx)
	if err != nil {
		return err
	}
	if exists {
		return pkerror.ErrVaultKeyCheckAlreadyExists
	}
	if err := v.db.ExecContext(ctx, `
		INSERT INTO vault_key_check (nonce, cipher_text,created_at)
		VALUES (?,?,?)
	`, keyCheck.Nonce, keyCheck.CipherText, keyCheck.CreatedAt); err != nil {
		return fmt.Errorf("insert vault_meta failed: %w", err)
	}
	return nil
}

func (v *VaultRepositoryImpl) KeyCheck(ctx context.Context) (models.VaultKeyCheckModel, error) {
	rows, err := v.db.QueryContext(ctx, `
		SELECT * FROM vault_key_check WHERE id = 1
	`)
	if err != nil {
		return models.VaultKeyCheckModel{}, fmt.Errorf("query vault key check: %w", err)
	}
	if len(rows) == 0 {
		return models.VaultKeyCheckModel{}, pkerror.ErrVaultKeyCheckNotFound
	}
	row := rows[0]
	return models.VaultKeyCheckModel{
		ID:         1,
		Nonce:      row["nonce"].([]byte),
		CipherText: row["cipher_text"].([]byte),
		CreatedAt:  row["created_at"].(time.Time),
	}, nil
}

func (v *VaultRepositoryImpl) SaveRecovery(ctx context.Context, recovery models.VaultRecoveryModel) error {
	exists, err := v.vaultRecoveryExists(ctx)
	if err != nil {
		return err
	}
	if exists {
		return pkerror.ErrVaultRecoveryAlreadyExists
	}
	if err := v.db.ExecContext(ctx, `
		INSERT INTO vault_recovery (
			kdf_algo,
			kdf_salt,
			kdf_time_cost,
			kdf_memory_cost,
			kdf_parallelism,
			nonce,
			cipher_text,
			created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, recovery.KdfAlgo,
		recovery.KdfSalt,
		recovery.KdfTimeCost,
		recovery.KdfMemoryCost,
		recovery.KdfParallelism,
		recovery.Nonce,
		recovery.CipherText,
		recovery.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert vault recovery failed: %w", err)
	}
	return nil
}

func (v *VaultRepositoryImpl) Recovery(ctx context.Context) (models.VaultRecoveryModel, error) {
	rows, err := v.db.QueryContext(ctx, `
		SELECT * FROM vault_recovery WHERE id = 1
	`)
	if err != nil {
		return models.VaultRecoveryModel{}, fmt.Errorf("query vault recovery: %w", err)
	}
	if len(rows) == 0 {
		return models.VaultRecoveryModel{}, pkerror.ErrVaultRecoveryNotFound
	}

	return mapVaultRecoveryModel(rows[0])
}

func (v *VaultRepositoryImpl) vaultKeyCheckExists(ctx context.Context) (bool, error) {
	rows, err := v.db.QueryContext(ctx, `
		SELECT nonce FROM vault_key_check WHERE id = 1
	`)
	if err != nil {
		return false, fmt.Errorf("query vault key check exists: %w", err)
	}
	return len(rows) > 0, nil
}

func (v *VaultRepositoryImpl) vaultRecoveryExists(ctx context.Context) (bool, error) {
	rows, err := v.db.QueryContext(ctx, `
		SELECT nonce FROM vault_recovery WHERE id = 1
	`)
	if err != nil {
		return false, fmt.Errorf("query vault recovery exists: %w", err)
	}
	return len(rows) > 0, nil
}

func mapVaultRecoveryModel(row map[string]any) (models.VaultRecoveryModel, error) {
	kdfTimeCost, err := sqliteInt(row["kdf_time_cost"])
	if err != nil {
		return models.VaultRecoveryModel{}, fmt.Errorf("map vault recovery kdf_time_cost: %w", err)
	}
	kdfMemoryCost, err := sqliteInt(row["kdf_memory_cost"])
	if err != nil {
		return models.VaultRecoveryModel{}, fmt.Errorf("map vault recovery kdf_memory_cost: %w", err)
	}
	kdfParallelism, err := sqliteInt(row["kdf_parallelism"])
	if err != nil {
		return models.VaultRecoveryModel{}, fmt.Errorf("map vault recovery kdf_parallelism: %w", err)
	}

	return models.VaultRecoveryModel{
		ID:             1,
		KdfAlgo:        row["kdf_algo"].(string),
		KdfSalt:        row["kdf_salt"].([]byte),
		KdfTimeCost:    kdfTimeCost,
		KdfMemoryCost:  kdfMemoryCost,
		KdfParallelism: kdfParallelism,
		Nonce:          row["nonce"].([]byte),
		CipherText:     row["cipher_text"].([]byte),
		CreatedAt:      row["created_at"].(time.Time),
	}, nil
}

func (i *ItemRepositoryImpl) Create(ctx context.Context, item models.ItemModel) error {
	if err := i.db.ExecContext(ctx, `
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
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	`,
		item.ID,
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
		return fmt.Errorf("insert items failed: %w", err)
	}
	return nil
}

func (i *ItemRepositoryImpl) Update(ctx context.Context, item models.ItemModel) error {
	if _, err := i.GetByID(ctx, item.ID); err != nil {
		return fmt.Errorf("update item failed: %w", err)
	}
	if err := i.db.ExecContext(ctx, `
		UPDATE items
		SET 
			title = ?,
			username_enc = ?,
			password_enc = ?,
			website_enc = ?,
			notes_enc = ?,
			nonce_username = ?,
			nonce_password = ?,
			nonce_website = ?,
			nonce_notes = ?,
			category = ?,
			favorite = ?,
			updated_at = ?
		WHERE id = ? 
	`,
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
		item.UpdatedAt,
		item.ID,
	); err != nil {
		return fmt.Errorf("update item: %w", err)
	}
	return nil
}

func (i *ItemRepositoryImpl) Delete(ctx context.Context, id string) error {
	err := i.db.ExecContext(ctx, `
		DELETE FROM items WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("Delete items: %w", err)
	}
	return nil
}

func (i *ItemRepositoryImpl) GetByID(ctx context.Context, id string) (models.ItemModel, error) {
	rows, err := i.db.QueryContext(ctx, `
		SELECT * FROM items WHERE id = ?
	`, id)
	if err != nil {
		return models.ItemModel{}, fmt.Errorf("query item by id: %w", err)
	}
	if len(rows) == 0 {
		return models.ItemModel{}, pkerror.ErrItemsNotFound
	}
	return mapItemModel(rows[0])
}

func (i *ItemRepositoryImpl) List(ctx context.Context, filter models.ItemListFilter) ([]models.ItemModel, error) {
	query := strings.Builder{}
	query.WriteString(`
		SELECT DISTINCT items.* FROM items
	`)

	args := make([]any, 0)
	conditions := make([]string, 0)
	if filter.TagID != "" {
		query.WriteString(`
			INNER JOIN item_tags ON item_tags.item_id = items.id
		`)
		conditions = append(conditions, "item_tags.tag_id = ?")
		args = append(args, filter.TagID)
	}

	if filter.Keyword != "" {
		conditions = append(conditions, "items.title LIKE ?")
		args = append(args, "%"+filter.Keyword+"%")
	}

	if filter.Category != "" {
		conditions = append(conditions, "items.category = ?")
		args = append(args, filter.Category)
	}

	if filter.Favorite != nil {
		conditions = append(conditions, "items.favorite = ?")
		args = append(args, sqliteBoolInt(*filter.Favorite))
	}

	if len(conditions) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(conditions, " AND "))
	}

	query.WriteString(" ORDER BY items.updated_at DESC")
	rows, err := i.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list items:%w", err)
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

func mapItemModel(row map[string]any) (models.ItemModel, error) {
	favorite, err := sqliteBool(row["favorite"])
	if err != nil {
		return models.ItemModel{}, fmt.Errorf("map item favorite: %w", err)
	}

	return models.ItemModel{
		ID:            row["id"].(string),
		Title:         row["title"].(string),
		UsernameEnc:   row["username_enc"].([]byte),
		PasswordEnc:   row["password_enc"].([]byte),
		WebsiteEnc:    row["website_enc"].([]byte),
		NotesEnc:      row["notes_enc"].([]byte),
		NonceUsername: row["nonce_username"].([]byte),
		NoncePassword: row["nonce_password"].([]byte),
		NonceWebsite:  row["nonce_website"].([]byte),
		NonceNotes:    row["nonce_notes"].([]byte),
		Category:      row["category"].(string),
		Favorite:      favorite,
		CreatedAt:     row["created_at"].(time.Time),
		UpdatedAt:     row["updated_at"].(time.Time),
	}, nil
}

func sqliteBool(value any) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case int32:
		return v != 0, nil
	case uint:
		return v != 0, nil
	case uint64:
		return v != 0, nil
	case uint32:
		return v != 0, nil
	default:
		return false, fmt.Errorf("unsupported bool storage type %T", value)
	}
}

func sqliteBoolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func sqliteInt(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case int32:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint64:
		return int(v), nil
	case uint32:
		return int(v), nil
	default:
		return 0, fmt.Errorf("unsupported int storage type %T", value)
	}
}

func (t *TagRepositoryImpl) Create(ctx context.Context, tag models.TagModel) error {
	_, err := t.GetByName(ctx, tag.Name)
	if err == nil {
		return pkerror.ErrTagAlreadyExists
	}
	if !errors.Is(err, pkerror.ErrTagNotFound) {
		return err
	}
	if err := t.db.ExecContext(ctx, `
		INSERT INTO tags (id, name, created_at)
		VALUES (?, ?, ?)
	`, tag.ID, tag.Name, tag.CreatedAt); err != nil {
		return fmt.Errorf("create tag: %w", err)
	}
	return nil
}

func (t *TagRepositoryImpl) GetByName(ctx context.Context, name string) (models.TagModel, error) {
	rows, err := t.db.QueryContext(ctx, `
		SELECT * FROM tags WHERE name = ?
	`, name)
	if err != nil {
		return models.TagModel{}, fmt.Errorf("query tag by name: %w", err)
	}
	if len(rows) == 0 {
		return models.TagModel{}, pkerror.ErrTagNotFound
	}
	row := rows[0]
	return models.TagModel{
		ID:        row["id"].(string),
		Name:      row["name"].(string),
		CreatedAt: row["created_at"].(time.Time),
	}, nil
}

func (t *TagRepositoryImpl) List(ctx context.Context) ([]models.TagModel, error) {
	rows, err := t.db.QueryContext(ctx, `
		SELECT * FROM tags ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
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

func (i *ItemTagRepositoryImpl) ReplaceItemTags(ctx context.Context, itemID string, tagIDs []string) error {
	if err := i.db.ExecContext(ctx, `
	DELETE FROM item_tags WHERE item_id = ?
	`, itemID); err != nil {
		return fmt.Errorf("delete old item tags: %w", err)
	}

	for _, tagID := range tagIDs {
		if err := i.db.ExecContext(ctx, `
		INSERT INTO item_tags (item_id, tag_id)
		VALUES (?, ?)
		`, itemID, tagID); err != nil {
			return fmt.Errorf("insert item tag failed: %w", err)
		}
	}
	return nil
}

func (i *ItemTagRepositoryImpl) GetTagsIDsByItemID(ctx context.Context, itemID string) ([]string, error) {
	rows, err := i.db.QueryContext(ctx, `
	SELECT * FROM item_tags WHERE item_id = ?
	`, itemID)
	if err != nil {
		return nil, fmt.Errorf("get item tag ids: %w", err)
	}
	if len(rows) == 0 {
		return []string{}, nil
	}
	tagIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		tagIDs = append(tagIDs, row["tag_id"].(string))
	}
	return tagIDs, nil
}

func (t *TxManagerImpl) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return pkerror.ErrTransactionUnsupported
}
