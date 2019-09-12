package models

type Type string

const (
	Existing Type = "existing"
	New      Type = "new"
)

type RequestType string

const (
	POST RequestType = "post"
	PUT  RequestType = "put"
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
const (
	ProjectGetEndpoint = "/raccoon/projects/{projectId}"
	WeaselGetEndpoint  = "/weasel/network/{cloud}/{projectId}"
)

const (
	VaultEndpoint = "/robin/api/v1"

	VaultGetKeyURI     = "/template/sshKey/{cloud}/{keyName}"
	VaultGetAllKeysURI = "/template/sshKey/{cloud}"
	VaultGetProfileURI = "/template/{cloud}/credentials/{profileId}"
	VaultCreateKeyURI  = "/template/sshKey"
)

const (
	RbacEndpoint = "/security/api/rbac/"

	RbacListURI     = "list"
	RbacAccessURI   = "allowed"
	RbacEvaluateURI = "evaluate"
	RbacInfoURI     = "token/info"
	RbacPolicyURI   = "policy"
)

type Logger string

const (
	//////logging///////////
	Backend_Logging Logger = "backend-logging"
	Audit_Trails    Logger = "audit-trails"

	LOGGING_LEVEL_INFO    = "info"
	LOGGING_LEVEL_ERROR   = "error"
	LOGGING_LEVEL_WARNING = "warning"
)

const (
	LoggingEndpoint = "/elephant/api/v1/"

	BackEndLoggingURI    = "backend/logging"
	FrontEndLoggingURI   = "frontend/logging"
	AuditTrailLoggingURI = "audit/store"
)
