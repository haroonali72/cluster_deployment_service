package models

type Type string

const (
	Existing Type = "existing"
	New      Type = "new"
)

type Cloud string

const (
	AWS   Cloud = "aws"
	Azure Cloud = "azure"
	GCP   Cloud = "gcp"
)

type KeyType string

const (
	NEWKey  KeyType = "new"
	CPKey   KeyType = "cp"
	AWSKey  KeyType = "aws"
	USERKey KeyType = "user"
)

type OsDiskType string

const (
	StandardHDD OsDiskType = "standard hdd"
	StandardSSD OsDiskType = "standard ssd"
	PremiumSSD  OsDiskType = "premium ssd"
)
