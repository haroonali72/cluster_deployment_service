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
	NEWKey   KeyType = "new"
	CPKey    KeyType = "cp"
	AWSKey   KeyType = "aws"
	AZUREKey KeyType = "azure"
	USERKey  KeyType = "user"
)

type OsDiskType string

const (
	StandardHDD OsDiskType = "standard hdd"
	StandardSSD OsDiskType = "standard ssd"
	PremiumSSD  OsDiskType = "premium ssd"
)

type GCPDiskType string

const (
	PdStandard GCPDiskType = "pd-standard"
	PdSSD      GCPDiskType = "pd-ssd"
)

type CredentialsType string

const (
	Password CredentialsType = "password"
	SSHKey   CredentialsType = "key"
)

type Logger string

const (
	//////logging///////////
	Backend_Logging Logger = "backend-logging"
	Audit_Trails    Logger = "audit-trails"

	AUDIT_TRAIL_ENDPOINT  = "elephant/api/v1/audit/store/"
	LOGGING_LEVEL_INFO    = "info"
	LOGGING_LEVEL_ERROR   = "error"
	LOGGING_LEVEL_WARNING = "warning"
)
