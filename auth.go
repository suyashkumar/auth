package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"github.com/suyashkumar/auth/db"
	"golang.org/x/crypto/bcrypt"
)

type Auth interface {
	Register(user User, password string) error
	Login(email string, password string) (token string, err error)
	Verify(token string) (bool, error)
}

type auth struct {
	dbConnection string
	signingKey   []byte
}

type Claims struct {
	UserUUID    string `json:"user_uuid"`
	Permissions int64  `json:"permissions"`
	Email       string `json:"email"`
	jwt.StandardClaims
}

func NewAuthenticator(dbConnection string, signingKey []byte) (*auth, error) {
	d, err := db.Get(dbConnection)
	if err != nil {
		return nil, err
	}
	d.AutoMigrate(&User{})

	return &auth{
		dbConnection: dbConnection,
		signingKey:   signingKey,
	}, nil
}

// Register adds a new user
func (a *auth) Register(newUser User, password string) error {
	// Hash password, add to the newUser struct
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	newUser.HashedPassword = string(hash)

	// Upsert user
	d, err := db.Get("")
	if err != nil {
		return err
	}
	err = d.Create(&newUser).Error
	if err != nil {
		return err
	}

	return nil
}

func (a *auth) Login(email string, password string) (token string, err error) {
	// Check database for User and verify credentials
	var user User
	d, err := db.Get("")
	if err != nil {
		return "", err
	}
	err = d.Where(&User{Email: email}).First(&user).Error
	if err != nil {
		logrus.Error(err)
		return "", err
	}

	// Check hashed password
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		// Passwords don't match!
		return "", err
	}

	// Generate a login token for this user
	c := Claims{
		UserUUID:    user.UUID.String(),
		Permissions: int64(user.Permissions),
		Email:       user.Email,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	token, err = t.SignedString(a.signingKey)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (a *auth) Verify(token string) (bool, error) {
	return false, nil
}
