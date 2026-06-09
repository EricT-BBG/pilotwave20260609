package database

import (
	"fmt"
	"log"

	auth_model "git.brobridge.com/pilotwave/pilotwave/pkg/auth/authenticator/model"
	router_model "git.brobridge.com/pilotwave/pilotwave/pkg/router_manager/router/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase() *Database {
	return &Database{}
}

func (database *Database) Init(dbType string, uri string) error {

	log.Printf("Connecting to database server %s ...\n", uri)

	db, err := gorm.Open(dbType, uri)
	if err != nil {
		return err
	}

	db.LogMode(viper.GetBool("database.debug_mode"))

	if err := db.AutoMigrate(&auth_model.User{}).Error; err != nil {
		return err
	}
	if err := db.AutoMigrate(&router_model.Grafana{}).Error; err != nil {
		return err
	}

	database.db = db

	err = ensureDefaultAdmin(db)
	if err != nil {
		return err
	}

	return nil
}

func ensureDefaultAdmin(db *gorm.DB) error {
	adminHash, err := hashPassword("admin")
	if err != nil {
		return err
	}

	// New a default user
	u := auth_model.User{}
	err = db.Table("users").Where("username = ?", "admin").Find(&u).Error

	if err != nil {
		// Create User
		u := auth_model.User{
			Name:        "admin",
			Username:    "admin",
			Password:    string(adminHash),
			Permissions: "admin",
		}

		err = db.Create(&u).Error
		if err != nil {
			log.Println(err)
		}
	} else if viper.GetBool("dev.reset_admin_password") {
		err = db.Model(&u).Where("username = ?", "admin").Updates(map[string]interface{}{
			"Password":    string(adminHash),
			"Permissions": "admin",
			"IsDisabled":  false,
		}).Error
		if err != nil {
			return err
		}
		log.Println("Reset built-in admin password for local development")
	}

	return nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func resetUserPassword(db *gorm.DB, username string, passwordHash string) error {
	result := db.Model(&auth_model.User{}).Where("username = ?", username).Updates(map[string]interface{}{
		"Password":   passwordHash,
		"IsDisabled": false,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user %q not found", username)
	}
	return nil
}

func (database *Database) ResetUserPassword(username string, password string) error {
	if database.db == nil {
		return fmt.Errorf("database is not initialized")
	}
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}

	hash, err := hashPassword(password)
	if err != nil {
		return err
	}

	return resetUserPassword(database.db, username, hash)
}

func (database *Database) GetInstance() *gorm.DB {
	return database.db
}
