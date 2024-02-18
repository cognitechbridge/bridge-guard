package types

type Recipient struct {
	Email    string `json:"email,omitempty"`
	Public   string `json:"public,omitempty"`
	ClientId string `json:"clientId,omitempty"`
}
