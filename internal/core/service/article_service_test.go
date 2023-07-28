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

func Test_articleService_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	ar := mock_ports.NewMockArticleRepository(ctrl)
	ur := mock_ports.NewMockUserRepository(ctrl)

	ar.EXPECT().
		FindBySlug(gomock.Eq("test-slug")).
		Return(domain.Article{
			Model:  gorm.Model{ID: 1},
			Slug:   "test-slug",
			Author: domain.Author{ID: 1},
		}, nil).
		AnyTimes()
	ar.EXPECT().
		Save(gomock.Any()).
		Return(domain.Article{}, nil)
	ar.EXPECT().
		FindFavorite(gomock.Any(), gomock.Eq(uint(1))).
		Return(domain.Favorite{}, nil)

	s := NewArticleService(ar, ur, zap.NewNop())
	t.Run("글 수정 성공", func(t *testing.T) {
		_, err := s.Update(1, "test-slug", ports.ArticleUpdateFields{})

		assert.NoError(t, err)
	})
	t.Run("다른 유저의 글 수정", func(t *testing.T) {
		_, err := s.Update(2, "test-slug", ports.ArticleUpdateFields{})

		assert.ErrorIs(t, err, ports.ErrNonOwnedContent)
	})
}

func Test_articleService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	ar := mock_ports.NewMockArticleRepository(ctrl)
	ur := mock_ports.NewMockUserRepository(ctrl)

	ar.EXPECT().
		FindBySlug(gomock.Eq("test-slug")).
		Return(domain.Article{
			Model:  gorm.Model{ID: 1},
			Slug:   "test-slug",
			Author: domain.Author{ID: 1},
		}, nil).
		AnyTimes()
	ar.EXPECT().
		DeleteBySlug("test-slug").
		Return(nil).
		AnyTimes()

	s := NewArticleService(ar, ur, zap.NewNop())
	t.Run("글 삭제 성공", func(t *testing.T) {
		err := s.Delete(1, "test-slug")

		assert.NoError(t, err)
	})
	t.Run("다른 유저의 글 삭제", func(t *testing.T) {
		err := s.Delete(2, "test-slug")

		assert.ErrorIs(t, err, ports.ErrNonOwnedContent)
	})
}
