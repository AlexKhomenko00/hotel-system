package auth

import "github.com/google/uuid"

type userContextKey string

var UsrCtxKey userContextKey = "user"

type UserContext struct {
	Id    uuid.UUID
	Email string
}
