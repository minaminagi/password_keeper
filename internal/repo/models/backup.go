package models

import "time"

type VaultBackupModel struct {
	Version    int
	ExportedAt time.Time
	Meta       VaultMetaModel
	KeyCheck   VaultKeyCheckModel
	Recovery   VaultRecoveryModel
	Items      []ItemModel
	Tags       []TagModel
	ItemTags   []ItemTagModel
}
