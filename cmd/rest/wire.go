//go:build wireinject
// +build wireinject

package main

import (
	"github.com/KumKeeHyun/gin-realworld/internal/core/service"
	"github.com/KumKeeHyun/gin-realworld/internal/repository/sqlite"
	"github.com/KumKeeHyun/gin-realworld/internal/rest"
	"github.com/KumKeeHyun/gin-realworld/internal/rest/controller"
	"github.com/KumKeeHyun/gin-realworld/internal/rest/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"go.uber.org/zap"
)

var SqliteRepositorySet = wire.NewSet(
	sqlite.NewUserRepository,
	sqlite.NewArticleRepository,
	sqlite.NewCommentRepository,
)

var ServiceSet = wire.NewSet(
	service.NewAuthService,
	service.NewProfileService,
	service.NewArticleService,
	service.NewCommentService,
)

var ControllerSet = wire.NewSet(
	controller.NewAuthController,
	controller.NewProfileController,
	controller.NewArticleController,
	controller.NewCommentController,
)

var MiddlewareSet = wire.NewSet(
	middleware.NewCheckJwtMiddleware,
	middleware.NewEnsureAuthMiddleware,
	middleware.NewEnsureNotAuthMiddleware,
	middleware.NewTransactionMiddleware,
	middleware.NewErrorsMiddleware,
)

func InitRouterUsingSqlite(cfg *config, logger *zap.Logger) (*gin.Engine, error) {
	wire.Build(
		InitDatasource,
		InitJwtUtil,
		rest.NewRouter,

		MiddlewareSet,
		ControllerSet,
		ServiceSet,
		SqliteRepositorySet,
	)
	return nil, nil
}
