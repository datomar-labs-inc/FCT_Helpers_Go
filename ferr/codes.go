package ferr

type Code string

const (
	CodeUnknown             = Code("unknown")
	CodeWrapped             = "wrapped"
	CodeAccountExists       = "accout_exists"
	CodeTimeout             = "timeout"
	CodePanic               = "panic"
	CodeInvalidLoginDetails = "invalid_login_details"
	CodeFlowCompleted       = "flow_already_completed"
	CodeFlowFailed          = "flow_failed"
	CodeMissingPermissions  = "missing_permissions"
	CodeNotAuthenticated    = "not_authenticated"
	CodeMissingArgument     = "missing_argument"
	CodeNotFound            = "not_found"
	CodeInvalidInput        = "invalid_input"
)
