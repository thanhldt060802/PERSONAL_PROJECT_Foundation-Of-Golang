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

type RunCampaignListMessage struct {
	CampaignInfoList  []CampaignInfo
	SummaryReportChan chan SummaryReport
}
