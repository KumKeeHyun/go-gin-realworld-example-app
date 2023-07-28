package service

import (
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports/mock_ports"
	"github.com/KumKeeHyun/gin-realworld/pkg/crypto"
	"github.com/KumKeeHyun/gin-realworld/pkg/jwtutil"
	"github.com/KumKeeHyun/gin-realworld/pkg/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"testing"
)

func Test_authService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	ur := mock_ports.NewMockUserRepository(ctrl)
	ar := mock_ports.NewMockArticleRepository(ctrl)

	ur.EXPECT().
		FindByEmailOrUsername(gomock.Eq("test@example.com"), gomock.Eq("test")).
		Return(domain.User{}, gorm.ErrRecordNotFound)
	ur.EXPECT().
		FindByEmailOrUsername(gomock.Eq("dup@example.com"), gomock.Eq("dup")).
		Return(domain.User{}, nil)
	ur.EXPECT().
		Save(gomock.Any()).
		Return(domain.User{
			Email:    "test@example.com",
			Username: "test",
			Password: types.Password{Encrypted: true},
		}, nil)

	s := NewAuthService(ur, ar, jwtutil.New(jwt.SigningMethodHS256, []byte("test-secret")), zap.NewNop())
	t.Run("회원가입 성공", func(t *testing.T) {
		user, err := s.Register("test@example.com", "test", "test-password")

		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "test", user.Username)
		assert.NotEqual(t, "", user.Token)
	})
	t.Run("이메일 or 이름 중복", func(t *testing.T) {
		_, err := s.Register("dup@example.com", "dup", "test-password")

		assert.ErrorIs(t, err, ports.ErrDuplicatedEmailOrUsername)
	})
}

func Test_authService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	ur := mock_ports.NewMockUserRepository(ctrl)
	ar := mock_ports.NewMockArticleRepository(ctrl)

	hashPassword, _ := crypto.HashPassword("test-password")
	ur.EXPECT().
		FindByEmail(gomock.Eq("test@example.com")).
		Return(domain.User{
			Password: types.Password{String: hashPassword, Encrypted: true},
		}, nil).
		AnyTimes()
	ur.EXPECT().
		FindByEmail(gomock.Eq("null@example.com")).
		Return(domain.User{}, gorm.ErrRecordNotFound)

	s := NewAuthService(ur, ar, jwtutil.New(jwt.SigningMethodHS256, []byte("test-secret")), zap.NewNop())
	t.Run("로그인 성공", func(t *testing.T) {
		_, err := s.Login("test@example.com", "test-password")

		assert.NoError(t, err)
	})
	t.Run("틀린 비밀번호", func(t *testing.T) {
		_, err := s.Login("test@example.com", "invalid-password")

		assert.ErrorIs(t, err, ports.ErrInvalidPassword)
	})
	t.Run("없는 유저", func(t *testing.T) {
		_, err := s.Login("null@example.com", "test-password")

		assert.ErrorIs(t, err, ports.ErrResourceNotFound)
	})
}
