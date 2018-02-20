package auth

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sirupsen/logrus"
)

const DefaultMaxIdleConns = 5

var ErrorNoConnectionString = errors.New("A connection string must be specified on the first call to Get")

type Storer interface {
	GetUser(u User) (User, error)
	UpsertUser(u User) error
}

type storer struct {
	db *gorm.DB
}

func NewStorer(dbConnection string) (Storer, error) {
	db, err := getDB(dbConnection)
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&User{})

	return &storer{
		db: db,
	}, nil
}

func (a *storer) GetUser(u User) (User, error) {
	var foundUser User
	err := a.db.Where(&u).First(&foundUser).Error
	if err != nil {
		return User{}, err
	}

	return foundUser, nil
}

func (a *storer) UpsertUser(u User) error {
	err := a.db.Where(u).Assign(u).FirstOrCreate(&User{}).Error
	if err != nil {
		return err
	}

	return nil
}

func getDB(dbConnection string) (*gorm.DB, error) {

	if dbConnection == "" {
		return nil, ErrorNoConnectionString
	}

	d, err := gorm.Open("postgres", dbConnection)
	if err != nil {
		logrus.WithField("DBConnString", dbConnection).Error("Unable to connect to database")
		logrus.Error(err)
	}

	d.DB().SetMaxIdleConns(DefaultMaxIdleConns)

	return d, nil

}
