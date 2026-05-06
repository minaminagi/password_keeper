package transport

import (
	"context"
	"time"

	"passwordkeeper/internal/domain"
	"passwordkeeper/internal/service"
)

type Handler struct {
	vaultService  service.VaultService
	itemService   service.ItemService
	backupService service.BackupService
}

func NewHandler(vaultService service.VaultService, itemService service.ItemService, backupService service.BackupService) *Handler {
	return &Handler{
		vaultService:  vaultService,
		itemService:   itemService,
		backupService: backupService,
	}
}

type InitVaultRequest struct {
	VaultName      string `json:"vault_name"`
	MasterPassword string `json:"master_password"`
}

type UnlockVaultRequest struct {
	MasterPassword string `json:"master_password"`
}

type RecoverVaultRequest struct {
	RecoveryCode string `json:"recovery_code"`
}

type ChangeMasterPasswordRequest struct {
	CurrentMasterPassword string `json:"current_master_password"`
	RecoveryCode          string `json:"recovery_code"`
	NewMasterPassword     string `json:"new_master_password"`
}

type CreateItemRequest struct {
	Title    string   `json:"title"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Website  string   `json:"website"`
	Notes    string   `json:"notes"`
	Category string   `json:"category"`
	Favorite bool     `json:"favorite"`
	Tags     []string `json:"tags"`
}

type UpdateItemRequest struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Website  string   `json:"website"`
	Notes    string   `json:"notes"`
	Category string   `json:"category"`
	Favorite bool     `json:"favorite"`
	Tags     []string `json:"tags"`
}

type ListItemsRequest struct {
	Keyword  string `json:"keyword"`
	Tag      string `json:"tag"`
	Favorite *bool  `json:"favorite"`
	Category string `json:"category"`
}

type ExportBackupRequest struct {
	ExportPassword string `json:"export_password"`
}

type ImportBackupRequest struct {
	ExportPassword string `json:"export_password"`
	CipherText     string `json:"cipher_text"`
}

type VaultMetaResponse struct {
	Name         string `json:"name"`
	RecoveryCode string `json:"recovery_code"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type ItemResponse struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Website   string   `json:"website"`
	Notes     string   `json:"notes"`
	Category  string   `json:"category"`
	Favorite  bool     `json:"favorite"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type ExportBackupResponse struct {
	CipherText string `json:"cipher_text"`
}

func (h *Handler) IsVaultInitialized(ctx context.Context) (bool, error) {
	return h.vaultService.IsInitialized(ctx)
}

func (h *Handler) InitVault(ctx context.Context, req InitVaultRequest) (VaultMetaResponse, error) {
	meta, err := h.vaultService.Init(ctx, domain.InitVaultInput{
		VaultName:      req.VaultName,
		MasterPassword: req.MasterPassword,
	})
	if err != nil {
		return VaultMetaResponse{}, err
	}

	return toVaultMetaResponse(meta), nil
}

func (h *Handler) UnlockVault(ctx context.Context, req UnlockVaultRequest) error {
	return h.vaultService.Unlock(ctx, domain.UnlockVaultInput{
		MasterPassword: req.MasterPassword,
	})
}

func (h *Handler) RecoverVault(ctx context.Context, req RecoverVaultRequest) error {
	return h.vaultService.Recover(ctx, domain.RecoverVaultInput{
		RecoveryCode: req.RecoveryCode,
	})
}

func (h *Handler) ChangeMasterPassword(ctx context.Context, req ChangeMasterPasswordRequest) (VaultMetaResponse, error) {
	meta, err := h.vaultService.ChangeMasterPassword(ctx, domain.ChangeMasterPasswordInput{
		CurrentMasterPassword: req.CurrentMasterPassword,
		RecoveryCode:          req.RecoveryCode,
		NewMasterPassword:     req.NewMasterPassword,
	})
	if err != nil {
		return VaultMetaResponse{}, err
	}
	return toVaultMetaResponse(meta), nil
}

func (h *Handler) LockVault(ctx context.Context) error {
	return h.vaultService.Lock(ctx)
}

func (h *Handler) GetVaultMeta(ctx context.Context) (VaultMetaResponse, error) {
	meta, err := h.vaultService.GetMeta(ctx)
	if err != nil {
		return VaultMetaResponse{}, err
	}

	return toVaultMetaResponse(meta), nil
}

func (h *Handler) ExportBackup(ctx context.Context, req ExportBackupRequest) (ExportBackupResponse, error) {
	cipherText, err := h.backupService.Export(ctx, domain.ExportBackupInput{
		ExportPassword: req.ExportPassword,
	})
	if err != nil {
		return ExportBackupResponse{}, err
	}
	return ExportBackupResponse{CipherText: cipherText}, nil
}

func (h *Handler) ImportBackup(ctx context.Context, req ImportBackupRequest) error {
	return h.backupService.Import(ctx, domain.ImportBackupInput{
		ExportPassword: req.ExportPassword,
		CipherText:     req.CipherText,
	})
}

func (h *Handler) CreateItem(ctx context.Context, req CreateItemRequest) (ItemResponse, error) {
	item, err := h.itemService.Create(ctx, domain.CreateItemInput{
		Title:    req.Title,
		Username: req.Username,
		Password: req.Password,
		Website:  req.Website,
		Notes:    req.Notes,
		Category: req.Category,
		Favorite: req.Favorite,
		Tags:     req.Tags,
	})
	if err != nil {
		return ItemResponse{}, err
	}

	return toItemResponse(item), nil
}

func (h *Handler) UpdateItem(ctx context.Context, req UpdateItemRequest) (ItemResponse, error) {
	item, err := h.itemService.Update(ctx, domain.UpdateItemInput{
		ID:       req.ID,
		Title:    req.Title,
		Username: req.Username,
		Password: req.Password,
		Website:  req.Website,
		Notes:    req.Notes,
		Category: req.Category,
		Favorite: req.Favorite,
		Tags:     req.Tags,
	})
	if err != nil {
		return ItemResponse{}, err
	}

	return toItemResponse(item), err
}

func (h *Handler) DeleteItem(ctx context.Context, id string) error {
	return h.itemService.Delete(ctx, id)
}

func (h *Handler) GetItem(ctx context.Context, id string) (ItemResponse, error) {
	item, err := h.itemService.GetByID(ctx, id)
	if err != nil {
		return ItemResponse{}, err
	}
	return toItemResponse(item), err
}

func (h *Handler) ListItems(ctx context.Context, req ListItemsRequest) ([]ItemResponse, error) {
	items, err := h.itemService.GetList(ctx, domain.ListItemsFilter{
		Keyword:  req.Keyword,
		Tag:      req.Tag,
		Favorite: req.Favorite,
		Category: req.Category,
	})
	if err != nil {
		return nil, err
	}

	resp := make([]ItemResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toItemResponse(item))
	}
	return resp, nil
}

func toVaultMetaResponse(meta domain.VaultMeta) VaultMetaResponse {
	return VaultMetaResponse{
		Name:         meta.Name,
		RecoveryCode: meta.RecoveryCode,
		CreatedAt:    meta.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    meta.UpdatedAt.Format(time.RFC3339),
	}
}

func toItemResponse(item domain.Item) ItemResponse {
	return ItemResponse{
		ID:        item.ID,
		Title:     item.Title,
		Username:  item.Username,
		Password:  item.Password,
		Website:   item.Website,
		Notes:     item.Notes,
		Category:  item.Category,
		Favorite:  item.Favorite,
		Tags:      item.Tags,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.Format(time.RFC3339),
	}
}
