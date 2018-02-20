package auth

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

const DefaultMaxIdleConns = 5

var ErrorNoConnectionString = errors.New("A connection string must be specified on the first call to Get")

type DatabaseHandler interface {
	GetUser(u User) (User, error)
	UpsertUser(u User) error
}

type databaseHandler struct {
	db *gorm.DB
}

func NewDatabaseHandler(dbConnection string) (DatabaseHandler, error) {
	db, err := getDB(dbConnection)
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&User{})

	return &databaseHandler{
		db: db,
	}, nil
}

func (a *databaseHandler) GetUser(u User) (User, error) {
	var foundUser User
	err := a.db.Where(&u).First(&foundUser).Error
	if err != nil {
		return User{}, err
	}

	return foundUser, nil
}

func (a *databaseHandler) UpsertUser(u User) error {
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
		return nil, err
	}

	d.DB().SetMaxIdleConns(DefaultMaxIdleConns)

	return d, nil

}
