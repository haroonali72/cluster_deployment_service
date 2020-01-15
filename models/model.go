package models

type Type string

const (
	Existing    Type = "existing"
	New         Type = "new"
	Created     Type = "created"
	Deploying   Type = "deploying"
	Terminating Type = "terminating"
)

type RequestType string

const (
	POST RequestType = "post"
	PUT  RequestType = "put"
)

type PoolRole string

const (
	Master PoolRole = "master"
	Slave  PoolRole = "slave"
)

type Role string

const (
	SuperUser Role = "Super-User"
	Admin     Role = "Admin"
)

type Cloud string

const (
	AWS   Cloud = "aws"
	Azure Cloud = "azure"
	GCP   Cloud = "gcp"
	DO    Cloud = "do"
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
	WoodpeckerEnpoint = "/agent/api/v1/clientconfig"
)
const (
	VaultEndpoint = "/robin/api/v1"

	VaultGetKeyURI     = "/template/sshKey/{cloud}/{region}/{keyName}"
	VaultGetAllKeysURI = "/template/sshKey/{cloud}/{region}"
	VaultGetProfileURI = "/template/{cloud}/credentials/{profileId}"
	VaultCreateKeyURI  = "/template/sshKey"
	VaultDeleteKeyURI  = "/template/sshKey/{cloudType}/{region}/{name}"
)

const (
	RbacEndpoint = "/security/api/rbac/"

	RbacListURI     = "list"
	RbacAccessURI   = "allowed"
	RbacEvaluateURI = "evaluate"
	RbacInfoURI     = "token/info"
	RbacPolicyURI   = "policy"
	RbacExtractURI  = "token/extract"
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

type Machine struct {
	InstanceType string `json: "instanceType" `
	Cores        int64  `json: "cores" `
}
type GCPMachine struct {
	InstanceType string  `json: "instanceType" `
	Cores        float64 `json: "cores" `
}

type Limits struct {
	CoreCount      int64 `json: "CoreCount" `
	DeveloperCount int64 `json: "DeveloperCount" `
	MeshCount      int64 `json: "MeshCount" `
	MeshSize       int64 `json: "MeshSize" `
}

const (
	NETWORK_CONTRIBUTOR_GUID = "4d97b98b-1d4f-4787-a291-c67834d212e7"
	VM_CONTRIBUTOR_GUID      = "9980e02c-c2be-4d73-94e8-173b1dc7cf3c"
	STORAGE_CONTRIBUTOR_GUID = "17d1049b-9a84-46fb-8f53-869881c3d3ab"
	AVERE_CONTRIBUTER_GUID   = "4f8fab4f-1852-4a58-a46a-8eaf358af14a"
)
