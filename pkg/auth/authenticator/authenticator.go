package authenticator

import (
	"log"
	"strings"

	"git.brobridge.com/pilotwave/pilotwave/pkg/app"
	"git.brobridge.com/pilotwave/pilotwave/pkg/auth"
	"git.brobridge.com/pilotwave/pilotwave/pkg/auth/authenticator/model"
	"golang.org/x/crypto/bcrypt"
)

type Authenticator struct {
	app app.App
}

func NewAuthenticator(a app.App) *Authenticator {
	return &Authenticator{
		app: a,
	}
}

func (authenticator *Authenticator) GetUserInfo(username string) (*auth.UserInfo, error) {

	// Getting data from database
	db := authenticator.app.GetDatabase()

	// Check user
	user := model.User{}
	if db.Where("username = ?", username).First(&user).RecordNotFound() {
		return nil, nil
	}

	// Parse permissions string
	perms := []string{}
	if user.Permissions != "" {
		perms = strings.Split(user.Permissions, ",")
	}

	return &auth.UserInfo{
		ID:          user.ID,
		Username:    user.Username,
		Password:    user.Password,
		Name:        user.Name,
		Email:       user.Email,
		IsDisabled:  user.IsDisabled,
		Permissions: perms,
	}, nil
}

func (authenticator *Authenticator) Authenticate(username string, password string) (*auth.AuthenticateResponse, error) {

	// Getting user information
	user, err := authenticator.GetUserInfo(username)
	if err != nil {
		return nil, err
	}

	// Not exist
	if user == nil {
		return nil, nil
	}

	log.Println(user)
	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Println(err)
		return nil, nil
	}

	// Disabled already
	if user.IsDisabled {
		return nil, nil
	}

	return &auth.AuthenticateResponse{
		ID:          user.ID,
		Username:    user.Username,
		Name:        user.Name,
		Email:       user.Email,
		Permissions: user.Permissions,
	}, nil
}

func (authenticator *Authenticator) AuthenticateWithAD(username string, password string) (*auth.AuthenticateResponse, error) {

	connector := NewADConnector()

	// Verify user with active directory
	exists, err := connector.Verify(username, password)
	if err != nil {
		return nil, nil
	}

	if !exists {
		return nil, nil
	}

	// Getting user information
	user, err := authenticator.GetUserInfo(username)
	if err != nil {
		return nil, err
	}

	// Not exist
	if user == nil {
		return nil, nil
	}

	// Disabled already
	if user.IsDisabled {
		return nil, nil
	}

	return &auth.AuthenticateResponse{
		ID:          user.ID,
		Username:    user.Username,
		Name:        user.Name,
		Email:       user.Email,
		Permissions: user.Permissions,
	}, nil
}
