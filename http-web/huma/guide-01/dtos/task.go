package dtos

type (
	// DTO for getting request with filtering
	GetsTaskFilter struct {
		SortBy       string `query:"sort_by" example:"created_at:desc" default:"created_at:desc" doc:"Sort by one or more fields separated by commas. For example: sort_by=name,created_at:desc will sort by name in ascending order, then by created_at in descending order."`
		TaskName     string `query:"task_name" example:"Some text" doc:"Search by task name."`
		State        string `query:"state_eq" example:"todo" enum:"todo,in progress,done" doc:"Filter by state."`
		Priority     string `query:"priority_eq" example:"high" enum:"low,medium,high" doc:"Filter by priority."`
		Progress     string `query:"progress" example:"60" pattern:"^(0|[1-9]\\d*)$" doc:"Search by progress equals this value."`
		ProgressGTE  string `query:"progress_gte" example:"50" pattern:"^(0|[1-9]\\d*)$" doc:"Search by progress greater than or equals this value."`
		ProgressLTE  string `query:"progress_lte" example:"90" pattern:"^(0|[1-9]\\d*)$" doc:"Search by progress less than or equals this value."`
		CreatedAtGTE string `query:"created_at_gte" example:"2025-03-18T00:00:00" pattern:"^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}$" doc:"Search leads created on or after this date and time, using the format YYYY-MM-ddTHH:mm:ss"`
		CreatedAtLTE string `query:"created_at_lte" example:"2025-03-18T23:59:59" pattern:"^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}$" doc:"Search leads with created until this value in format YYYY-MM-ddTHH:mm:ss"`
	}
	GetsTaskCustomFilter struct {
		SortBy       string   `query:"sort_by" example:"created_at:desc" default:"created_at:desc" doc:"Sort by one or more fields separated by commas. For example: sort_by=name,created_at:desc will sort by name in ascending order, then by created_at in descending order."`
		StateIn      []string `query:"state_in" example:"in progress,done" enum:"todo,in progress,done" doc:"Filter by list states."`
		CreatedAtGTE string   `query:"created_at_gte" example:"2025-03-18T00:00:00" pattern:"^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}$" doc:"Search leads created on or after this date and time, using the format YYYY-MM-ddTHH:mm:ss"`
		CreatedAtLTE string   `query:"created_at_lte" example:"2025-03-18T23:59:59" pattern:"^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}$" doc:"Search leads with created until this value in format YYYY-MM-ddTHH:mm:ss"`
	}
)

type (
	// DTO for creating request
	CreateTaskDTO struct {
		Password    *string `json:"password,omitempty" example:"Some text" doc:"Password."`
		TaskName    string  `json:"task_name" required:"true" example:"Some text" minLength:"1" doc:"Task name."`
		Description *string `json:"description,omitempty" example:"Some text" doc:"Description"`
		State       string  `json:"state" required:"true" enum:"todo,in progress,done" doc:"State."`
		Priority    string  `json:"priority" required:"true" enum:"low,medium,high" doc:"Priority."`
		Progress    int     `json:"progress" required:"true" minimum:"0" maximum:"100" doc:"Progress."`
	}
	CreateTaskCustomDTO struct {
		Password    *string `json:"password,omitempty" example:"Some text" doc:"Password."`
		TaskName    string  `json:"task_name" required:"true" example:"Some text" minLength:"1" doc:"Task name."`
		Description *string `json:"description,omitempty" example:"Some text" doc:"Description"`
	}

	// DTO for updating request
	UpdateTaskDTO struct {
		Password    string `json:"password" required:"true" example:"Some text" doc:"Password."`
		TaskName    string `json:"task_name" required:"true" example:"Some text" minLength:"1" doc:"Field 2 (private)."`
		Description string `json:"description" required:"true" example:"Some text" doc:"Description"`
		State       string `json:"state" required:"true" enum:"todo,in progress,done" doc:"State."`
		Priority    string `json:"priority" required:"true" enum:"low,medium,high" doc:"Priority."`
		Progress    int    `json:"progress" required:"true" minimum:"0" maximum:"100" doc:"Progress."`
	}
	PatchTaskDTO struct {
		Password    *string `json:"password,omitempty" example:"Some text" doc:"Password."`
		TaskName    *string `json:"task_name,omitempty" example:"Some text" minLength:"1" doc:"Field 2 (private)."`
		Description *string `json:"description,omitempty" example:"Some text" doc:"Description"`
		State       *string `json:"state,omitempty" enum:"todo,in progress,done" doc:"State."`
		Priority    *string `json:"priority,omitempty" enum:"low,medium,high" doc:"Priority."`
		Progress    *int    `json:"progress,omitempty" minimum:"0" maximum:"100" doc:"Progress."`
	}
	PatchTaskCustomDTO struct {
		State    string `json:"state" required:"true" enum:"todo,in progress,done" doc:"State."`
		Progress int    `json:"progress" required:"true" minimum:"0" maximum:"100" doc:"Progress."`
	}
)
