package controller

import (
	"bytes"
	"encoding/json"
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
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

func commentRoute(commentController *CommentController) *gin.Engine {
	logger := zap.NewNop()
	errorHandler := middleware.NewErrorsMiddleware(logger).GinHandlerFunc()
	checkJwt := middleware.NewCheckJwtMiddleware(jwtutil.New(jwt.SigningMethodHS256, []byte("test-secret")), logger).GinHandlerFunc()
	ensureAuth := middleware.NewEnsureAuthMiddleware(logger).GinHandlerFunc()

	r := gin.New()
	api := r.Group("api", errorHandler, checkJwt)
	articles := api.Group("articles")
	comments := articles.Group(":slug/comments")
	comments.POST("", ensureAuth, commentController.AddCommentToArticle)
	comments.GET("", commentController.GetCommentsFromArticle)
	comments.DELETE("/:id", ensureAuth, commentController.DeleteComment)

	return r
}

func TestCommentController_AddCommentToArticle(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_ports.NewMockCommentService(ctrl)

	cs.EXPECT().
		Create(gomock.Eq(uint(1)), gomock.Eq("test-slug"), gomock.Eq("test body")).
		Return(domain.CommentView{
			ID:              1,
			Body:            "test-body",
			ArticleID:       1,
			AuthorID:        1,
			AuthorUsername:  "test",
			AuthorFollowing: false,
		}, nil).
		AnyTimes()

	c := NewCommentController(cs)
	r := commentRoute(c)

	t.Run("댓글 작성 성공", func(t *testing.T) {
		w := httptest.NewRecorder()

		commentReq := AddCommentRequest{}
		commentReq.Comment.Body = "test body"
		body, err := json.Marshal(&commentReq)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/articles/test-slug/comments", bytes.NewReader(body))
		setAuthorization(req, 1, "test")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		resp := CommentResponse{}
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test-body", resp.Comment.Body)
	})
}

func TestCommentController_DeleteComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	cs := mock_ports.NewMockCommentService(ctrl)

	cs.EXPECT().
		Delete(gomock.Eq(uint(1)), gomock.Eq(uint(1))).
		Return(nil).
		AnyTimes()

	c := NewCommentController(cs)
	r := commentRoute(c)

	t.Run("댓글 삭제 성공", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/articles/test-slug/comments/1", nil)
		setAuthorization(req, 1, "test")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("인증 없이 댓글 삭제", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/articles/test-slug/comments/1", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
