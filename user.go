package auth

import (
	"database/sql/driver"
	"time"

	"github.com/satori/go.uuid"
)

type Permissions int64

const (
	PERMISSIONS_USER  = 0
	PERMISSIONS_ADMIN = 1
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
	UUID           uuid.UUID `sql:"type:uuid;"`
	Email          string    `gorm:"unique_index"`
	HashedPassword string
	FirstName      string
	LastName       string
	Permissions    Permissions
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      time.Time `sql:"default:NULL"`
}
