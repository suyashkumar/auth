package main

import (
	"fmt"

	"github.com/satori/go.uuid"
	"github.com/suyashkumar/auth"
)

const db_string = "host=localhost port=5432 user=postgres sslmode=disable dbname=postgres password=postgres123test"

var signing_key = []byte("fake-signing-key")

func main() {
	a, _ := auth.NewAuthenticator(db_string, signing_key)

	u := auth.User{
		UUID:        uuid.NewV4(),
		Email:       "test@test.com",
		Permissions: auth.PERMISSIONS_USER,
	}

	// Register a new user
	a.Register(u, "password")

	// Login as user
	token, _ := a.Login(u.Email, "password")
	fmt.Printf("JWT Token: %s\n\n", token)

	// Validate the user's token
	claims, _ := a.Validate(token)
	fmt.Printf("%+v", claims)

}
