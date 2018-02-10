# auth [WIP]
A simple (but opinionated) Golang authentication library with a very simple interface (below). A gRPC microservice wrapping this interface is in progress and can be found at suyashkumar/auth-grpc.

```go
type Auth interface {
	Register(user User, password string) error
	Login(email string, password string, requestedPermissions Permissions) (token string, err error)
	Validate(token string) (*Claims, error)
}
```

You only need to provide a database `connectionString` and `signingKey`, and everything else is taken care of for you including:
* table and database setup (including uniqueness constraints and useful indicies)
* hashing passwords using `bcrypt` on register
* comparing hashed passwords on login
* validation of new user fields like "Email"
* extraction of embedded fields that might be stored in the JWT
* ensuring that a token's requested permissions does not exceed the user's maximum permission level

A minimal example is below:
```go
a, _ := auth.NewAuthenticator(db_string, signing_key)

u := auth.User{
	UUID:               uuid.NewV4(),
	Email:              "test@test.com",
	MaxPermissionLevel: auth.PERMISSIONS_USER,
}

// Register a new user
a.Register(u, "password")

// Login as user
token, err := a.Login(u.Email, "password", auth.PERMISSIONS_USER)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("JWT Token: %s\n\n", token)

// Validate the user's token
claims, _ := a.Validate(token)
fmt.Printf("%+v", claims)
```

