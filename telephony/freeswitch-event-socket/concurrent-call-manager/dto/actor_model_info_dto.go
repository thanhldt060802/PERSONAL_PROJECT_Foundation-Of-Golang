package dto

type ExistedWorkers struct {
	WorkerNames []string `json:"worker_names"`
	Running     []string `json:"running"`
	Available   []string `json:"available"`
	Sleep       []string `json:"sleep"`
}

type WorkerReport struct {
	WorkerName string `json:"worker_name"`
	DelayTime  string `json:"delay_time"`
	Cmd        string `json:"cmd"`
}

type SummaryReport struct {
	Total         int            `json:"total"`
	Complete      int            `json:"complete"`
	TotalTime     string         `json:"total_time"`
	WorkerReports []WorkerReport `json:"worker_reports"`
}
