package types

type GetExistedWorkersMessage struct {
	WorkerNamesChan chan []string
	RunningChan     chan []string
	AvailableChan   chan []string
}

type DispatchTaskMessage struct {
	WorkerName string
	TaskId     int64
}

type RunTaskMessage struct {
	TaskId int64
}

type RunTaskListMessage struct {
	TaskIdList []int64
}
