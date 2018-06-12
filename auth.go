package auth

import (
	"errors"

	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

// Authenticator exposes the minimal set of operations needed for authentication
type Authenticator interface {
	Register(user *User, password string) error
	GetToken(email string, password string, opts *GetTokenOpts) (token string, err error)
	Validate(token string) (*Claims, error)
}

type auth struct {
	databaseHandler DatabaseHandler
	signingKey      []byte
}

type TokenData map[string]string

// Claims represents data that are encoded into an authentication token
type Claims struct {
	UserUUID    string    `json:"user_uuid"`
	Permissions int64     `json:"permissions"`
	Email       string    `json:"email"`
	Data        TokenData `json:"data"`
	jwt.StandardClaims
}

type GetTokenOpts struct {
	RequestedPermissions Permissions
	TimeToLive           time.Duration
	Data                 TokenData
}

// ErrorValidatingToken indicates issues validating the provided token
var ErrorValidatingToken = errors.New("problem validating token")

// ErrorExceededMaxPermissionLevel indicates that a requested op exceeded the maximum permission level
var ErrorExceededMaxPermissionLevel = errors.New(
	"you're requesting a token permission level that exceeds this user's maximum permission level",
)

// DefaultTTL represents the default time to live for a newly issued token in nanoseconds, in this case 8 hours
const DefaultTTL = time.Duration(2.88e13)

// NewAuthenticator returns a newly initialized Authenticator
func NewAuthenticator(dbConnection string, signingKey []byte) (Authenticator, error) {
	d, err := NewDatabaseHandler(dbConnection)
	if err != nil {
		return nil, err
	}

	return &auth{
		databaseHandler: d,
		signingKey:      signingKey,
	}, nil
}

// NewAuthenticatorFromGORM returns a newly init'd Authenticator from a *gorm.DB
func NewAuthenticatorFromGORM(db *gorm.DB, signingKey []byte) (Authenticator, error) {
	d, err := NewDatabaseHandlerFromGORM(db)
	if err != nil {
		return nil, err
	}

	return &auth{
		databaseHandler: d,
		signingKey:      signingKey,
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
	err = a.databaseHandler.UpsertUser(*newUser)
	if err != nil {
		return err
	}

	return nil
}

// GetToken mints a new authentication token at the given requestedPermissions level, if possible.
func (a *auth) GetToken(email string, password string, opts *GetTokenOpts) (string, error) {
	// Check database for User and verify credentials
	user, err := a.databaseHandler.GetUser(User{Email: email})

	if err != nil {
		return "", err
	}

	// Check hashed password
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		// Passwords don't match!
		return "", err
	}

	// Verify requestedPermissions
	if opts.RequestedPermissions > user.MaxPermissionLevel {
		return "", ErrorExceededMaxPermissionLevel
	}

	// Generate a login token for this user
	var ttl = opts.TimeToLive
	if ttl == 0 {
		ttl = DefaultTTL
	}

	c := Claims{
		UserUUID:    user.UUID.String(),
		Permissions: int64(opts.RequestedPermissions),
		Email:       user.Email,
		Data:        opts.Data,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(ttl).Unix(),
		},
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
