package user

import (
	"log"
	"strconv"
	"strings"

	"git.brobridge.com/pilotwave/pilotwave/pkg/app"
	"git.brobridge.com/pilotwave/pilotwave/pkg/pagination"
	"git.brobridge.com/pilotwave/pilotwave/pkg/user_manager"
	// "git.brobridge.com/pilotwave/pilotwave/pkg/user_manager/user/model"
	"git.brobridge.com/pilotwave/pilotwave/pkg/auth/authenticator/model"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	app app.App
}

func NewUser(a app.App) *User {
	return &User{
		app: a,
	}
}

func (user *User) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(bytes), err
}

func (user *User) ConfirmUser(username string) (string, error) {

	// Getting data from database
	db := user.app.GetDatabase()
	u := model.User{}

	// Check router is exists or not
	err := db.Where("username = ?", username).Find(&u).Error
	if err != nil {
		log.Println(err)
		return "", err
	}

	return u.ID, err
}

func (user *User) GetUsers(page int, perPage int, search string, isDisabled string) ([]user_manager.UserResponse, int, error) {

	paginationOptions, err := pagination.CreatePaginationOptions(page, perPage)
	if err != nil {
		log.Println(err)
		return []user_manager.UserResponse{}, 0, err
	}

	// TODO: sort

	// Getting data from database
	db := user.app.GetDatabase()

	// Preparing query
	query := db.Model(model.User{})

	// Search isDisabled
	if isDisabled != "" {
		disable, _ := strconv.ParseBool(isDisabled)
		query = query.Where("is_disabled = ?", disable)
	}

	// Search keyword
	searchCol := "name"
	if search != "" {
		searchableColumns := []string{
			"username",
			"email",
		}

		query = query.Where(searchCol+" LIKE ?", "%"+search+"%")
		for _, column := range searchableColumns {
			query = query.Or(column+" LIKE ?", "%"+search+"%")
		}
	}

	// Counting records
	var totalRecords int
	if query.Count(&totalRecords).Error != nil {
		log.Println(err)
		return []user_manager.UserResponse{}, totalRecords, err
	}

	// Fetching records
	items := make([]user_manager.UserResponse, 0)

	if totalRecords > 0 {
		var users []model.User

		if perPage >= 0 { // Paginations
			query = query.Limit(paginationOptions.PerPage)
			query = query.Offset(paginationOptions.Offset)
		}

		// Querying
		err := query.
			// Order(order).
			Find(&users).
			Error
		if err != nil {
			log.Println(err)
			return []user_manager.UserResponse{}, totalRecords, err
		}

		// Preparing JSON structure
		for _, u := range users {
			perms := []string{}
			if u.Permissions != "" {
				perms = strings.Split(u.Permissions, ",")
			}

			items = append(items, user_manager.UserResponse{
				ID:          u.ID,
				Name:        u.Name,
				Username:    u.Username,
				Email:       u.Email,
				Permissions: perms,
				IsDisabled:  u.IsDisabled,
				CreatedAt:   u.CreatedAt.Unix(),
				UpdatedAt:   u.UpdatedAt.Unix(),
			})
		}
	}

	return items, totalRecords, nil
}

func (user *User) CreateUser(name string, username string, password string, email string, permissions []string) (string, error) {

	// Getting data from database
	db := user.app.GetDatabase()

	hash, err := user.HashPassword(password)
	if err != nil {
		log.Println(err)
		return "", err
	}

	// Create User
	u := model.User{
		Name:        name,
		Username:    username,
		Password:    hash,
		Permissions: strings.Join(permissions, ", "),
		Email:       email,
	}

	err = db.Create(&u).Error
	if err != nil {
		return "", err
	}

	return u.ID, err
}

func (user *User) UpdateUser(userId string, name string, email string, permissions []string) (string, error) {

	// Getting data from database
	db := user.app.GetDatabase()

	// Update User
	u := model.User{}

	err := db.Model(&u).Where("id = ?", userId).Updates(map[string]interface{}{
		"Name":        name,
		"Permissions": strings.Join(permissions, ", "),
		"Email":       email,
	}).Error

	if err != nil {
		return "", err
	}

	return userId, err
}

func (user *User) GetUser(userId string) (*user_manager.UserResponse, error) {

	// Getting data from database
	db := user.app.GetDatabase()
	u := model.User{}

	err := db.Where("id = ?", userId).Find(&u).Error
	if err != nil {
		return &user_manager.UserResponse{}, err
	}

	perms := []string{}
	if u.Permissions != "" {
		perms = strings.Split(u.Permissions, ",")
	}
	return &user_manager.UserResponse{
		ID:          u.ID,
		Name:        u.Name,
		Username:    u.Username,
		Email:       u.Email,
		Permissions: perms,
		IsDisabled:  u.IsDisabled,
		CreatedAt:   u.CreatedAt.Unix(),
		UpdatedAt:   u.UpdatedAt.Unix(),
	}, nil
}

func (user *User) DeleteUser(userId string) (string, error) {

	// Getting data from database
	db := user.app.GetDatabase()
	u := model.User{}

	err := db.Unscoped().Delete(&u, "id = ?", userId).Error
	if err != nil {
		return "", err
	}

	return userId, err
}

func (user *User) EnableUser(userId string, enabled bool) (string, error) {

	// Getting data from database
	db := user.app.GetDatabase()

	// Enabled User
	u := model.User{}
	err := db.Model(&u).Where("id = ?", userId).Update("IsDisabled", !enabled).Error

	if err != nil {
		return "", err
	}

	return userId, err
}

func (user *User) UpdateUserPassword(userId string, password string) (string, error) {

	// Getting data from database
	db := user.app.GetDatabase()

	// Update User
	u := model.User{}

	hash, err := user.HashPassword(password)
	if err != nil {
		log.Println(err)
		return "", err
	}

	err = db.Model(&u).Where("id = ?", userId).Updates(map[string]interface{}{
		"Password": hash,
	}).Error

	if err != nil {
		return "", err
	}

	return userId, err
}
