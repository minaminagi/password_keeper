package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"passwordkeeper/internal/crypto"
	"passwordkeeper/internal/domain"
	"passwordkeeper/internal/pkerror"
	"passwordkeeper/internal/repo"
	"passwordkeeper/internal/repo/models"

	"github.com/google/uuid"
)

type itemService struct {
	itemRepo    repo.ItemRepository
	tagRepo     repo.TagRepository
	itemTagRepo repo.ItemTagRepository
	encryptor   crypto.Encryptor
	session     VaultSession
}

func NewItemService(
	itemRepo repo.ItemRepository,
	tagRepo repo.TagRepository,
	itemTagRepo repo.ItemTagRepository,
	encryptor crypto.Encryptor,
	session VaultSession,
) ItemService {
	return &itemService{
		itemRepo:    itemRepo,
		tagRepo:     tagRepo,
		itemTagRepo: itemTagRepo,
		encryptor:   encryptor,
		session:     session,
	}
}

func validateCreateItemInput(input domain.CreateItemInput) error {
	if input.Title == "" {
		return fmt.Errorf("item title is required")
	}
	return nil
}

func (s *itemService) encryptFields(input domain.CreateItemInput, masterKey []byte) (models.ItemModel, error) {
	username, err := s.encryptor.Encrypt([]byte(input.Username), masterKey)
	if err != nil {
		return models.ItemModel{}, err
	}

	password, err := s.encryptor.Encrypt([]byte(input.Password), masterKey)
	if err != nil {
		return models.ItemModel{}, err
	}

	website, err := s.encryptor.Encrypt([]byte(input.Website), masterKey)
	if err != nil {
		return models.ItemModel{}, err
	}

	notes, err := s.encryptor.Encrypt([]byte(input.Notes), masterKey)
	if err != nil {
		return models.ItemModel{}, err
	}

	return models.ItemModel{
		Title:         input.Title,
		UsernameEnc:   username.CipherText,
		NonceUsername: username.Nonce,
		PasswordEnc:   password.CipherText,
		NoncePassword: password.Nonce,
		WebsiteEnc:    website.CipherText,
		NonceWebsite:  website.Nonce,
		NotesEnc:      notes.CipherText,
		NonceNotes:    notes.Nonce,
		Category:      input.Category,
		Favorite:      input.Favorite,
	}, nil
}

func (s *itemService) ensureTags(ctx context.Context, names []string) ([]string, error) {
	tagIDs := make([]string, 0, len(names))
	seen := make(map[string]struct{}, len(names))

	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}

		tag, err := s.tagRepo.GetByName(ctx, name)
		if err == nil {
			tagIDs = append(tagIDs, tag.ID)
			continue
		}
		if !errors.Is(err, pkerror.ErrTagNotFound) {
			return nil, err
		}

		tag = models.TagModel{
			ID:        uuid.NewString(),
			Name:      name,
			CreatedAt: time.Now(),
		}
		if err := s.tagRepo.Create(ctx, tag); err != nil {
			return nil, err
		}
		tagIDs = append(tagIDs, tag.ID)
	}
	return tagIDs, nil
}

func (s *itemService) Create(ctx context.Context, input domain.CreateItemInput) (domain.Item, error) {
	masterKey, ok := s.session.GetMasterKey()
	if !ok {
		return domain.Item{}, pkerror.ErrVaultLocked
	}

	if err := validateCreateItemInput(input); err != nil {
		return domain.Item{}, err
	}

	itemModel, err := s.encryptFields(input, masterKey)
	if err != nil {
		return domain.Item{}, err
	}

	now := time.Now()
	itemModel.ID = uuid.NewString()
	itemModel.CreatedAt = now
	itemModel.UpdatedAt = now
	if itemModel.Category == "" {
		itemModel.Category = "login"
	}

	if err = s.itemRepo.Create(ctx, itemModel); err != nil {
		return domain.Item{}, err
	}

	tagIDs, err := s.ensureTags(ctx, input.Tags)
	if err != nil {
		return domain.Item{}, err
	}

	if err := s.itemTagRepo.ReplaceItemTags(ctx, itemModel.ID, tagIDs); err != nil {
		return domain.Item{}, err
	}

	return domain.Item{
		ID:        itemModel.ID,
		Title:     input.Title,
		Username:  input.Username,
		Password:  input.Password,
		Website:   input.Website,
		Notes:     input.Notes,
		Category:  itemModel.Category,
		Favorite:  input.Favorite,
		Tags:      input.Tags,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *itemService) decryptItem(ctx context.Context, item models.ItemModel, masterKey []byte) (domain.Item, error) {
	username, err := s.encryptor.Decrypt(item.UsernameEnc, item.NonceUsername, masterKey)
	if err != nil {
		return domain.Item{}, err
	}

	password, err := s.encryptor.Decrypt(item.PasswordEnc, item.NoncePassword, masterKey)
	if err != nil {
		return domain.Item{}, err
	}

	website, err := s.encryptor.Decrypt(item.WebsiteEnc, item.NonceWebsite, masterKey)
	if err != nil {
		return domain.Item{}, err
	}

	notes, err := s.encryptor.Decrypt(item.NotesEnc, item.NonceNotes, masterKey)
	if err != nil {
		return domain.Item{}, err
	}

	tagIDs, err := s.itemTagRepo.GetTagsIDsByItemID(ctx, item.ID)
	if err != nil {
		return domain.Item{}, err
	}
	tagNames, err := s.tagNamesByIDs(ctx, tagIDs)
	if err != nil {
		return domain.Item{}, err
	}

	return domain.Item{
		ID:        item.ID,
		Title:     item.Title,
		Username:  string(username),
		Password:  string(password),
		Website:   string(website),
		Notes:     string(notes),
		Category:  item.Category,
		Favorite:  item.Favorite,
		Tags:      tagNames,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}, nil
}

func (s *itemService) tagNamesByIDs(ctx context.Context, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}

	tags, err := s.tagRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	namesByID := make(map[string]string, len(tags))
	for _, tag := range tags {
		namesByID[tag.ID] = tag.Name
	}

	names := make([]string, 0, len(ids))
	for _, id := range ids {
		name, ok := namesByID[id]
		if !ok {
			continue
		}
		names = append(names, name)
	}
	return names, nil
}

func (s *itemService) GetByID(ctx context.Context, id string) (domain.Item, error) {
	masterKey, ok := s.session.GetMasterKey()
	if !ok {
		return domain.Item{}, pkerror.ErrVaultLocked
	}

	item, err := s.itemRepo.GetByID(ctx, id)
	if err != nil {
		return domain.Item{}, err
	}

	return s.decryptItem(ctx, item, masterKey)
}

func (s *itemService) GetList(ctx context.Context, filter domain.ListItemsFilter) ([]domain.Item, error) {
	masterKey, ok := s.session.GetMasterKey()
	if !ok {
		return nil, pkerror.ErrVaultLocked
	}

	items, err := s.itemRepo.List(ctx, models.ItemListFilter{
		Keyword:  filter.Keyword,
		TagID:    filter.Tag,
		Favorite: filter.Favorite,
		Category: filter.Category,
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.Item, 0, len(items))
	for _, item := range items {
		decrypted, err := s.decryptItem(ctx, item, masterKey)
		if err != nil {
			return nil, err
		}
		result = append(result, decrypted)
	}
	return result, nil
}

func (s *itemService) Update(ctx context.Context, input domain.UpdateItemInput) (domain.Item, error) {
	masterKey, ok := s.session.GetMasterKey()
	if !ok {
		return domain.Item{}, pkerror.ErrVaultLocked
	}

	createInput := domain.CreateItemInput{
		Title:    input.Title,
		Username: input.Username,
		Password: input.Password,
		Website:  input.Website,
		Notes:    input.Notes,
		Category: input.Category,
		Favorite: input.Favorite,
		Tags:     input.Tags,
	}

	itemModel, err := s.encryptFields(createInput, masterKey)
	if err != nil {
		return domain.Item{}, err
	}

	now := time.Now()
	itemModel.ID = input.ID
	itemModel.UpdatedAt = now
	if itemModel.Category == "" {
		itemModel.Category = "login"
	}

	if err = s.itemRepo.Update(ctx, itemModel); err != nil {
		return domain.Item{}, err
	}

	tagIDs, err := s.ensureTags(ctx, input.Tags)
	if err != nil {
		return domain.Item{}, err
	}
	if err = s.itemTagRepo.ReplaceItemTags(ctx, input.ID, tagIDs); err != nil {
		return domain.Item{}, err
	}

	item, err := s.GetByID(ctx, input.ID)
	if err != nil {
		return domain.Item{}, err
	}

	return item, err
}

func (s *itemService) Delete(ctx context.Context, id string) error {
	if !s.session.IsUnlocked() {
		return pkerror.ErrVaultLocked
	}
	return s.itemRepo.Delete(ctx, id)
}
