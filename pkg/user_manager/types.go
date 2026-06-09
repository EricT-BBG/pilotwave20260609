package user_manager

type UserResponse struct {
	ID          string   `json:"uid"`
	Name        string   `json:"name"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Permissions []string `json:"permissions"`
	IsDisabled  bool     `json:"isDisabled"`
	UpdatedAt   int64    `json:"updatedAt"`
	CreatedAt   int64    `json:"createdAt"`
}

type UserRequest struct {
	Name        string   `json:"name"`
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	Email       string   `json:"email"`
	Permissions []string `json:"permissions"`
}

type UserEnableRequest struct {
	Enabled bool `json:"enabled"`
}

type UserPasswordRequest struct {
	Password string `json:"password"`
}

type User interface {
	GetUsers(int, int, string, string) ([]UserResponse, int, error)
	CreateUser(string, string, string, string, []string) (string, error)
	UpdateUser(string, string, string, []string) (string, error)
	GetUser(string) (*UserResponse, error)
	DeleteUser(string) (string, error)
	EnableUser(string, bool) (string, error)
	UpdateUserPassword(string, string) (string, error)
}
