package internal

import (
	"context"

	"github.com/hyphengolang/prelude/types/email"
	"github.com/hyphengolang/prelude/types/password"
	"github.com/hyphengolang/prelude/types/suid"
)

type ContextKey string

type User struct {
	ID       suid.UUID             `json:"id"`
	Username string                `json:"username"`
	Email    email.Email           `json:"email"`
	Password password.PasswordHash `json:"-"`
}

type UserRepo interface {
	RUserRepo
	WUserRepo
}

type RUserRepo interface {
	Context() context.Context
	Close(ctx context.Context) error
	SelectMany(ctx context.Context) ([]User, error)
	// Method takes a context and a key value. If key is not of type
	// `internal.Email` or `suid.SUID` or `string` an invalid type error will be returned.
	Select(ctx context.Context, key any) (*User, error)
}
type WUserRepo interface {
	Context() context.Context
	Close(ctx context.Context) error
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

// type Email string

// func (e Email) String() string { return string(e) }

// func (e Email) Validate() error {
// 	_, err := mail.ParseAddress(string(e))
// 	return err
// }

// func (e Email) IsValid() bool { return e.Validate() == nil }

// func (e *Email) UnmarshalJSON(b []byte) error {
// 	*e = Email(b[1 : len(b)-1])
// 	return e.Validate()
// }

// const minEntropy float64 = 40.0 // during production, this value needs to be > 40

// type Password string

// func (p Password) String() string { return string(p) }

// func (p Password) Validate() error {
// 	return gpv.Validate(p.String(), minEntropy)
// }

// func (p Password) IsValid() bool { return p.Validate() == nil }

// func (p *Password) UnmarshalJSON(b []byte) error {
// 	*p = Password(b[1 : len(b)-1])
// 	return p.Validate()
// }

// func (p Password) MarshalJSON() (b []byte, err error) {
// 	return []byte(`"` + p.String() + `"`), nil
// }

// func (p Password) Hash() (PasswordHash, error) {
// 	return bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
// }

// func (p Password) MustHash() PasswordHash {
// 	h, err := p.Hash()
// 	if err != nil {
// 		panic(err)
// 	}
// 	return h
// }

// type PasswordHash []byte

// func (h PasswordHash) String() string { return string(h) }

// func (h PasswordHash) Compare(cmp string) error {
// 	return bcrypt.CompareHashAndPassword(h, []byte(cmp))
// }
