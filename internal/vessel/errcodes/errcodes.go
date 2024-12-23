package errcodes

const (
	ErrInvalidMessage    = "400001"
	ErrInvalidAction     = "400002"
	ErrEngineCompile     = "400008"
	ErrEngineRun         = "400009"
	ErrRepositoryGetFile = "400010"
	ErrWorkerNotFound    = "400011"
	ErrWorkerStarting    = "400012"
)

const (
	ErrInvalidResponse   = "500001"
	ErrNewResponseFailed = "500002"
)
