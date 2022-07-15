package ferr

type Code string

const (
	CodeUnknown             = Code("unknown")
	CodeWrapped             = "wrapped"
	CodeAccountExists       = "account_exists"
	CodeTimeout             = "timeout"
	CodePanic               = "panic"
	CodeInvalidLoginDetails = "invalid_login_details"
	CodeFlowCompleted       = "flow_already_completed"
	CodeFlowFailed          = "flow_failed"
	CodeMissingPermissions  = "missing_permissions"
	CodeNotAuthenticated    = "not_authenticated"
	CodeAccountDisabled     = "account_disabled"
	CodeMissingArgument     = "missing_argument"
	CodeNotFound            = "not_found"
	CodeInvalidInput        = "invalid_input"
	CodeInvalidAction       = "invalid_action"
	CodeOperationFailed     = "operation_failed"
)
