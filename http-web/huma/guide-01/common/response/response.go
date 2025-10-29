package response

type PaginationBodyResponse[T any] struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Total   int    `json:"total"`
}

type PaginationResponse[T any] struct {
	Body PaginationBodyResponse[T]
}

type GenericBodyResponse[T any] struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type GenericResponse[T any] struct {
	Body GenericBodyResponse[T]
}

func OK[T any](data T, msg string) (res *GenericResponse[T]) {
	res = &GenericResponse[T]{
		Body: GenericBodyResponse[T]{
			Code:    "OK",
			Message: msg,
			Data:    data,
		},
	}
	return
}

func OK_Only(msg string) (res *GenericResponse[any]) {
	res = &GenericResponse[any]{
		Body: GenericBodyResponse[any]{
			Code:    "OK",
			Message: msg,
		},
	}
	return
}

func Pagination[T any](data T, total int, msg string) (res *PaginationResponse[T]) {
	res = &PaginationResponse[T]{
		Body: PaginationBodyResponse[T]{
			Code:    "OK",
			Message: msg,
			Data:    data,
			Total:   total,
		},
	}
	return
}
