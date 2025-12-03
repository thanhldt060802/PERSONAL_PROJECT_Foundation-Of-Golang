package types

type WorkerReport struct {
	WorkerName string
	DelayTime  string
	Cmd        string
}

type SummaryReport struct {
	Total         int
	Complete      int
	TotalTime     string
	WorkerReports []WorkerReport
}
