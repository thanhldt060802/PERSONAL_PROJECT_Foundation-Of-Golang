package otel

type (
	// clientIPKeyType is a unique type used as a key in context.Context
	// to store and retrieve the client's IP address
	clientIPKeyType struct{}
)

var (
	// ClientIP is the context key for storing and retrieving
	// the client's IP address from request context
	ClientIP = clientIPKeyType{}
)
