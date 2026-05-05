package service

import (
	"context"

	"passwordkeeper/internal/domain"
)

type ItemService interface {
	Create(ctx context.Context, input domain.CreateItemInput) (domain.Item, error)
	Update(ctx context.Context, input domain.UpdateItemInput) (domain.Item, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (domain.Item, error)
	GetList(ctx context.Context, filter domain.ListItemsFilter) ([]domain.Item, error)
}

type VaultService interface {
	Init(ctx context.Context, input domain.InitVaultInput) (domain.VaultMeta, error)
	IsInitialized(ctx context.Context) (bool, error)
	Unlock(ctx context.Context, input domain.UnlockVaultInput) error
	Recover(ctx context.Context, input domain.RecoverVaultInput) error
	ChangeMasterPassword(ctx context.Context, input domain.ChangeMasterPasswordInput) (domain.VaultMeta, error)
	Lock(ctx context.Context) error
	GetMeta(ctx context.Context) (domain.VaultMeta, error)
}

type VaultSession interface {
	SetMasterKey(key []byte)
	GetMasterKey() ([]byte, bool)
	Clear()
	IsUnlocked() bool
}
