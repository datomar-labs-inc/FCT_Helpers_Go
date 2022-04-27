package ferr

type Code int

const (
	CodeUnknown = Code(iota)
	CodeWrapped
	CodeDBNotConnected
	CodeDBNoRows
	CodeAccountExists
	CodeTimeout
	CodePanic
	CodeInvalidLoginDetails
	CodeFlowCompleted
	CodeFlowFailed
	CodeMissingPermissions
	CodeNotAuthenticated
)
