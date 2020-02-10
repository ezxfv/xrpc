package xrpc

import "errors"

type Authenticator interface {
	Authenticate(args map[string]interface{}) error
}

func NewEmptyAuthenticator() Authenticator {
	return &emptyAuthenticator{}
}

type emptyAuthenticator struct{}

func (a *emptyAuthenticator) Authenticate(args map[string]interface{}) (err error) {
	return nil
}

func NewAdminAuthenticator(user, pass string) Authenticator {
	return &AdminAuthenticator{
		user: user,
		pass: pass,
	}
}

type AdminAuthenticator struct {
	user string
	pass string
}

func (a *AdminAuthenticator) Authenticate(args map[string]interface{}) (err error) {
	var user, pass string
	var ok bool
	if user, ok = args["user"].(string); !ok {
		err = errors.New("admin_auth: get user failed")
		return
	}
	if pass, ok = args["pass"].(string); !ok {
		err = errors.New("admin_auth: get pass failed")
		return
	}
	ok = a.user == user && a.pass == pass
	if !ok {
		err = errors.New("admin_auth: user or pass is incorrect")
		return
	}
	return err
}
