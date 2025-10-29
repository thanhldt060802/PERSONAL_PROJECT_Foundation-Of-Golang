package dtos

type (
	PagingCommon struct {
		Limit  int `query:"limit" example:"10" default:"10" min:"1" max:"2999" doc:"Limit default is 10"`
		Offset int `query:"offset" example:"0" default:"0" min:"1" max:"2999" doc:"Offset default is 0"`
	}
)
