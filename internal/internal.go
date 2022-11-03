package internal

import (
	"context"

	"github.com/hyphengolang/prelude/types/email"
	"github.com/hyphengolang/prelude/types/password"
	"github.com/hyphengolang/prelude/types/suid"
)

type ContextKey string

type UserTyp string

const (
	GuestUser      UserTyp = "guest"
	RegisteredUser UserTyp = "registered"
	AdminUser      UserTyp = "admin"
)

type User struct {
	ID       suid.UUID             `json:"id"`
	Username string                `json:"username"`
	Email    email.Email           `json:"email,omitempty"`
	Password password.PasswordHash `json:"-"`
}

type UserRepo interface {
	RUserRepo
	WUserRepo
}

type RUserRepo interface {
	Context() context.Context
	SelectMany(ctx context.Context) ([]User, error)
	// Method takes a context and a key value. If key is not of type
	// `internal.Email` or `suid.SUID` or `string` an invalid type error will be returned.
	Select(ctx context.Context, key any) (*User, error)
}
type WUserRepo interface {
	Context() context.Context
	Insert(ctx context.Context, u *User) error
	// Method takes a context and a key value. If key is not of type
	// `internal.Email` or `suid.SUID` an invalid type error will be returned.
	//
	// Perform either soft or hard delete. Default will be a "soft delete"
	// if no value was passed via the context
	//
	//	ctx := context.WithValue(context.Background(), RuleSoftDeletion, HardDelete)
	//	r.Delete(ctx, "fizz@mail.com")
	Delete(ctx context.Context, key any) error
}
