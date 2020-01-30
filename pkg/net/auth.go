package net

type Authenticator interface {
	Authenticate(user, password string) bool
}
