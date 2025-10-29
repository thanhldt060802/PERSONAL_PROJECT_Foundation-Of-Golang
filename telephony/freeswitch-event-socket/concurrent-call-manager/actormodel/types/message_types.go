package types

type GetExistedWorkersMessage struct {
	WorkerNamesChan chan []string
	RunningChan     chan []string
	AvailableChan   chan []string
	SleepChan       chan []string
}

type OpenConnectionMessage struct {
	WorkerName string
}

type CloseConnectionMessage struct {
	WorkerName string
}

type DispatchCmdMessage struct {
	WorkerName       string
	Cmd              string
	WorkerReportChan chan WorkerReport
}

type DispatchCmdListMessage struct {
	CmdList           []string
	SummaryReportChan chan SummaryReport
}
