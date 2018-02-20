package auth

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

const DefaultMaxIdleConns = 5

var ErrorNoConnectionString = errors.New("A connection string must be specified on the first call to Get")

// DatabaseHandler abstracts away common persistence operations needed for this package
type DatabaseHandler interface {
	// GetUser gets a user from the database that matches constraints on the input user
	GetUser(u User) (User, error)
	// UpsertUser updates a user (if input user UUID matches one in the db) or inserts a user
	UpsertUser(u User) error
}

type databaseHandler struct {
	db *gorm.DB
}

// NewDatabaseHandler initializes and returns a new DatabaseHandler
func NewDatabaseHandler(dbConnection string) (DatabaseHandler, error) {
	db, err := getDB(dbConnection)
	if err != nil {
		return nil, err
	}
	// AutoMigrate relevant schemas
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
	err := a.db.Where(User{UUID: u.UUID}).Assign(u).FirstOrCreate(&User{}).Error
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
