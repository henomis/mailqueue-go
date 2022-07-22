package restmodel

type Response struct {
	Status int         `json:"status"`
	Error  string      `json:"error,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

type EmailID struct {
	ID string `json:"id"`
}
