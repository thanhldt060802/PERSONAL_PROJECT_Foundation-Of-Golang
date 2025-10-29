package dto

type RunTaskRequest struct {
	Body struct {
		TaskId int64 `json:"task_id" required:"true" doc:"Task id will send to Worker to process."`
	}
}
