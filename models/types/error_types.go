package types

type CustomCPError struct {
	//	Status				string 			 `json:"status,omitempty"  bson:"status"`
	StatusCode  int    `json:"code,omitempty"  bson:"code"`
	Type        string `json:"type,omitempty"  bson:"type"`
	Message     string `json:"message,omitempty"  bson:"message"`
	Description string `json:"description,omitempty"  bson:"description"`
}
