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
	CampaignName  string         `json:"campaign_name"`
	Total         int            `json:"total"`
	Complete      int            `json:"complete"`
	TotalTime     string         `json:"total_time"`
	WorkerReports []WorkerReport `json:"worker_reports"`
}

type Campaign struct {
	CampaignName string   `json:"campaign_name"`
	CmdList      []string `json:"cmd_list"`
}
