package controller

import (
	"bytes"
	"encoding/json"
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports/mock_ports"
	"github.com/KumKeeHyun/gin-realworld/internal/rest/middleware"
	"github.com/KumKeeHyun/gin-realworld/pkg/jwtutil"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func articleRoute(articleController *ArticleController) *gin.Engine {
	logger := zap.NewNop()
	errorHandler := middleware.NewErrorsMiddleware(logger).GinHandlerFunc()
	checkJwt := middleware.NewCheckJwtMiddleware(jwtutil.New(jwt.SigningMethodHS256, []byte("test-secret")), logger).GinHandlerFunc()
	ensureAuth := middleware.NewEnsureAuthMiddleware(logger).GinHandlerFunc()

	r := gin.New()
	api := r.Group("api", errorHandler, checkJwt)
	articles := api.Group("articles")
	articles.GET("", articleController.ListArticles)
	articles.GET("/feed", ensureAuth, articleController.FeedArticles)
	articles.GET("/:slug", articleController.GetArticle)
	articles.POST("", ensureAuth, articleController.CreateArticle)
	articles.PUT("/:slug", ensureAuth, articleController.UpdateArticle)
	articles.DELETE("/:slug", ensureAuth, articleController.DeleteArticle)
	articles.POST("/:slug/favorite", ensureAuth, articleController.FavoriteArticle)
	articles.DELETE("/:slug/favorite", ensureAuth, articleController.UnfavoriteArticle)

	return r
}

func setAuthorization(req *http.Request, id uint, username string) {
	jwtUtil := jwtutil.New(jwt.SigningMethodHS256, []byte("test-secret"))
	token, _ := jwtUtil.SignClaims(domain.AccessClaim{
		UID:      id,
		Username: username,
	})
	req.Header["Authorization"] = []string{"token " + token}
}

func TestArticleController_CreateArticle(t *testing.T) {
	ctrl := gomock.NewController(t)
	as := mock_ports.NewMockArticleService(ctrl)

	as.EXPECT().
		Create(gomock.Eq(uint(1)), gomock.Eq("test title"), gomock.Eq("test desc"), gomock.Eq("test body"), gomock.Any()).
		Return(domain.ArticleView{
			ID:          1,
			Slug:        "test-slug",
			Title:       "test title",
			Description: "test desc",
			Body:        "test body",
		}, nil).
		AnyTimes()

	c := NewArticleController(as)
	r := articleRoute(c)

	t.Run("글 작성 성공", func(t *testing.T) {
		w := httptest.NewRecorder()

		createReq := CreateArticleRequest{}
		createReq.Article.Title = "test title"
		createReq.Article.Description = "test desc"
		createReq.Article.Body = "test body"
		body, err := json.Marshal(&createReq)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/articles", bytes.NewReader(body))
		setAuthorization(req, 1, "test")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		resp := ArticleResponse{}
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test-slug", resp.Article.Slug)
		assert.Equal(t, "test title", resp.Article.Title)
		assert.Equal(t, "test desc", resp.Article.Description)
		assert.Equal(t, "test body", resp.Article.Body)
	})
	t.Run("인증 없이 요청", func(t *testing.T) {
		w := httptest.NewRecorder()

		createReq := CreateArticleRequest{}
		createReq.Article.Title = "test title"
		createReq.Article.Description = "test desc"
		createReq.Article.Body = "test body"
		body, err := json.Marshal(&createReq)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/articles", bytes.NewReader(body))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestArticleController_GetArticle(t *testing.T) {
	ctrl := gomock.NewController(t)
	as := mock_ports.NewMockArticleService(ctrl)

	as.EXPECT().
		Find(gomock.Any(), "test-slug").
		Return(domain.ArticleView{
			ID:   1,
			Slug: "test-slug",
		}, nil).
		AnyTimes()
	as.EXPECT().
		Find(gomock.Any(), "null").
		Return(domain.ArticleView{}, ports.ErrResourceNotFound)

	c := NewArticleController(as)
	r := articleRoute(c)

	t.Run("글 조회 성공", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/articles/test-slug", nil)
		setAuthorization(req, 1, "test")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		resp := ArticleResponse{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test-slug", resp.Article.Slug)
	})
	t.Run("인증 없이 글 조회 성공", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/articles/test-slug", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		resp := ArticleResponse{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test-slug", resp.Article.Slug)
	})
	t.Run("없는 글 조회", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/articles/null", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestArticleController_DeleteArticle(t *testing.T) {
	ctrl := gomock.NewController(t)
	as := mock_ports.NewMockArticleService(ctrl)

	as.EXPECT().
		Delete(gomock.Any(), "test-slug").
		Return(nil).
		AnyTimes()
	as.EXPECT().
		Delete(gomock.Any(), "null").
		Return(ports.ErrResourceNotFound)

	c := NewArticleController(as)
	r := articleRoute(c)

	t.Run("글 삭제 성공", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/articles/test-slug", nil)
		setAuthorization(req, 1, "test")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("인증 없이 글 삭제", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/articles/test-slug", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("없는 글 삭제", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/articles/null", nil)
		setAuthorization(req, 1, "test")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestArticleController_FavoriteArticle(t *testing.T) {
	ctrl := gomock.NewController(t)
	as := mock_ports.NewMockArticleService(ctrl)

	as.EXPECT().
		Favorite(uint(1), "test-slug").
		Return(domain.ArticleView{
			ID:             1,
			Slug:           "test-slug",
			Title:          "test-title",
			Description:    "test-desc",
			Body:           "test-body",
			Favorited:      true,
			FavoritesCount: 1,
			AuthorID:       2,
			AuthorUsername: "test2",
		}, nil).
		AnyTimes()

	c := NewArticleController(as)
	r := articleRoute(c)

	t.Run("글 좋아요 성공", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/articles/test-slug/favorite", nil)
		setAuthorization(req, 1, "test")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		resp := ArticleResponse{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test-slug", resp.Article.Slug)
		assert.True(t, resp.Article.Favorited)
	})
	t.Run("인증 없이 글 좋아요", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/articles/test-slug/favorite", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
