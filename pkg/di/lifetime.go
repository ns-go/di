package di

type Lifetime string

const (
	// Specifies that a single instance of the service will be created.
	Singleton Lifetime = "singleton"
	// Specifies that a new instance of the service will be created for each scope.
	Scoped Lifetime = "scoped"
	// Specifies that a new instance of the service will be created every time it is requested.
	Transient Lifetime = "transient"
)
