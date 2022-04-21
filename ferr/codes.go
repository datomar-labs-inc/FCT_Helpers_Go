package ferr

type Code int

const (
	CodeUnknown = Code(iota)
	CodeWrapped
	CodeDBNotConnected
	CodeDBNoRows
	CodeUserExists
	CodeTimeout
	CodePanic
	CodeInvalidLoginDetails
)
