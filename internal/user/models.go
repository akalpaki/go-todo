package user

import (
	"net/mail"
)

// User is the model representing the User entity
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserRequest is a model that represents the minimum required information to create a new user.
// Requests should always be validated with the Valid method before being accepted.
type UserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r UserRequest) Valid() bool {
	return r.Email != "" && isEmail(r.Email) && r.Password != ""
}

func isEmail(email string) bool {
	address, err := mail.ParseAddress(email)
	return address.String() == email && err == nil
}
