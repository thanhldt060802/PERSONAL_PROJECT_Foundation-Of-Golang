package types

import "time"

type WorkerReport struct {
	WorkerName string
	DelayTime  string
	Cmd        string
}

type SummaryReport struct {
	CampaignName  string
	Total         int
	Complete      int
	TotalTime     string
	WorkerReports []WorkerReport
}

type CampaignInfo struct {
	CampaignName string

	CmdList  []string
	Index    int
	Complete int

	StartTime time.Time
	EndTime   time.Time

	WorkerReportChans []chan WorkerReport
}
