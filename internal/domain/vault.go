package domain

import "time"

type VaultMeta struct {
	Name         string
	RecoveryCode string
	KDF          KDFParams
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type KDFParams struct {
	Algo        string
	Salt        []byte
	TimeCost    int
	MemoryCost  int
	Parallelism int
}

type InitVaultInput struct {
	VaultName      string
	MasterPassword string
}

type UnlockVaultInput struct {
	MasterPassword string
}

type RecoverVaultInput struct {
	RecoveryCode string
}
