package auth

type Principal struct {
	Issuer  string `json:"issuer"`
	Subject string `json:"subject"`
	Email   string `json:"email"`
	Name    string `json:"name"`
}
