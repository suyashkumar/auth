package db

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sirupsen/logrus"
)

const DefaultMaxIdleConns = 5

var ErrorNoConnectionString = errors.New("A connection string must be specified on the first call to Get")

var db *gorm.DB

func Get(dbConnection string) (*gorm.DB, error) {
	if db != nil {
		return db, nil
	}

	if dbConnection == "" {
		return nil, ErrorNoConnectionString
	}

	d, err := gorm.Open("postgres", dbConnection)
	if err != nil {
		logrus.WithField("DBConnString", dbConnection).Error("Unable to connect to database")
		logrus.Error(err)
	}

	d.DB().SetMaxIdleConns(DefaultMaxIdleConns)

	db = d

	return db, nil

}
