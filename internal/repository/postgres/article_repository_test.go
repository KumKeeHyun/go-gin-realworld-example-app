//go:build postgres

package postgres

import (
	"fmt"
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"sort"
	"testing"
)

type postgresFixture struct {
	t       *testing.T
	db      *gorm.DB
	givenFn func(tx *gorm.DB) error
}

func newPostgresFixture(t *testing.T) *postgresFixture {
	host := os.Getenv("TEST_POSTGRES_HOST")
	port := os.Getenv("TEST_POSTGRES_PORT")
	user := os.Getenv("TEST_POSTGRES_USER")
	pw := os.Getenv("TEST_POSTGRES_PW")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=realworld", host, port, user, pw)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(&domain.User{}, &domain.Follow{}, &domain.Article{}, &domain.Favorite{}, &domain.Comment{})
	if err != nil {
		t.Fatal(err)
	}
	f := &postgresFixture{
		t:  t,
		db: db,
	}
	t.Cleanup(func() {
		f.close()
	})
	return f
}

func (f *postgresFixture) expectGiven(fn func(tx *gorm.DB) error) {
	if fn == nil {
		f.givenFn = func(tx *gorm.DB) error {
			return nil
		}
		return
	}
	f.givenFn = fn
}

func (f *postgresFixture) run(fn func(t *testing.T, tx *gorm.DB)) {
	tx := f.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			f.t.Fatal(r)
		}
	}()

	err := f.givenFn(tx)
	assert.NoError(f.t, err)

	fn(f.t, tx)

	tx.Rollback()
}

func (f *postgresFixture) close() {
}

func Test_articleRepository_FindBySearchConditions(t *testing.T) {
	f := newPostgresFixture(t)

	tests := []struct {
		name    string
		givenFn func(tx *gorm.DB) error
		thenFn  func(t *testing.T, tx *gorm.DB)
	}{
		{
			name: "태그 검색",
			givenFn: func(tx *gorm.DB) error {
				user := &domain.User{Email: "test@example.com", Username: "test1"}
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
					Tags:        []string{"tag3"},
					Author:      domain.Author{ID: user.ID, Username: user.Username},
				}
				article3 := domain.Article{
					Slug:        "test3",
					Title:       "test3 title",
					Description: "test3 desc",
					Body:        "test3 body",
					Tags:        []string{"tag1", "tag3", "tag4", "tag5"},
					Author:      domain.Author{ID: user.ID, Username: user.Username},
				}
				tx.Create(&article1)
				tx.Create(&article2)
				tx.Create(&article3)
				return nil
			},
			thenFn: func(t *testing.T, tx *gorm.DB) {
				r := NewArticleRepository(tx)
				tag1 := "tag1"
				cond := ports.ArticleSearchConditions{
					Tag:      &tag1,
					Pageable: ports.Pageable{Limit: 20, Offset: 0},
				}
				articles, err := r.FindBySearchConditions(cond)
				if err != nil {
					return
				}
				assert.NoError(t, err)
				assert.Len(t, articles, 2)
				assert.Equal(t, "test1", articles[0].Slug)
				assert.Equal(t, "test3", articles[1].Slug)
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

func Test_articleRepository_FindTags(t *testing.T) {
	f := newPostgresFixture(t)

	tests := []struct {
		name    string
		givenFn func(tx *gorm.DB) error
		thenFn  func(t *testing.T, tx *gorm.DB)
	}{
		{
			name: "태그 조회",
			givenFn: func(tx *gorm.DB) error {
				user := &domain.User{Email: "test@example.com", Username: "test1"}
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
					Tags:        []string{"tag3"},
					Author:      domain.Author{ID: user.ID, Username: user.Username},
				}
				article3 := domain.Article{
					Slug:        "test3",
					Title:       "test3 title",
					Description: "test3 desc",
					Body:        "test3 body",
					Tags:        []string{"tag1", "tag3", "tag4", "tag5"},
					Author:      domain.Author{ID: user.ID, Username: user.Username},
				}
				tx.Create(&article1)
				tx.Create(&article2)
				tx.Create(&article3)
				return nil
			},
			thenFn: func(t *testing.T, tx *gorm.DB) {
				r := NewArticleRepository(tx)

				tags, err := r.FindTags()
				assert.NoError(t, err)
				sort.Strings(tags)
				assert.Len(t, tags, 5)
				assert.Equal(t, "tag1", tags[0])
				assert.Equal(t, "tag2", tags[1])
				assert.Equal(t, "tag3", tags[2])
				assert.Equal(t, "tag4", tags[3])
				assert.Equal(t, "tag5", tags[4])
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
