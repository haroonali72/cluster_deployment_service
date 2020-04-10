package types

type CustomCPError struct {
	StatusCode  string `json:"code,omitempty"  bson:"code"`
	Type        string `json:"type,omitempty"  bson:"type"`
	Message     string `json:"message,omitempty"  bson:"message"`
	Description string `json:"description,omitempty"  bson:"description"`
}
