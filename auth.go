package auth

import (
	"errors"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/suyashkumar/auth/db"
	"golang.org/x/crypto/bcrypt"
)

// Auth exposes the minimal set of operations needed for authentication
type Auth interface {
	Register(user *User, password string) error
	GetToken(email string, password string, requestedPermissions Permissions) (token string, err error)
	Validate(token string) (*Claims, error)
}

type auth struct {
	db         *gorm.DB
	signingKey []byte
}

// Claims represents data that are encoded into an authentication token
type Claims struct {
	UserUUID    string `json:"user_uuid"`
	Permissions int64  `json:"permissions"`
	Email       string `json:"email"`
	jwt.StandardClaims
}

var ErrorValidatingToken = errors.New("problem validating token")
var ErrorExceededMaxPermissionLevel = errors.New(
	"you're requesting a token permission level that exceeds this user's maximum permission level",
)

// NewAuthenticator returns a newly initialized Auth
func NewAuthenticator(dbConnection string, signingKey []byte) (Auth, error) {
	d, err := db.Get(dbConnection)
	if err != nil {
		return nil, err
	}

	// AutoMigrate any auth specific schemas:
	d.AutoMigrate(&User{})

	return &auth{
		db:         d,
		signingKey: signingKey,
	}, nil
}

// Register adds a new user.
func (a *auth) Register(newUser *User, password string) error {
	// Always generate a new UUID for newUser
	newUser.UUID = uuid.NewV4()

	// Hash password, add to the newUser struct
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	newUser.HashedPassword = string(hash)

	// Upsert user
	if err != nil {
		return err
	}
	err = a.db.Create(&newUser).Error
	if err != nil {
		return err
	}

	return nil
}

// GetToken mints a new authentication token at the given requestedPermissions level, if possible.
func (a *auth) GetToken(email string, password string, requestedPermissions Permissions) (string, error) {
	// Check database for User and verify credentials
	var user User
	err := a.db.Where(&User{Email: email}).First(&user).Error
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

	// Verify requestedPermissions
	if requestedPermissions > user.MaxPermissionLevel {
		return "", ErrorExceededMaxPermissionLevel
	}

	// Generate a login token for this user
	c := Claims{
		UserUUID:    user.UUID.String(),
		Permissions: int64(requestedPermissions),
		Email:       user.Email,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	token, err := t.SignedString(a.signingKey)
	if err != nil {
		return "", err
	}
	return token, nil
}

// Validate decrypts and validates a token. Returns any claims embedded in the token.
func (a *auth) Validate(token string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(token, &Claims{}, func(jt *jwt.Token) (interface{}, error) {
		return []byte(a.signingKey), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := t.Claims.(*Claims); ok && t.Valid {
		return claims, nil
	}

	return &Claims{}, ErrorValidatingToken
}
