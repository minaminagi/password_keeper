package models

import (
	"time"
)

// 保存保险库本身的信息，以及主密码派生密钥所需的参数
type VaultMetaModel struct {
	ID             int64
	VaultName      string
	KdfAlgo        string
	KdfSalt        []byte
	KdfTimeCost    int
	KdfMemoryCost  int
	KdfParallelism int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// 校验用户输入的主密码是否正确
type VaultKeyCheckModel struct {
	ID         int64
	Nonce      []byte
	CipherText []byte
	CreatedAt  time.Time
}

// 数据库存储模型，敏感字段以密文形式保存
type ItemModel struct {
	ID            string
	Title         string
	UsernameEnc   []byte
	PasswordEnc   []byte
	WebsiteEnc    []byte
	NotesEnc      []byte
	NonceUsername []byte
	NoncePassword []byte
	NonceWebsite  []byte
	NonceNotes    []byte
	Category      string
	Favorite      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// 保存标签本身的信息
type TagModel struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

// 表示密码项和标签之间的多对多关系
type ItemTagModel struct {
	ItemID string
	TagID  string
}

type ItemListFilter struct {
	Keyword  string
	TagID    string
	Favorite *bool
	Category string
}
