package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
)

const (
	keyTx = "db_tx"
)

var (
	ErrTransactionNotExists = errors.New("transaction not exists")
)

func StatusInList(status int, statusList []int) bool {
	for _, i := range statusList {
		if i == status {
			return true
		}
	}
	return false
}

type (
	TransactionMiddleware struct {
		fn gin.HandlerFunc
	}
)

func (m TransactionMiddleware) GinHandlerFunc() gin.HandlerFunc {
	return m.fn
}

func NewTransactionMiddleware(db *gorm.DB, rawLogger *zap.Logger) TransactionMiddleware {
	logger := rawLogger.Sugar().Named("transactionMiddleware")
	return TransactionMiddleware{
		fn: func(ctx *gin.Context) {
			logger.Debugw("begin transaction")
			tx := db.Begin()
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
				}
			}()

			ctx.Set(keyTx, tx)
			ctx.Next()

			if StatusInList(ctx.Writer.Status(), []int{http.StatusOK, http.StatusCreated}) {
				logger.Debugw("commit transaction")
				if err := tx.Commit().Error; err != nil {
					logger.Errorw("failed to commit transaction", "err", err)
				}
			} else {
				logger.Infow("rollback transaction due to status code", "code", ctx.Writer.Status())
				tx.Rollback()
			}
		},
	}
}

func GetTransaction(ctx *gin.Context) (*gorm.DB, error) {
	tx, exists := ctx.Get(keyTx)
	if !exists {
		return nil, ErrTransactionNotExists
	}
	return tx.(*gorm.DB), nil
}
