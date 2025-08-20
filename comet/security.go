package comet

import (
	"context"
	"time"
)

// PoliciesConfig is a map with all controller policies
// configuration, such as Role, Permission or custom policies
type PoliciesConfig map[string][]Policy

type Policy struct {
	Validation AuthorizerFunction
	Value      interface{}
}

func Authorize(fn AuthorizerFunction, val interface{}) Policy {
	return Policy{
		Validation: fn,
		Value:      val,
	}
}

type AuthorizerFunction = func(RequestHandler, interface{}) RequestHandler

type AuthorizationMap map[interface{}]AuthorizerFunction

type Claims map[string]interface{}

type SessionManager interface {
	GetToken(claims Claims) string
	Validate(req *Request) (Claims, error)
	GetUser(ctx context.Context) (ApplicationUser, error)
}

type ApplicationUser interface{}

type User struct {
	ApplicationUser
	ID           string  `gorm:"id,primaryKey,size:255"`
	Username     string  `gorm:"username,not null,unique,size:255"`
	Email        string  `gorm:"email,not null,size:255"`
	PhoneNumber  *string `gorm:"phone_number,size:255"`
	PasswordHash string  `gorm:"password_hash,not null,size:256"`
	IsActive     bool    `gorm:"is_active,default:false"`

	CreatedAt time.Time `gorm:"created_at"`
	UpdatedAt int64     `gorm:"updated_at,autoUpdate:milli"`

	Roles []*Role `gorm:"many2many:user_roles"`
}

func (User) TableName() string {
	return "users"
}

type Role struct {
	ID   string `gorm:"id,primaryKey,size:255"`
	Name string `gorm:"name,not null,size:255"`

	_           []User        `gorm:"many2many:user_roles"`
	Permissions []*Permission `gorm:"many2many:role_permissions"`
}

func (Role) TableName() string {
	return "roles"
}

type Permission struct {
	ID   string `gorm:"id,primaryKey,size:255"`
	Name string `gorm:"name,size:255"`

	_ []*Role `gorm:"many2many:role_permissions"`
}

func (Permission) TableName() string {
	return "permissions"
}

const (
	upperRegex   = `[A-Z]`
	lowerRegex   = `[a-z]`
	digitRegex   = `[0-9]`
	specialRegex = `[^a-zA-Z0-9]`
)

type PasswordConfig struct {
	MinimumChars     int
	NeedUppercase    bool
	NeedDigits       bool
	NeedSpecialChars bool
	NeedLowercase    bool
}

type UserConfig struct {
	PasswordConfig  PasswordConfig
	NeedsActivation bool
}
