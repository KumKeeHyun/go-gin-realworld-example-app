package types

import (
	"context"
	"errors"
	"github.com/KumKeeHyun/gin-realworld/pkg/crypto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Password struct {
	String    string
	Encrypted bool `gorm:"-:all"`
}

func (p Password) GormDataType() string {
	return "text"
}

func (p Password) GormValue(ctx context.Context, db *gorm.DB) (expr clause.Expr) {
	if p.Encrypted {
		return clause.Expr{SQL: "?", Vars: []interface{}{p.String}}
	} else if encrypted, err := crypto.HashPassword(p.String); err == nil {
		return clause.Expr{SQL: "?", Vars: []interface{}{encrypted}}
	} else {
		db.AddError(errors.New("failed to encrypted password"))
	}
	return
}

func (p *Password) Scan(value any) error {
	var ok bool
	p.String, ok = value.(string)
	if !ok {
		return errors.New("failed to unmarshal Password value")
	}
	p.Encrypted = true
	return nil
}
