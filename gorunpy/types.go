package gorunpy

const (
	ExitCodeSuccess      = 0
	ExitCodeHandledError = 1
	ExitCodeCrash        = 2
)

type Request struct {
	Function string         `json:"function"`
	Args     map[string]any `json:"args"`
}

type Response struct {
	OK     bool         `json:"ok"`
	Result *ResultValue `json:"result,omitempty"`
}

type ResultValue struct {
	Value any `json:"value"`
}

type ErrorResponse struct {
	OK    bool         `json:"ok"`
	Error *ErrorDetail `json:"error,omitempty"`
}

type ErrorDetail struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}
