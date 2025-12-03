package types

type GetExistedWorkersMessage struct {
	WorkerNamesChan chan []string
	RunningChan     chan []string
	AvailableChan   chan []string
}

type RunTaskMessage struct {
	TaskId int64
}
