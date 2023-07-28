package service

import (
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports/mock_ports"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"testing"
)

func Test_profileService_Find(t *testing.T) {
	ctrl := gomock.NewController(t)
	ur := mock_ports.NewMockUserRepository(ctrl)

	ur.EXPECT().
		FindByUsername(gomock.Eq("test")).
		Return(domain.User{
			Model:    gorm.Model{ID: 2},
			Email:    "test@example.com",
			Username: "test",
		}, nil)
	ur.EXPECT().
		FindByUsername(gomock.Eq("null")).
		Return(domain.User{}, gorm.ErrRecordNotFound)
	ur.EXPECT().
		FindProfile(gomock.Any(), gomock.Eq(uint(2))).
		Return(domain.Profile{
			ID:       2,
			Username: "test",
		}, nil)

	s := NewProfileService(ur, zap.NewNop())

	t.Run("조회 성공", func(t *testing.T) {
		profile, err := s.Find(1, "test")

		assert.NoError(t, err)
		assert.Equal(t, uint(2), profile.ID)
		assert.Equal(t, "test", profile.Username)
	})
	t.Run("없는 유저 조회", func(t *testing.T) {
		_, err := s.Find(1, "null")

		assert.ErrorIs(t, err, ports.ErrResourceNotFound)
	})
}

func Test_profileService_Follow(t *testing.T) {
	ctrl := gomock.NewController(t)
	ur := mock_ports.NewMockUserRepository(ctrl)

	ur.EXPECT().
		FindByUsername(gomock.Eq("test")).
		Return(domain.User{
			Model:    gorm.Model{ID: 2},
			Email:    "test@example.com",
			Username: "test",
		}, nil)
	ur.EXPECT().
		FindByUsername(gomock.Eq("self")).
		Return(domain.User{
			Model:    gorm.Model{ID: 1},
			Email:    "self@example.com",
			Username: "self",
		}, nil)
	ur.EXPECT().
		FindByUsername(gomock.Eq("null")).
		Return(domain.User{}, gorm.ErrRecordNotFound)
	ur.EXPECT().
		CreateFollow(gomock.Any(), gomock.Eq(uint(2))).
		Return(domain.Follow{}, nil)

	s := NewProfileService(ur, zap.NewNop())
	t.Run("팔로우 성공", func(t *testing.T) {
		profile, err := s.Follow(1, "test")

		assert.NoError(t, err)
		assert.Equal(t, "test", profile.Username)
		assert.Equal(t, true, profile.Following)
	})
	t.Run("자신 팔로우", func(t *testing.T) {
		_, err := s.Follow(1, "self")

		assert.ErrorIs(t, err, ports.ErrSelfFollowing)
	})
	t.Run("없는 유저 팔로우", func(t *testing.T) {
		_, err := s.Follow(1, "null")

		assert.ErrorIs(t, err, ports.ErrResourceNotFound)
	})
}

func Test_profileService_Unfollow(t *testing.T) {
	ctrl := gomock.NewController(t)
	ur := mock_ports.NewMockUserRepository(ctrl)

	ur.EXPECT().
		FindByUsername(gomock.Eq("test1")).
		Return(domain.User{
			Model:    gorm.Model{ID: 2},
			Email:    "test1@example.com",
			Username: "test1",
		}, nil)
	ur.EXPECT().
		FindByUsername(gomock.Eq("test2")).
		Return(domain.User{
			Model:    gorm.Model{ID: 3},
			Email:    "test2@example.com",
			Username: "test2",
		}, nil)
	ur.EXPECT().
		FindByUsername(gomock.Eq("null")).
		Return(domain.User{}, gorm.ErrRecordNotFound)
	ur.EXPECT().
		DeleteFollow(gomock.Any(), gomock.Eq(uint(2))).
		Return(nil)
	ur.EXPECT().
		DeleteFollow(gomock.Any(), gomock.Eq(uint(3))).
		Return(gorm.ErrRecordNotFound)

	s := NewProfileService(ur, zap.NewNop())
	t.Run("언팔로우 성공", func(t *testing.T) {
		profile, err := s.Unfollow(1, "test1")

		assert.NoError(t, err)
		assert.Equal(t, "test1", profile.Username)
		assert.Equal(t, false, profile.Following)
	})
	t.Run("팔로운 안한 유저 언팔로우", func(t *testing.T) {
		profile, err := s.Unfollow(1, "test2")

		assert.NoError(t, err)
		assert.Equal(t, "test2", profile.Username)
		assert.Equal(t, false, profile.Following)
	})
	t.Run("없는 유저 언팔로우", func(t *testing.T) {
		_, err := s.Unfollow(1, "null")

		assert.ErrorIs(t, err, ports.ErrResourceNotFound)
	})
}
