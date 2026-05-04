package service

import (
	"context"
	"errors"
	"sort"
	"testing"

	"passwordkeeper/internal/crypto"
	"passwordkeeper/internal/domain"
	"passwordkeeper/internal/pkerror"
	"passwordkeeper/internal/repo/models"
)

type memoryItemRepo struct {
	items map[string]models.ItemModel
}

func newMemoryItemRepo() *memoryItemRepo {
	return &memoryItemRepo{items: make(map[string]models.ItemModel)}
}

func (r *memoryItemRepo) Create(ctx context.Context, item models.ItemModel) error {
	r.items[item.ID] = item
	return nil
}

func (r *memoryItemRepo) Update(ctx context.Context, item models.ItemModel) error {
	if _, ok := r.items[item.ID]; !ok {
		return pkerror.ErrItemsNotFound
	}
	r.items[item.ID] = item
	return nil
}

func (r *memoryItemRepo) Delete(ctx context.Context, id string) error {
	delete(r.items, id)
	return nil
}

func (r *memoryItemRepo) GetByID(ctx context.Context, id string) (models.ItemModel, error) {
	item, ok := r.items[id]
	if !ok {
		return models.ItemModel{}, pkerror.ErrItemsNotFound
	}
	return item, nil
}

func (r *memoryItemRepo) List(ctx context.Context, filter models.ItemListFilter) ([]models.ItemModel, error) {
	items := make([]models.ItemModel, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}
	return items, nil
}

type memoryTagRepo struct {
	byName      map[string]models.TagModel
	createCalls int
}

func newMemoryTagRepo(tags ...models.TagModel) *memoryTagRepo {
	repo := &memoryTagRepo{byName: make(map[string]models.TagModel)}
	for _, tag := range tags {
		repo.byName[tag.Name] = tag
	}
	return repo
}

func (r *memoryTagRepo) Create(ctx context.Context, tag models.TagModel) error {
	if _, ok := r.byName[tag.Name]; ok {
		return pkerror.ErrTagAlreadyExists
	}
	r.createCalls++
	r.byName[tag.Name] = tag
	return nil
}

func (r *memoryTagRepo) GetByName(ctx context.Context, name string) (models.TagModel, error) {
	tag, ok := r.byName[name]
	if !ok {
		return models.TagModel{}, pkerror.ErrTagNotFound
	}
	return tag, nil
}

func (r *memoryTagRepo) List(ctx context.Context) ([]models.TagModel, error) {
	tags := make([]models.TagModel, 0, len(r.byName))
	for _, tag := range r.byName {
		tags = append(tags, tag)
	}
	return tags, nil
}

type memoryItemTagRepo struct {
	byItem map[string][]string
}

func newMemoryItemTagRepo() *memoryItemTagRepo {
	return &memoryItemTagRepo{byItem: make(map[string][]string)}
}

func (r *memoryItemTagRepo) ReplaceItemTags(ctx context.Context, itemID string, tagIDs []string) error {
	for _, tagID := range tagIDs {
		if tagID == "" {
			return errors.New("empty tag id")
		}
	}
	r.byItem[itemID] = append([]string(nil), tagIDs...)
	return nil
}

func (r *memoryItemTagRepo) GetTagsIDsByItemID(ctx context.Context, itemID string) ([]string, error) {
	return append([]string(nil), r.byItem[itemID]...), nil
}

func newUnlockedItemService(
	itemRepo *memoryItemRepo,
	tagRepo *memoryTagRepo,
	itemTagRepo *memoryItemTagRepo,
) ItemService {
	session := NewMemoryVaultSession()
	session.SetMasterKey([]byte("12345678901234567890123456789012"))

	return NewItemService(
		itemRepo,
		tagRepo,
		itemTagRepo,
		crypto.NewAESGCMEncryptor(),
		session,
	)
}

func TestItemServiceCreateCreatesTagsAndAssociations(t *testing.T) {
	itemRepo := newMemoryItemRepo()
	tagRepo := newMemoryTagRepo()
	itemTagRepo := newMemoryItemTagRepo()
	svc := newUnlockedItemService(itemRepo, tagRepo, itemTagRepo)

	item, err := svc.Create(context.Background(), domain.CreateItemInput{
		Title:    "GitHub",
		Username: "octo",
		Password: "secret",
		Website:  "https://github.com",
		Notes:    "recovery codes are stored offline",
		Category: "dev",
		Favorite: true,
		Tags:     []string{"work", "personal", "work", " "},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if item.Notes != "recovery codes are stored offline" {
		t.Fatalf("Create returned Notes %q, want input notes", item.Notes)
	}
	if _, ok := tagRepo.byName["work"]; !ok {
		t.Fatalf("Create did not create work tag")
	}
	if _, ok := tagRepo.byName["personal"]; !ok {
		t.Fatalf("Create did not create personal tag")
	}
	if tagRepo.createCalls != 2 {
		t.Fatalf("Create created %d tags, want 2", tagRepo.createCalls)
	}

	tagIDs := append([]string(nil), itemTagRepo.byItem[item.ID]...)
	sort.Strings(tagIDs)
	wantTagIDs := []string{
		tagRepo.byName["personal"].ID,
		tagRepo.byName["work"].ID,
	}
	sort.Strings(wantTagIDs)
	if len(tagIDs) != len(wantTagIDs) {
		t.Fatalf("ReplaceItemTags received %v, want %v", tagIDs, wantTagIDs)
	}
	for idx := range tagIDs {
		if tagIDs[idx] != wantTagIDs[idx] {
			t.Fatalf("ReplaceItemTags received %v, want %v", tagIDs, wantTagIDs)
		}
	}
}

func TestItemServiceCreateReusesExistingTag(t *testing.T) {
	itemRepo := newMemoryItemRepo()
	tagRepo := newMemoryTagRepo(models.TagModel{ID: "tag-work", Name: "work"})
	itemTagRepo := newMemoryItemTagRepo()
	svc := newUnlockedItemService(itemRepo, tagRepo, itemTagRepo)

	item, err := svc.Create(context.Background(), domain.CreateItemInput{
		Title: "GitLab",
		Tags:  []string{"work"},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if tagRepo.createCalls != 0 {
		t.Fatalf("Create created %d tags, want 0", tagRepo.createCalls)
	}
	tagIDs := itemTagRepo.byItem[item.ID]
	if len(tagIDs) != 1 || tagIDs[0] != "tag-work" {
		t.Fatalf("ReplaceItemTags received %v, want [tag-work]", tagIDs)
	}
}

func TestItemServiceGetByIDReturnsTagNames(t *testing.T) {
	itemRepo := newMemoryItemRepo()
	tagRepo := newMemoryTagRepo(
		models.TagModel{ID: "tag-work", Name: "work"},
		models.TagModel{ID: "tag-personal", Name: "personal"},
	)
	itemTagRepo := newMemoryItemTagRepo()
	svc := newUnlockedItemService(itemRepo, tagRepo, itemTagRepo)

	created, err := svc.Create(context.Background(), domain.CreateItemInput{
		Title:    "GitHub",
		Password: "secret",
		Tags:     []string{"work", "personal"},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	item, err := svc.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}

	sort.Strings(item.Tags)
	wantTags := []string{"personal", "work"}
	if len(item.Tags) != len(wantTags) {
		t.Fatalf("GetByID returned tags %v, want %v", item.Tags, wantTags)
	}
	for idx := range item.Tags {
		if item.Tags[idx] != wantTags[idx] {
			t.Fatalf("GetByID returned tags %v, want %v", item.Tags, wantTags)
		}
	}
	if item.Password != "secret" {
		t.Fatalf("GetByID returned password %q, want original password", item.Password)
	}
}
