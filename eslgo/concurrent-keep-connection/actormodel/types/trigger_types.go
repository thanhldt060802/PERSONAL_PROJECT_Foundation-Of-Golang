package types

type DoProcessCall struct {
	Cmd              string
	WorkerReportChan chan WorkerReport
}

type DoSafetyTerminate struct{}

type CompleteMessage struct {
	WorkerName string
}
