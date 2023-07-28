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

func Test_commentService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	cr := mock_ports.NewMockCommentRepository(ctrl)
	ar := mock_ports.NewMockArticleRepository(ctrl)
	ur := mock_ports.NewMockUserRepository(ctrl)

	ur.EXPECT().
		FindByID(gomock.Any()).
		Return(domain.User{
			Model:    gorm.Model{ID: 1},
			Email:    "test@example.com",
			Username: "test",
		}, nil).
		AnyTimes()
	ar.EXPECT().
		FindBySlug(gomock.Eq("test-slug")).
		Return(domain.Article{Model: gorm.Model{ID: 1}}, nil)
	ar.EXPECT().
		FindBySlug("null").
		Return(domain.Article{}, gorm.ErrRecordNotFound)
	cr.EXPECT().Save(gomock.Any()).Return(domain.Comment{
		Model:     gorm.Model{ID: 1},
		Body:      "test-body",
		ArticleID: 1,
		Author:    domain.Author{ID: 1, Username: "test"},
	}, nil)

	s := NewCommentService(cr, ar, ur, zap.NewNop())
	t.Run("댓글 생성 성공", func(t *testing.T) {
		comment, err := s.Create(1, "test-slug", "test-body")

		assert.NoError(t, err)
		assert.Equal(t, "test-body", comment.Body)
		assert.Equal(t, "test", comment.AuthorUsername)
	})
	t.Run("없는 글에 댓글 생성", func(t *testing.T) {
		_, err := s.Create(1, "null", "test-body")

		assert.ErrorIs(t, err, ports.ErrResourceNotFound)
	})
}

func Test_commentService_GetFromArticle(t *testing.T) {
	ctrl := gomock.NewController(t)
	cr := mock_ports.NewMockCommentRepository(ctrl)
	ar := mock_ports.NewMockArticleRepository(ctrl)
	ur := mock_ports.NewMockUserRepository(ctrl)

	ar.EXPECT().
		FindBySlug(gomock.Eq("test-slug")).
		Return(domain.Article{Model: gorm.Model{ID: 1}}, nil)
	ar.EXPECT().
		FindBySlug("null").
		Return(domain.Article{}, gorm.ErrRecordNotFound)
	cr.EXPECT().
		FindFromArticle(gomock.Any()).
		Return([]domain.Comment{
			{
				Model:     gorm.Model{ID: 1},
				Body:      "test-body-1",
				ArticleID: 1,
				Author:    domain.Author{ID: 1, Username: "test1"},
			},
			{
				Model:     gorm.Model{ID: 2},
				Body:      "test-body-2",
				ArticleID: 1,
				Author:    domain.Author{ID: 2, Username: "test2"},
			},
		}, nil)
	ur.EXPECT().
		FindFollows(gomock.Any(), gomock.Any()).
		Return([]domain.Follow{
			{
				Model:       gorm.Model{ID: 1},
				FollowerID:  1,
				FollowingID: 2,
			},
		}, nil)

	s := NewCommentService(cr, ar, ur, zap.NewNop())
	t.Run("댓글 조회 성공", func(t *testing.T) {
		comments, err := s.GetFromArticle(1, "test-slug")

		assert.NoError(t, err)
		assert.Len(t, comments, 2)
		assert.Equal(t, "test1", comments[0].AuthorUsername)
		assert.Equal(t, false, comments[0].AuthorFollowing)
		assert.Equal(t, "test2", comments[1].AuthorUsername)
		assert.Equal(t, true, comments[1].AuthorFollowing)
	})
	t.Run("없는 글의 댓글 조회", func(t *testing.T) {
		_, err := s.GetFromArticle(1, "null")

		assert.ErrorIs(t, err, ports.ErrResourceNotFound)
	})
}
