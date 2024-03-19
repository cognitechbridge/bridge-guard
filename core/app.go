package core

import "encoding/xml"

type AppResult struct {
	XMLName xml.Name    `json:"-" yaml:"-" xml:"Result"`
	Ok      bool        `json:"ok,omitempty" yaml:"ok,omitempty"`
	Err     error       `json:"err,omitempty" yaml:"err,omitempty"`
	Result  interface{} `json:"result,omitempty" yaml:"result,omitempty"`
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

func AppOkResultWithResult(result interface{}) AppResult {
	return AppResult{
		Ok:     true,
		Err:    nil,
		Result: result,
	}
}

// This is a struct for the repository status
type RepositroyStatus struct {
	IsValid  bool `json:"is_valid" yaml:"is_valid" xml:"is_valid"`
	IsJoined bool `json:"is_joined" yaml:"is_joined" xml:"is_joined"`
}
