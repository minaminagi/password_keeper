package pkerror

import "errors"

// VaultRepository
var (
	ErrVaultAlreadyExists = errors.New("vault meta already exists")
	ErrVaultNotFound      = errors.New("vault meta not found")

	ErrVaultKeyCheckAlreadyExists = errors.New("vault key check already exists")
	ErrVaultKeyCheckNotFound      = errors.New("vault key not found")
	ErrVaultRecoveryAlreadyExists = errors.New("vault recovery already exists")
	ErrVaultRecoveryNotFound      = errors.New("vault recovery not found")

	ErrItemsAlreadyExists = errors.New("items already exists")
	ErrItemsNotFound      = errors.New("items not found")

	ErrTagAlreadyExists = errors.New("tag already exists")
	ErrTagNotFound      = errors.New("tag not found")

	ErrItemTagsNotFound = errors.New("item tags not found")

	ErrTransactionUnsupported = errors.New("transaction is not supported by current sqlite service wrapper")

	ErrInvalidMasterPassword = errors.New("invalid master password")
	ErrInvalidRecoveryCode   = errors.New("invalid recovery code")

	ErrVaultLocked = errors.New("vault is locked")
)
