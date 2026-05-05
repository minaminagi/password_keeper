package repo

import (
	"context"

	"passwordkeeper/internal/repo/models"
)

type VaultRepository interface {
	CreateMeta(ctx context.Context, meta models.VaultMetaModel) error
	UpdateMeta(ctx context.Context, meta models.VaultMetaModel) error
	Meta(ctx context.Context) (models.VaultMetaModel, error)
	VaultMetaExists(ctx context.Context) (bool, error)
	SaveKeyCheck(ctx context.Context, keyCheck models.VaultKeyCheckModel) error
	UpdateKeyCheck(ctx context.Context, keyCheck models.VaultKeyCheckModel) error
	KeyCheck(ctx context.Context) (models.VaultKeyCheckModel, error)
	SaveRecovery(ctx context.Context, recovery models.VaultRecoveryModel) error
	UpdateRecovery(ctx context.Context, recovery models.VaultRecoveryModel) error
	Recovery(ctx context.Context) (models.VaultRecoveryModel, error)
}

type ItemRepository interface {
	Create(ctx context.Context, item models.ItemModel) error
	Update(ctx context.Context, item models.ItemModel) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (models.ItemModel, error)
	List(ctx context.Context, filter models.ItemListFilter) ([]models.ItemModel, error)
}

type TagRepository interface {
	Create(ctx context.Context, tag models.TagModel) error
	GetByName(ctx context.Context, name string) (models.TagModel, error)
	List(ctx context.Context) ([]models.TagModel, error)
}

type ItemTagRepository interface {
	ReplaceItemTags(ctx context.Context, itemID string, tagIDs []string) error
	GetTagsIDsByItemID(ctx context.Context, itemID string) ([]string, error)
}

type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
