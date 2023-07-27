package rest

import (
	"github.com/KumKeeHyun/gin-realworld/internal/rest/controller"
	"github.com/KumKeeHyun/gin-realworld/internal/rest/middleware"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

func NewRouter(
	logger *zap.Logger,
	checkJwtMiddleware middleware.CheckJwtMiddleware,
	ensureAuthMiddleware middleware.EnsureAuthMiddleware,
	ensureNotAuthMiddleware middleware.EnsureNotAuthMiddleware,
	transactionMiddleware middleware.TransactionMiddleware,
	errorsMiddleware middleware.ErrorsMiddleware,
	authController *controller.AuthController,
	profileController *controller.ProfileController,
	articleController *controller.ArticleController,
	commentController *controller.CommentController) *gin.Engine {

	checkJwt := checkJwtMiddleware.GinHandlerFunc()
	ensureAuth := ensureAuthMiddleware.GinHandlerFunc()
	ensureNotAuth := ensureNotAuthMiddleware.GinHandlerFunc()
	transaction := transactionMiddleware.GinHandlerFunc()
	errorHandler := errorsMiddleware.GinHandlerFunc()

	r := gin.New()
	r.Use(ginzap.GinzapWithConfig(logger, &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		TraceID:    false,
	}))
	r.Use(ginzap.RecoveryWithZap(logger, false))

	api := r.Group("api", errorHandler, checkJwt)

	users := api.Group("users")
	users.POST("/login", ensureNotAuth, authController.AuthenticateUser)
	users.POST("", ensureNotAuth, authController.RegisterUser)

	user := api.Group("user")
	user.GET("", ensureAuth, authController.GetCurrentUser)
	user.PUT("", ensureAuth, transaction, authController.UpdateUser)

	profiles := api.Group("profiles")
	profiles.GET("/:username", profileController.GetProfile)
	profiles.POST("/:username/follow", ensureAuth, profileController.FollowUser)
	profiles.DELETE("/:username/follow", ensureAuth, profileController.UnfollowUser)

	articles := api.Group("articles")
	articles.GET("", articleController.ListArticles)
	articles.GET("/feed", ensureAuth, articleController.FeedArticles)
	articles.GET("/:slug", articleController.GetArticle)
	articles.POST("", ensureAuth, articleController.CreateArticle)
	articles.PUT("/:slug", ensureAuth, articleController.UpdateArticle)
	articles.DELETE("/:slug", ensureAuth, articleController.DeleteArticle)
	articles.POST("/:slug/favorite", ensureAuth, articleController.FavoriteArticle)
	articles.DELETE("/:slug/favorite", ensureAuth, articleController.UnfavoriteArticle)

	comments := articles.Group(":slug/comments")
	comments.POST("", ensureAuth, commentController.AddCommentToArticle)
	comments.GET("", commentController.GetCommentsFromArticle)
	comments.DELETE("/:id", ensureAuth, commentController.DeleteComment)

	api.GET("/tags", articleController.GetTags)

	return r
}
