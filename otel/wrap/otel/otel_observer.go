package otel

type ObserverEndPointConfig struct {
	ServiceName  string
	Host         string
	Port         int
	LocalLogFile string
}

func NewOtelObserver(config *ObserverEndPointConfig) func() {
	shutdownTracer := initTracer(config)
	shutdownLogger := initLogger(config)

	return func() {
		shutdownTracer()
		shutdownLogger()
	}
}
