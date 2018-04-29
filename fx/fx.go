package fx

// Function types specific to observable.Observable and its operators
type (
	// EmitFunc defines a emitting function for the Start operator.
	// It is used to wrap a blocking operation which will be emitted on
	// the source Observable at a point in time.
	EmitFunc func() interface{}

	// KeySelectFunc defines a selecting function for Distinct operator.
	KeySelectFunc func(interface{}) interface{}

	// CombineFunc defines a combining function for the CombineLatast operator.
	CombineFunc func([]interface{}) interface{}
)
