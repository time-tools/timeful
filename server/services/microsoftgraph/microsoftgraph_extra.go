package microsoftgraph

// UserInfo and the Graph user response shape stay handwritten because their
// json tags define the wire shape.
type UserInfo struct {
	FirstName string `json:"givenName"`
	LastName  string `json:"surname"`
	Email     string `json:"mail"`
}

type microsoftGraphUserResponse struct {
	GivenName string `json:"givenName"`
	Surname   string `json:"surname"`
	Mail      string `json:"mail"`
}
