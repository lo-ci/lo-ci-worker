package sandbox

type DoneResponse struct {
	ExitCode int64
}

type Container interface {
	Start() error
	Wait() (<-chan DoneResponse, <-chan error)
	WatchLogs() (<-chan string, error)
	Dispose()
}
