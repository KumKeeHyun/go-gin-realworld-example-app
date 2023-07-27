package middleware

import (
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"net/http"
)

type ErrorsResponse struct {
	Errors struct {
		Body []string `json:"body"`
	} `json:"errors"`
}

func NewErrorsResponse(errs ...error) ErrorsResponse {
	var resp ErrorsResponse
	if errs != nil {
		resp.Errors.Body = lo.Map(errs, func(err error, idx int) string { return err.Error() })
	}
	return resp
}

type (
	ErrorsMiddleware struct {
		fn gin.HandlerFunc
	}
)

func (m ErrorsMiddleware) GinHandlerFunc() gin.HandlerFunc {
	return m.fn
}

func NewErrorsMiddleware(rawLogger *zap.Logger) ErrorsMiddleware {
	logger := rawLogger.Sugar().Named("errorsMiddleware")
	return ErrorsMiddleware{
		fn: func(ctx *gin.Context) {
			ctx.Next()
			if len(ctx.Errors) == 0 {
				return
			}

			var errs []error
			for _, err := range ctx.Errors {
				if validationErrs, ok := err.Err.(validator.ValidationErrors); ok {
					ctx.JSON(http.StatusBadRequest, NewErrorsResponse(validationErrs))
					return
				}
				switch err.Err {
				case ports.ErrInternal:
					ctx.JSON(http.StatusInternalServerError, NewErrorsResponse(err))
					return
				case ports.ErrResourceNotFound,
					ports.ErrInvalidPassword,
					ports.ErrSelfFollowing,
					ports.ErrDuplicatedEmailOrUsername,
					ErrEnsureNotAuth:
					ctx.JSON(http.StatusBadRequest, NewErrorsResponse(err))
					return
				case ports.ErrNonOwnedContent:
					ctx.JSON(http.StatusForbidden, NewErrorsResponse(err))
					return
				case ErrEnsureAuth:
					ctx.JSON(http.StatusUnauthorized, NewErrorsResponse(err))
					return
				default:
					errs = append(errs, err)
				}
			}
			logger.Warnw("unhandled error", "errs", errs)
			ctx.JSON(http.StatusInternalServerError, NewErrorsResponse(errs...))
		},
	}
}
