package sqlite

import (
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

func Test_commentRepository_FindFromArticle(t *testing.T) {
	f := newSqliteFixture(t)
	defer f.close()

	tests := []struct {
		name    string
		givenFn func(tx *gorm.DB) error
		thenFn  func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository, cr ports.CommentRepository)
	}{
		{
			name: "find comments",
			givenFn: func(tx *gorm.DB) error {
				user1 := &domain.User{Email: "test1@example.com", Username: "test1"}
				user2 := domain.User{Email: "test2@example.com", Username: "test2"}
				tx.Create(&user1)
				tx.Create(&user2)
				article1 := domain.Article{
					Slug:        "test1",
					Title:       "test1 title",
					Description: "test1 desc",
					Body:        "test1 body",
					Tags:        []string{"tag1", "tag2"},
					Author:      domain.Author{ID: user1.ID, Username: user1.Username},
				}
				article2 := domain.Article{
					Slug:        "test2",
					Title:       "test2 title",
					Description: "test2 desc",
					Body:        "test2 body",
					Tags:        []string{"tag2"},
					Author:      domain.Author{ID: user1.ID, Username: user1.Username},
				}
				tx.Create(&article1)
				tx.Create(&article2)
				comment1 := &domain.Comment{
					Body:      "test1 body",
					ArticleID: article1.ID,
					Author:    domain.Author{ID: user1.ID, Username: user1.Username},
				}
				comment2 := &domain.Comment{
					Body:      "test2 body",
					ArticleID: article1.ID,
					Author:    domain.Author{ID: user1.ID, Username: user1.Username},
				}
				tx.Create(&comment1)
				tx.Create(&comment2)
				return nil
			},
			thenFn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository, cr ports.CommentRepository) {
				comments, err := cr.FindFromArticle("test1")
				assert.NoError(t, err)
				assert.Equal(t, 2, len(comments))
				assert.Equal(t, "test1 body", comments[0].Body)
				assert.Equal(t, "test2 body", comments[1].Body)

				comments, err = cr.FindFromArticle("test2")
				assert.NoError(t, err)
				assert.Equal(t, 0, len(comments))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f.expectGiven(tt.givenFn)
			f.runWithComment(tt.thenFn)
		})
	}
}

func Test_commentRepository_Delete(t *testing.T) {
	f := newSqliteFixture(t)
	defer f.close()

	tests := []struct {
		name    string
		givenFn func(tx *gorm.DB) error
		thenFn  func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository, cr ports.CommentRepository)
	}{
		{
			name: "find comments",
			givenFn: func(tx *gorm.DB) error {
				user1 := &domain.User{Email: "test1@example.com", Username: "test1"}
				user2 := domain.User{Email: "test2@example.com", Username: "test2"}
				tx.Create(&user1)
				tx.Create(&user2)
				article1 := domain.Article{
					Slug:        "test1",
					Title:       "test1 title",
					Description: "test1 desc",
					Body:        "test1 body",
					Tags:        []string{"tag1", "tag2"},
					Author:      domain.Author{ID: user1.ID, Username: user1.Username},
				}
				tx.Create(&article1)
				comment1 := &domain.Comment{
					Body:      "test1 body",
					ArticleID: article1.ID,
					Author:    domain.Author{ID: user1.ID, Username: user1.Username},
				}
				comment2 := &domain.Comment{
					Body:      "test2 body",
					ArticleID: article1.ID,
					Author:    domain.Author{ID: user1.ID, Username: user1.Username},
				}
				tx.Create(&comment1)
				tx.Create(&comment2)
				return nil
			},
			thenFn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository, cr ports.CommentRepository) {
				err := cr.Delete(1, 1)
				assert.NoError(t, err)

				err = cr.Delete(2, 2)
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f.expectGiven(tt.givenFn)
			f.runWithComment(tt.thenFn)
		})
	}
}
