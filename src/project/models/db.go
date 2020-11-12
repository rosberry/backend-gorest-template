package models

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"project/config"
)

const (
	ReopenDelay = 10 * time.Second
	ReopenCount = 4
)

const (
	DBTypeMysql    = "mysql"
	DBTypePostgres = "postgres"
)

var db *gorm.DB
var dbType string

// GetDB returns an instance of * gorm.DB that is created and configured
// (if not already created) according to the application configuration.
func GetDB() *gorm.DB {
	if db == nil {
		GetDBType()

		var err error
		/*
			db, err = gorm.Open(dbType, cfg.App.DB.Settings)
			if err != nil {
				panic("DB error: " + err.Error())
			}
		*/

		tryCount := ReopenCount
		for {
			db, err = gorm.Open(dbType, config.App.DB.Settings)
			if err == nil || tryCount == 0 {
				break
			}
			time.Sleep(ReopenDelay)
			tryCount--
		}
		if err != nil {
			panic("DB error: " + err.Error())
		}

		if config.App.Mode == config.ModeDebug {
			db.LogMode(true)
		}
		switch dbType {
		case "mysql":
			db.Exec("SET time_zone = '+00:00';")
			db.Exec("SET NAMES utf8mb4 COLLATE utf8mb4_bin;")
		case "postgres":
			db.Exec("SET TIME ZONE '+00:00';")
		}
	}
	return db
}

// GetDBType returns the type of the DBMS as defined in the application configuration
func GetDBType() string {
	if dbType == "" {
		dbType = config.App.DB.Type
	}
	return dbType
}

/*
tryCount := ReopenCount
		for {
			switch dbType {
			case DBTypeMysql:
				db, err = gorm.Open(mysql.Open(config.App.DB.DSN), gormConfig)
			case DBTypePostgres:
				db, err = gorm.Open(postgres.Open(config.App.DB.DSN), gormConfig)
			}
			if err == nil || tryCount == 0 {
				break
			}
			time.Sleep(ReopenDelay)
			tryCount--
		}
		if err != nil {
			panic("DB error: " + err.Error())
		}

*/
