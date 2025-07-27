package jwt

import (
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type Authenticator interface {
	Encode(claims map[string]interface{}) (jwt.Token, string, error)
	Verifier() func(http.Handler) http.Handler
}

type jwtAuthenticator struct {
	auth *jwtauth.JWTAuth
}

func NewAuthenticator(secret string) Authenticator {
	return &jwtAuthenticator{
		auth: jwtauth.New("HS256", []byte(secret), nil),
	}
}

func (j *jwtAuthenticator) Encode(claims map[string]interface{}) (jwt.Token, string, error) {
	return j.auth.Encode(claims)
}

func (j *jwtAuthenticator) Verifier() func(http.Handler) http.Handler {
	return jwtauth.Verifier(j.auth)
}
