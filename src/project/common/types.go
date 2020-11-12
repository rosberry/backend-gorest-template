package common

type EmptyResponse struct {
	Result bool `json:"result"`
}

type DataResponse struct {
	Result bool        `json:"result"`
	Data   interface{} `json:"data"`
}

type Err struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Result bool `json:"result"`
	Error  Err  `json:"error"`
}
