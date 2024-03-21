package core

import "encoding/xml"

// AppResult represents the result of an application operation.
type AppResult struct {
	XMLName xml.Name    `json:"-" yaml:"-" xml:"Result"`
	Ok      bool        `json:"ok,omitempty" yaml:"ok,omitempty"`
	Err     error       `json:"err,omitempty" yaml:"err,omitempty"`
	Result  interface{} `json:"result,omitempty" yaml:"result,omitempty"`
}

// NewAppResultWithError creates a new AppResult indicating a failed operation and includes an error.
func NewAppResultWithError(err error) AppResult {
	return AppResult{
		Ok:     false,
		Err:    err,
		Result: nil,
	}
}

// NewAppResult creates a new AppResult indicating a successful operation.
func NewAppResult() AppResult {
	return AppResult{
		Ok:     true,
		Err:    nil,
		Result: nil,
	}
}

// NewAppResultWithValue creates a new AppResult indicating a successful operation and includes a result value.
func NewAppResultWithValue(result interface{}) AppResult {
	return AppResult{
		Ok:     true,
		Err:    nil,
		Result: result,
	}
}

// RepositoryStatus represents the status of a repository.
type RepositoryStatus struct {
	IsValid  bool   `json:"is_valid" yaml:"is_valid" xml:"is_valid"`
	IsJoined bool   `json:"is_joined" yaml:"is_joined" xml:"is_joined"`
	RepoId   string `json:"repo_id" yaml:"repo_id" xml:"repo_id"`
}
