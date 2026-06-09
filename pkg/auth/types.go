package auth

type AuthenticateResponse struct {
	ID          string
	Username    string
	Name        string
	Email       string
	Permissions []string
}

type UserInfo struct {
	ID          string
	Username    string
	Password    string
	Name        string
	Email       string
	IsDisabled  bool
	Permissions []string
}

type Authenticator interface {
	GetUserInfo(string) (*UserInfo, error)
	Authenticate(string, string) (*AuthenticateResponse, error)
	AuthenticateWithAD(string, string) (*AuthenticateResponse, error)
}
