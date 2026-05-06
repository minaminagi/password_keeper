package domain

type ExportBackupInput struct {
	ExportPassword string
}

type ImportBackupInput struct {
	ExportPassword string
	CipherText     string
}
