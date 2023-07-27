package domain

import (
	"database/sql"
	"github.com/KumKeeHyun/gin-realworld/pkg/crypto"
	"github.com/KumKeeHyun/gin-realworld/pkg/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	Email    string `gorm:"unique;index"`
	Username string `gorm:"unique;index"`
	Password types.Password
	Bio      string
	Image    sql.NullString
	Token    string `gorm:"-:all"`
}

func (u *User) UpdatePassword(password string) {
	u.Password = types.Password{String: password, Encrypted: false}
}

func (u User) ValidPassword(password string) bool {
	return crypto.CheckHashAndPassword(u.Password.String, password)
}

type Follow struct {
	gorm.Model
	FollowerID  uint `gorm:"index:idx_follower_ing"`
	Follower    User
	FollowingID uint `gorm:"index:idx_follower_ing"`
	Following   User
}

type Profile struct {
	ID        uint
	Username  string
	Bio       string
	Image     sql.NullString
	Following bool
}

func (u User) AccessClaim() AccessClaim {
	return AccessClaim{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "go-gin-realworld",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		UID:      u.ID,
		Email:    u.Email,
		Username: u.Username,
		Bio:      u.Bio,
		Image:    lo.If(u.Image.Valid, &u.Image.String).Else(nil),
	}
}

type AccessClaim struct {
	jwt.RegisteredClaims
	UID      uint    `json:"user_id"`
	Email    string  `json:"email"`
	Username string  `json:"username"`
	Bio      string  `json:"bio"`
	Image    *string `json:"image"`
}
