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

type DispatchCmdRequest struct {
	Body struct {
		WorkerName string `json:"worker_name" required:"true" doc:"Name of Worker will receive conmmand."`
		Cmd        string `json:"cmd" required:"true" doc:"Command will send to Worker to process."`
	}
}

type DispatchCmdListRequest struct {
	Body struct {
		CmdList []string `json:"cmd_list" required:"true" doc:"List of commands will send to Worker to process."`
	}
}
