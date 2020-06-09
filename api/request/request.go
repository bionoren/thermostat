package request

type ApiResponse struct {
	Code int
	Msg  string
}

func NewResponse(code int, msg string) ApiResponse {
	return ApiResponse{code, msg}
}
