package dto

type ExistedWorkers struct {
	WorkerNames []string `json:"worker_names"`
	Running     []string `json:"running"`
	Available   []string `json:"available"`
}
