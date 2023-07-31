//go:build sqlite
// +build sqlite

package sqlite

import (
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

func Test_articleRepository_FindBySearchConditions(t *testing.T) {
	f := newSqliteFixture(t)

	tests := []struct {
		name    string
		givenFn func(tx *gorm.DB) error
		thenFn  func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository)
	}{
		{
			name: "find article by favorited",
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
				article3 := domain.Article{
					Slug:        "test3",
					Title:       "test3 title",
					Description: "test3 desc",
					Body:        "test3 body",
					Tags:        []string{"tag1"},
					Author:      domain.Author{ID: user2.ID, Username: user2.Username},
				}
				tx.Create(&article1)
				tx.Create(&article2)
				tx.Create(&article3)
				tx.Create(&domain.Favorite{
					UserID:    user1.ID,
					ArticleID: article1.ID,
				})
				tx.Create(&domain.Favorite{
					UserID:    user1.ID,
					ArticleID: article3.ID,
				})
				return nil
			},
			thenFn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository) {
				favorited := "test1"
				cond := ports.ArticleSearchConditions{
					Favorited: &favorited,
					Pageable:  ports.Pageable{Limit: 20, Offset: 0},
				}
				articles, err := ar.FindBySearchConditions(cond)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(articles))
				assert.Equal(t, "test1", articles[0].Slug)
				assert.Equal(t, "test3", articles[1].Slug)

				author := "test1"
				cond = ports.ArticleSearchConditions{
					Author:   &author,
					Pageable: ports.Pageable{Limit: 20, Offset: 0},
				}
				articles, err = ar.FindBySearchConditions(cond)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(articles))
				assert.Equal(t, "test1", articles[0].Slug)
				assert.Equal(t, "test2", articles[1].Slug)

				cond = ports.ArticleSearchConditions{
					Author:    &author,
					Favorited: &favorited,
					Pageable:  ports.Pageable{Limit: 20, Offset: 0},
				}
				articles, err = ar.FindBySearchConditions(cond)
				assert.NoError(t, err)
				assert.Equal(t, 1, len(articles))
				assert.Equal(t, "test1", articles[0].Slug)
			},
		},
		{
			name: "find article by tag",
			givenFn: func(tx *gorm.DB) error {
				user := &domain.User{Email: "test1@example.com", Username: "test1"}
				tx.Create(&user)
				article1 := domain.Article{
					Slug:        "test1",
					Title:       "test1 title",
					Description: "test1 desc",
					Body:        "test1 body",
					Tags:        []string{"tag1", "tag2"},
					Author:      domain.Author{ID: user.ID, Username: user.Username},
				}
				article2 := domain.Article{
					Slug:        "test2",
					Title:       "test2 title",
					Description: "test2 desc",
					Body:        "test2 body",
					Tags:        []string{"tag2"},
					Author:      domain.Author{ID: user.ID, Username: user.Username},
				}
				article3 := domain.Article{
					Slug:        "test3",
					Title:       "test3 title",
					Description: "test3 desc",
					Body:        "test3 body",
					Tags:        []string{"tag1"},
					Author:      domain.Author{ID: user.ID, Username: user.Username},
				}
				tx.Create(&article1)
				tx.Create(&article2)
				tx.Create(&article3)
				return nil
			},
			thenFn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository) {
				tag := "tag1"
				cond := ports.ArticleSearchConditions{
					Tag:      &tag,
					Pageable: ports.Pageable{Limit: 20, Offset: 0},
				}
				articles, err := ar.FindBySearchConditions(cond)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(articles))
				assert.Equal(t, "test1", articles[0].Slug)
				assert.Equal(t, "test3", articles[1].Slug)

			},
		},
		{
			name: "find not existing article",
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
				article3 := domain.Article{
					Slug:        "test3",
					Title:       "test3 title",
					Description: "test3 desc",
					Body:        "test3 body",
					Tags:        []string{"tag1"},
					Author:      domain.Author{ID: user2.ID, Username: user2.Username},
				}
				tx.Create(&article1)
				tx.Create(&article2)
				tx.Create(&article3)
				return nil
			},
			thenFn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository) {
				favorited := "test1"
				cond := ports.ArticleSearchConditions{
					Favorited: &favorited,
					Pageable:  ports.Pageable{Limit: 20, Offset: 0},
				}
				articles, err := ar.FindBySearchConditions(cond)
				assert.NoError(t, err)
				assert.Equal(t, 0, len(articles))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f.expectGiven(tt.givenFn)
			f.run(tt.thenFn)
		})
	}
}

func Test_articleRepository_FindFavorites(t *testing.T) {
	f := newSqliteFixture(t)

	tests := []struct {
		name    string
		givenFn func(tx *gorm.DB) error
		thenFn  func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository)
	}{
		{
			name: "find favorites",
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
				article3 := domain.Article{
					Slug:        "test3",
					Title:       "test3 title",
					Description: "test3 desc",
					Body:        "test3 body",
					Tags:        []string{"tag1"},
					Author:      domain.Author{ID: user2.ID, Username: user2.Username},
				}
				tx.Create(&article1)
				tx.Create(&article2)
				tx.Create(&article3)
				tx.Create(&domain.Favorite{
					UserID:    user1.ID,
					ArticleID: article1.ID,
				})
				tx.Create(&domain.Favorite{
					UserID:    user1.ID,
					ArticleID: article3.ID,
				})
				return nil
			},
			thenFn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository) {
				follows, err := ar.FindFavorites(1, []uint{1, 2, 3})
				assert.NoError(t, err)
				assert.Equal(t, 2, len(follows))
				assert.Equal(t, uint(1), follows[0].ArticleID)
				assert.Equal(t, uint(3), follows[1].ArticleID)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f.expectGiven(tt.givenFn)
			f.run(tt.thenFn)
		})
	}
}
