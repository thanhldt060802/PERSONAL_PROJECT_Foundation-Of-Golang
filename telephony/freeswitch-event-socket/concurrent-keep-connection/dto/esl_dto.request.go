package dto

type OpenConnectionRequest struct {
	Body struct {
		WorkerName string `json:"worker_name" required:"true" doc:"Name of Worker will alias for connection."`
	}
}

type CloseConnectionRequest struct {
	Body struct {
		WorkerName string `json:"worker_name" required:"true" doc:"Name of Worker will alias for connection."`
	}
}

type RunCampaignListRequest struct {
	Body struct {
		CampaignList      []Campaign `json:"campaign_list" required:"true" doc:"List of campaign will send to Worker to process."`
		SummaryReportChan chan SummaryReport
	}
}
