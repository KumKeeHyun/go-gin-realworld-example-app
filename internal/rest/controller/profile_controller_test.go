package controller

import (
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

func profileRoute(profileController *ProfileController) *gin.Engine {
	logger := zap.NewNop()
	errorHandler := middleware.NewErrorsMiddleware(logger).GinHandlerFunc()
	checkJwt := middleware.NewCheckJwtMiddleware(jwtutil.New(jwt.SigningMethodHS256, []byte("test-secret")), logger).GinHandlerFunc()
	ensureAuth := middleware.NewEnsureAuthMiddleware(logger).GinHandlerFunc()

	r := gin.New()
	api := r.Group("api", errorHandler, checkJwt)
	profiles := api.Group("profiles")
	profiles.GET("/:username", profileController.GetProfile)
	profiles.POST("/:username/follow", ensureAuth, profileController.FollowUser)
	profiles.DELETE("/:username/follow", ensureAuth, profileController.UnfollowUser)

	return r
}

func TestProfileController_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	ps := mock_ports.NewMockProfileService(ctrl)

	ps.EXPECT().
		Find(gomock.Eq(uint(1)), gomock.Eq("test2")).
		Return(domain.Profile{
			ID:        2,
			Username:  "test2",
			Following: true,
		}, nil).
		AnyTimes()
	ps.EXPECT().
		Find(gomock.Eq(uint(0)), gomock.Eq("test2")).
		Return(domain.Profile{
			ID:        2,
			Username:  "test2",
			Following: false,
		}, nil).
		AnyTimes()

	c := NewProfileController(ps)
	r := profileRoute(c)

	t.Run("프로필 조회 성공", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/profiles/test2", nil)
		setAuthorization(req, 1, "test")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		resp := ProfileResponse{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test2", resp.Profile.Username)
		assert.True(t, resp.Profile.Following)
	})
	t.Run("인증 없이 프로필 조회 성공", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/profiles/test2", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		resp := ProfileResponse{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test2", resp.Profile.Username)
		assert.False(t, resp.Profile.Following)
	})
}

func TestProfileController_FollowUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	ps := mock_ports.NewMockProfileService(ctrl)

	ps.EXPECT().
		Follow(gomock.Eq(uint(1)), gomock.Eq("test2")).
		Return(domain.Profile{
			ID:        2,
			Username:  "test2",
			Following: true,
		}, nil).
		AnyTimes()

	c := NewProfileController(ps)
	r := profileRoute(c)

	t.Run("팔로우 성공", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/profiles/test2/follow", nil)
		setAuthorization(req, 1, "test")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		resp := ProfileResponse{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test2", resp.Profile.Username)
		assert.True(t, resp.Profile.Following)
	})
	t.Run("인증 없이 팔로우", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/profiles/test2/follow", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestProfileController_UnfollowUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	ps := mock_ports.NewMockProfileService(ctrl)

	ps.EXPECT().
		Unfollow(gomock.Eq(uint(1)), gomock.Eq("test2")).
		Return(domain.Profile{
			ID:        2,
			Username:  "test2",
			Following: false,
		}, nil).
		AnyTimes()

	c := NewProfileController(ps)
	r := profileRoute(c)

	t.Run("언팔로우 성공", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/profiles/test2/follow", nil)
		setAuthorization(req, 1, "test")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		resp := ProfileResponse{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test2", resp.Profile.Username)
		assert.False(t, resp.Profile.Following)
	})
	t.Run("인증 없이 언팔로우", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/profiles/test2/follow", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
