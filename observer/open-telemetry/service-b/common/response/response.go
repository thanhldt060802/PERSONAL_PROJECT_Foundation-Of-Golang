package response

type PaginationBodyResponse[T any] struct {
	Code    string `json:"code" required:"false"`
	Message string `json:"message" required:"false"`
	Data    T      `json:"data,omitempty" required:"false"`
	Total   int    `json:"total" required:"false"`
	Offset  int    `json:"offset" required:"false"`
	Limit   int    `json:"limit" required:"false"`
}

type PaginationResponse[T any] struct {
	Body PaginationBodyResponse[T]
}

type GenericResponse[T any] struct {
	Body BodyResponse[T]
}

type BodyResponse[T any] struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

func Ok[T any](data T, msgs ...string) (res *GenericResponse[T]) {
	msg := "success"
	if len(msgs) > 0 {
		msg = msgs[0]
	}
	res = &GenericResponse[T]{
		Body: BodyResponse[T]{
			Code:    "OK",
			Message: msg,
			Data:    data,
		},
	}
	return
}

func OkOnly(msgs ...string) (res *GenericResponse[any]) {
	msg := "success"
	if len(msgs) > 0 {
		msg = msgs[0]
	}
	res = &GenericResponse[any]{
		Body: BodyResponse[any]{
			Code:    "OK",
			Message: msg,
		},
	}
	return
}

func Pagination[T any](data T, total int, offset int, limit int, msgs ...string) (res *PaginationResponse[T]) {
	msg := "success"
	if len(msgs) > 0 {
		msg = msgs[0]
	}
	res = &PaginationResponse[T]{
		Body: PaginationBodyResponse[T]{
			Code:    "OK",
			Message: msg,
			Data:    data,
			Total:   total,
			Offset:  offset,
			Limit:   limit,
		},
	}
	return
}
