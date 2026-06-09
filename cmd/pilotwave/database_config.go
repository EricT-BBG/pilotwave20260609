package main

import (
	"fmt"

	"github.com/spf13/viper"
)

func databaseURIFromConfig(dbType string) (string, error) {
	switch dbType {
	case "sqlite3":
		return viper.GetString("database.dbpath"), nil
	default:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True",
			viper.GetString("database.username"),
			viper.GetString("database.password"),
			viper.GetString("database.host"),
			viper.GetInt("database.port"),
			viper.GetString("database.dbname"),
		), nil
	}
}
