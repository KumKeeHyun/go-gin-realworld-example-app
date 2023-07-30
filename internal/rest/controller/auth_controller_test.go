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
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

func authRoute(authController *AuthController) *gin.Engine {
	logger := zap.NewNop()
	errorHandler := middleware.NewErrorsMiddleware(logger).GinHandlerFunc()
	checkJwt := middleware.NewCheckJwtMiddleware(jwtutil.New(jwt.SigningMethodHS256, []byte("test-secret")), logger).GinHandlerFunc()
	ensureAuth := middleware.NewEnsureAuthMiddleware(logger).GinHandlerFunc()
	ensureNotAuth := middleware.NewEnsureNotAuthMiddleware(logger).GinHandlerFunc()

	r := gin.New()
	api := r.Group("api", errorHandler, checkJwt)
	users := api.Group("users")
	users.POST("/login", ensureNotAuth, authController.AuthenticateUser)
	users.POST("", ensureNotAuth, authController.RegisterUser)

	user := api.Group("user")
	user.GET("", ensureAuth, authController.GetCurrentUser)
	return r
}

func TestAuthController_AuthenticateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	as := mock_ports.NewMockAuthService(ctrl)

	as.EXPECT().
		Login(gomock.Eq("test@example.com"), gomock.Eq("test-password")).
		Return(domain.User{
			Model: gorm.Model{ID: 1},
			Email: "test@example.com",
			Token: "test-token",
		}, nil)

	c := NewAuthController(as)
	r := authRoute(c)

	t.Run("로그인 성공", func(t *testing.T) {
		w := httptest.NewRecorder()

		authReq := AuthenticateUserRequest{}
		authReq.User.Email = "test@example.com"
		authReq.User.Password = "test-password"
		body, err := json.Marshal(&authReq)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/users/login", bytes.NewReader(body))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := UserResponse{}
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", resp.User.Email)
		assert.Equal(t, "test-token", resp.User.Token)
	})
}

func TestAuthController_RegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	as := mock_ports.NewMockAuthService(ctrl)

	as.EXPECT().
		Register(gomock.Eq("test@example.com"), gomock.Eq("test"), gomock.Eq("test-password")).
		Return(domain.User{
			Model:    gorm.Model{ID: 1},
			Email:    "test@example.com",
			Username: "test",
			Token:    "test-token",
		}, nil).
		AnyTimes()

	c := NewAuthController(as)
	r := authRoute(c)

	t.Run("회원가입 성공", func(t *testing.T) {
		w := httptest.NewRecorder()

		registerReq := RegisterUserRequest{}
		registerReq.User.Email = "test@example.com"
		registerReq.User.Username = "test"
		registerReq.User.Password = "test-password"
		body, err := json.Marshal(&registerReq)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := UserResponse{}
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", resp.User.Email)
		assert.Equal(t, "test", resp.User.Username)
		assert.Equal(t, "test-token", resp.User.Token)
	})
	t.Run("잘못된 요청 바디", func(t *testing.T) {
		w := httptest.NewRecorder()

		registerReq := RegisterUserRequest{}
		registerReq.User.Email = "test@example.com"
		registerReq.User.Password = "test-password"
		body, err := json.Marshal(&registerReq)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthController_GetCurrentUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	as := mock_ports.NewMockAuthService(ctrl)

	c := NewAuthController(as)
	r := authRoute(c)

	t.Run("유저 정보 조회 성공", func(t *testing.T) {
		w := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
		setAuthorization(req, 1, "test")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := UserResponse{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test", resp.User.Username)
	})
}
