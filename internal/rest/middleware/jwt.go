package middleware

import (
	"errors"
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
	"github.com/KumKeeHyun/gin-realworld/pkg/jwtutil"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strings"
)

const (
	keyClaim = "claim"
)

var (
	ErrInvalidToken   = errors.New("invalid token string")
	ErrClaimNotExists = errors.New("claim not exists, you should ensure auth")
	ErrTokenNotExists = errors.New("token not exists")
	ErrEnsureAuth     = errors.New("authentication is required")
	ErrEnsureNotAuth  = errors.New("authentication is not required")
)

type (
	CheckJwtMiddleware struct {
		fn gin.HandlerFunc
	}
	EnsureAuthMiddleware struct {
		fn gin.HandlerFunc
	}
	EnsureNotAuthMiddleware struct {
		fn gin.HandlerFunc
	}
)

func (m CheckJwtMiddleware) GinHandlerFunc() gin.HandlerFunc {
	return m.fn
}

func (m EnsureAuthMiddleware) GinHandlerFunc() gin.HandlerFunc {
	return m.fn
}

func (m EnsureNotAuthMiddleware) GinHandlerFunc() gin.HandlerFunc {
	return m.fn
}

func NewCheckJwtMiddleware(jwtUtil *jwtutil.JwtUtil, rawLogger *zap.Logger) CheckJwtMiddleware {
	logger := rawLogger.Sugar().Named("checkJwtMiddleware")
	return CheckJwtMiddleware{
		fn: func(ctx *gin.Context) {
			tokenString := ctx.GetHeader("Authorization")
			tokenString, err := stripBearerPrefix(tokenString)
			if err == nil {
				logger.Debugw("find authorization", "token", tokenString)
				token, err := jwtUtil.ParseToClaims(tokenString, &domain.AccessClaim{})
				if err == nil {
					if claims, ok := token.Claims.(*domain.AccessClaim); ok && token.Valid {
						ctx.Set(keyClaim, claims)
					}
				} else {
					logger.Errorw("failed to parse jwt token", "err", err)
				}
			}
			ctx.Next()
		},
	}
}

func stripBearerPrefix(tokenString string) (string, error) {
	parts := strings.Split(tokenString, " ")
	if len(parts) != 2 {
		return "", ErrInvalidToken
	}

	token := strings.TrimSpace(parts[1])
	if len(token) <= 0 {
		return "", ErrInvalidToken
	}
	return token, nil
}

func GetAccessClaim(ctx *gin.Context) (domain.AccessClaim, error) {
	claim, exists := ctx.Get(keyClaim)
	if !exists {
		return domain.AccessClaim{}, ErrClaimNotExists
	}
	return *(claim.(*domain.AccessClaim)), nil
}

func GetToken(ctx *gin.Context) (string, error) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		return "", ErrTokenNotExists
	}
	return stripBearerPrefix(token)
}

func NewEnsureAuthMiddleware(*zap.Logger) EnsureAuthMiddleware {
	return EnsureAuthMiddleware{
		fn: func(ctx *gin.Context) {
			if _, exist := ctx.Get(keyClaim); !exist {
				ctx.Error(ErrEnsureAuth)
				return
			}
			ctx.Next()
		},
	}
}

func NewEnsureNotAuthMiddleware(*zap.Logger) EnsureNotAuthMiddleware {
	return EnsureNotAuthMiddleware{
		fn: func(ctx *gin.Context) {
			if _, exists := ctx.Get(keyClaim); exists {
				ctx.Error(ErrEnsureNotAuth)
				return
			}
			ctx.Next()
		},
	}
}
