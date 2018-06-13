package tide

// Auth represents an authenticated user.
type Auth struct {
	AccessToken string     `json:"access_token"`
	Client      AuthClient `json:"client"`
}

// AuthClient represents a Data object with information about the user that is connected to Tide.
type AuthClient struct {
	Data ClientData `json:"data"`
}

// ClientData represents information about the user that is connected to Tide.
type ClientData struct {
	ID          string `json:"ID"`
	UserLogin   string `json:"user_login"`
	UserEmail   string `json:"user_email"`
	DisplayName string `json:"display_name"`
}

// ClientInterface describes a client that can authenticate with and send payloads to Tide API.
type ClientInterface interface {
	Authenticate(clientID, clientSecret, authEndpoint string) error
	SendPayload(method, endpoint, data string) (string, error)
}
