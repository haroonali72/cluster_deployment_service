package constants

type Logger string

const (
	//////logging///////////
	Backend_Logging Logger = "backend-logging"
	Audit_Trails    Logger = "auditTrails"

	AUDIT_TRAIL_ENDPOINT  = "elephant/api/v1/audit/store/"
	LOGGING_LEVEL_INFO    = "info"
	LOGGING_LEVEL_ERROR   = "error"
	LOGGING_LEVEL_WARNING = "warning"
)
