package utils

// ResponseJSON represents the standard API response structure
type ResponseJSON struct {
	Code  string      `json:"code"`
	Msg   string      `json:"msg"`
	Model interface{} `json:"model,omitempty"`
}

// ResponseWithModel creates a standard response with data
func ResponseWithModel(code string, msg string, model interface{}) ResponseJSON {
	return ResponseJSON{
		Code:  code,
		Msg:   msg,
		Model: model,
	}
}
