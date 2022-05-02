package ferr

type ErrorType string

const (
	// ETGeneric is for errors that do not fall into any other category
	ETGeneric = "generic"

	// ETValidation indicates that a piece of data is invalid, either user provided, or provided by the programmer
	ETValidation = "validation"

	// ETNetwork Indicates an error that occurred due to a networking issue. eg. failed to connect, failed to resolve host, etc...
	ETNetwork = "network"

	// ETSystem An error has occurred in the underlying system. eg. out of memory, disk full, etc...
	ETSystem = "system"

	// ETTemporal an error has occurred in temporal
	ETTemporal = "temporal"

	// ETAuth an error occurred during authentication, the requester failed to authenticate
	ETAuth = "auth"

	// ETDatabase an error occurred during a database operation
	ETDatabase = "database"

	// ETThirdPartySystem Any error originating from a system not controlled by Datomar (database is not a third party system)
	ETThirdPartySystem = "third_party"

	// ETPermissions an error caused by a user attempting to perform an action that they do not have permissions for
	ETPermissions = "permissions"
)

func (et ErrorType) String() string {
	return string(et)
}