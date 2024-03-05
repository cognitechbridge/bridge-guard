package core

type AppResult struct {
	Ok     bool        `json:"ok,omitempty"`
	Err    error       `json:"err,omitempty"`
	Result interface{} `json:"result,omitempty"`
}

func AppErrorResult(err error) AppResult {
	return AppResult{
		Ok:     false,
		Err:    err,
		Result: nil,
	}
}

func AppOkResult() AppResult {
	return AppResult{
		Ok:     true,
		Err:    nil,
		Result: nil,
	}
}
