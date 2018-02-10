package auth

import (
	"database/sql/driver"
	"time"

	"github.com/satori/go.uuid"
)

type Permissions int64

// These permission levels are ORDERED lowest to highest
const (
	PERMISSIONS_API   = 0
	PERMISSIONS_USER  = 1
	PERMISSIONS_ADMIN = 2
)

func (p *Permissions) Scan(value interface{}) error {
	*p = Permissions(value.(int64))
	return nil
}

func (p Permissions) Value() (driver.Value, error) {
	return int64(p), nil
}

// User represents a generic User that Suyash expects in his projects. TODO: generalize
type User struct {
	UUID               uuid.UUID `sql:"type:uuid;"`
	Email              string    `gorm:"unique_index"`
	HashedPassword     string
	FirstName          string
	LastName           string
	MaxPermissionLevel Permissions
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          time.Time `sql:"default:NULL"`
}
