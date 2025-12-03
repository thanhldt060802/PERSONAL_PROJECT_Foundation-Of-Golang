package dto

type DispatchTaskRequest struct {
	Body struct {
		WorkerName string `json:"worker_name" required:"true" doc:"Name of Worker will receive task."`
		TaskId     int64  `json:"task_id" required:"true" doc:"Task id will send to Worker to process."`
	}
}

type RunTaskRequest struct {
	Body struct {
		TaskId int64 `json:"task_id" required:"true" doc:"Task id will send to Worker to process."`
	}
}

type RunTaskListRequest struct {
	Body struct {
		TaskIdList []int64 `json:"task_id_list" required:"true" doc:"Task id list will distribute to Worker to process."`
	}
}
