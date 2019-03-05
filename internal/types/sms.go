package types

type SMS struct {
	Recipient  string `json:"recipient"`
	Originator string `json:"originator"`
	Message    string `json:"message"`
}
