package restmodel

type Status string

const (
	StatusSuccess = "success"
	StatusFail    = "fail"
	StatusError   = "error"
)

type ResposeStatus struct {
	Status Status `json:"status"`
}

type ResponseSuccess struct {
	Status `json:"status"`
	Data   interface{} `json:"data"`
}

type ResponseFail struct {
	Status `json:"status"`
	Data   interface{} `json:"data"`
}

type ResponseError struct {
	Status  `json:"status"`
	Message string `json:"message"`
}

func Success(data interface{}) *ResponseSuccess {
	return &ResponseSuccess{
		Status: StatusSuccess,
		Data:   data,
	}
}

func Fail(data interface{}) *ResponseFail {
	return &ResponseFail{
		Status: StatusFail,
		Data:   data,
	}
}

func Error(message string) *ResponseError {
	return &ResponseError{
		Status:  StatusError,
		Message: message,
	}
}
