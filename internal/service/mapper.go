package service

import (
	"passwordkeeper/internal/crypto"
	"passwordkeeper/internal/domain"
	"passwordkeeper/internal/repo/models"
)

func toCryptoKDFParams(params domain.KDFParams) crypto.KDFParams {
	return crypto.KDFParams{
		Algo:        params.Algo,
		Salt:        params.Salt,
		TimeCost:    params.TimeCost,
		MemoryCost:  params.MemoryCost,
		Parallelism: params.Parallelism,
		Keylength:   crypto.DefaultKeyLen,
	}
}

func toDomainVaultMeta(meta models.VaultMetaModel) domain.VaultMeta {
	return domain.VaultMeta{
		Name: meta.VaultName,
		KDF: domain.KDFParams{
			Algo:        meta.KdfAlgo,
			Salt:        meta.KdfSalt,
			TimeCost:    meta.KdfTimeCost,
			MemoryCost:  meta.KdfMemoryCost,
			Parallelism: meta.KdfParallelism,
		},
		CreatedAt: meta.CreatedAt,
		UpdatedAt: meta.UpdatedAt,
	}
}
