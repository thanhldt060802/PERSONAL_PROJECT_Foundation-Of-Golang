package dto

//
//
// Get response
// ################################################################################

type SuccessResponse struct {
	Body struct {
		Code    string `json:"code" example:"string"`
		Message string `json:"message" example:"string"`
	}
}

type BodyResponse[T any] struct {
	Body struct {
		Code    string `json:"code" example:"string"`
		Message string `json:"message" example:"string"`
		Data    T      `json:"data"`
	}
}

//
//
// Error response
// ################################################################################

type ErrorResponse struct {
	Code    string   `json:"code" example:"string"`
	Message string   `json:"message" example:"string"`
	Error_  string   `json:"error,omitempty"`
	Details []string `json:"details" example:"string"`
	Status  int      `json:"status" example:"1"`
}

func (err *ErrorResponse) Error() string {
	return err.Message
}

func (err *ErrorResponse) GetStatus() int {
	return err.Status
}
