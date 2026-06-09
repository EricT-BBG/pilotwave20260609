package instance

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
)

func (a *AppInstance) initDatabase() error {

	dbType := viper.GetString("database.type")

	var uri string

	switch dbType {
	case "sqlite3":
		uri = viper.GetString("database.dbpath")
	default:

		host := viper.GetString("database.host")
		port := viper.GetInt("database.port")
		dbname := viper.GetString("database.dbname")
		username := viper.GetString("database.username")
		password := viper.GetString("database.password")

		uri = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True",
			username,
			password,
			host,
			port,
			dbname,
		)
	}

	return a.database.Init(dbType, uri)
}

func (a *AppInstance) GetDatabase() *gorm.DB {
	return a.database.GetInstance()
}
