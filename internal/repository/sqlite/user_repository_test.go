//go:build sqlite
// +build sqlite

package sqlite

import (
	"database/sql"
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"github.com/KumKeeHyun/gin-realworld/pkg/crypto"
	"github.com/KumKeeHyun/gin-realworld/pkg/types"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"testing"
)

type sqliteFixture struct {
	t       *testing.T
	db      *gorm.DB
	givenFn func(tx *gorm.DB) error
}

func newSqliteFixture(t *testing.T) *sqliteFixture {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(&domain.User{}, &domain.Follow{}, &domain.Article{}, &domain.Favorite{}, &domain.Comment{})
	if err != nil {
		t.Fatal(err)
	}
	f := &sqliteFixture{
		t:  t,
		db: db,
	}
	t.Cleanup(func() {
		f.close()
	})
	return f
}

func (f *sqliteFixture) expectGiven(fn func(tx *gorm.DB) error) {
	if fn == nil {
		f.givenFn = func(tx *gorm.DB) error {
			return nil
		}
		return
	}
	f.givenFn = fn
}

func (f *sqliteFixture) run(fn func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository)) {
	tx := f.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			f.t.Fatal(r)
		}
	}()

	err := f.givenFn(tx)
	assert.NoError(f.t, err)

	fn(f.t, NewUserRepository(tx), NewArticleRepository(tx))

	tx.Rollback()
}

func (f *sqliteFixture) runWithComment(fn func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository, cr ports.CommentRepository)) {
	tx := f.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			f.t.Fatal(r)
		}
	}()

	err := f.givenFn(tx)
	assert.NoError(f.t, err)

	fn(f.t, NewUserRepository(tx), NewArticleRepository(tx), NewCommentRepository(tx))

	tx.Rollback()
}

func (f *sqliteFixture) close() {
	os.Remove("test.db")
}

func checkUser(t *testing.T, expected, actual domain.User) {
	if actual.Email != expected.Email ||
		actual.Username != expected.Username ||
		!crypto.CheckHashAndPassword(actual.Password.String, expected.Password.String) ||
		actual.Bio != expected.Bio ||
		actual.Image != expected.Image {
		t.Errorf("got = %v, want %v", actual, expected)
	}
}

func Test_sqliteRepository_Save(t *testing.T) {
	f := newSqliteFixture(t)

	tests := []struct {
		name    string
		givenFn func(tx *gorm.DB) error
		fn      func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository)
	}{
		{
			name: "save new user",
			fn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository) {
				newUser := domain.User{
					Email:    "test@example.com",
					Username: "test",
					Password: types.Password{String: "test password"},
					Bio:      "test bio",
					Image:    sql.NullString{},
				}

				saved, err := ur.Save(newUser)
				assert.NoError(t, err)

				saved, err = ur.FindByID(saved.ID)
				checkUser(t, newUser, saved)
			},
		},
		{
			name: "update user",
			givenFn: func(tx *gorm.DB) error {
				return tx.Create(&domain.User{
					Email:    "test@example.com",
					Username: "test",
					Password: types.Password{String: "test-password"},
				}).Error
			},
			fn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository) {
				user, err := ur.FindByUsername("test")
				assert.NoError(t, err)

				user.Email = "new@example.com"
				_, err = ur.Save(user)
				assert.NoError(t, err)

				updated, err := ur.FindByUsername("test")
				assert.NoError(t, err)
				assert.Equal(t, "new@example.com", updated.Email)
				assert.True(t, updated.ValidPassword("test-password"))
			},
		},
		{
			name: "duplicated email",
			givenFn: func(tx *gorm.DB) error {
				return tx.Create(&domain.User{
					Email:    "test@example.com",
					Username: "test",
				}).Error
			},
			fn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository) {
				newUser := domain.User{
					Email:    "test@example.com",
					Username: "test2",
				}
				_, err := ur.Save(newUser)
				assert.Error(t, err)
			},
		},
		{
			name: "duplicated username",
			givenFn: func(tx *gorm.DB) error {
				return tx.Create(&domain.User{
					Email:    "test1@example.com",
					Username: "test",
				}).Error
			},
			fn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository) {
				newUser := domain.User{
					Email:    "test2@example.com",
					Username: "test",
				}
				_, err := ur.Save(newUser)
				assert.Error(t, err)
			},
		},
		{
			name: "update password",
			givenFn: func(tx *gorm.DB) error {
				return tx.Create(&domain.User{
					Email:    "test@example.com",
					Username: "test",
					Password: types.Password{String: "test-password"},
				}).Error
			},
			fn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository) {
				user, err := ur.FindByEmail("test@example.com")
				assert.NoError(t, err)
				user.UpdatePassword("new-password")
				_, err = ur.Save(user)
				assert.NoError(t, err)

				updated, err := ur.FindByEmail("test@example.com")
				assert.NoError(t, err)
				assert.True(t, updated.ValidPassword("test-password"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f.expectGiven(tt.givenFn)
			f.run(tt.fn)
		})
	}
}

func Test_sqliteRepository_FindByID(t *testing.T) {
	f := newSqliteFixture(t)

	tests := []struct {
		name    string
		givenFn func(tx *gorm.DB) error
		fn      func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository)
	}{
		{
			name: "find email",
			givenFn: func(tx *gorm.DB) error {
				return tx.Create(&domain.User{
					Model:    gorm.Model{ID: 1},
					Email:    "test@example.com",
					Username: "test",
				}).Error
			},
			fn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository) {
				user, err := ur.FindByID(1)
				assert.NoError(t, err)
				assert.Equalf(t, uint(1), user.ID, "got = %v, want %v", user.ID, 1)
			},
		},
		{
			name: "not exists",
			fn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository) {
				_, err := ur.FindByID(1)
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f.expectGiven(tt.givenFn)
			f.run(tt.fn)
		})
	}
}

func Test_sqliteRepository_FindByEmail(t *testing.T) {
	f := newSqliteFixture(t)

	tests := []struct {
		name    string
		givenFn func(tx *gorm.DB) error
		fn      func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository)
	}{
		{
			name: "find email",
			givenFn: func(tx *gorm.DB) error {
				return tx.Create(&domain.User{
					Email:    "test@example.com",
					Username: "test",
				}).Error
			},
			fn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository) {
				user, err := ur.FindByEmail("test@example.com")
				assert.NoError(t, err)
				assert.Equalf(t, "test@example.com", user.Email, "got = %v, want %v", user.Email, "test@example.com")
			},
		},
		{
			name: "not exists",
			fn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository) {
				_, err := ur.FindByEmail("test@example.com")
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f.expectGiven(tt.givenFn)
			f.run(tt.fn)
		})
	}
}

func Test_sqliteRepository_FindByUsername(t *testing.T) {
	f := newSqliteFixture(t)

	tests := []struct {
		name    string
		givenFn func(tx *gorm.DB) error
		fn      func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository)
	}{
		{
			name: "find email",
			givenFn: func(tx *gorm.DB) error {
				return tx.Create(&domain.User{
					Email:    "test@example.com",
					Username: "test",
				}).Error
			},
			fn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository) {
				user, err := ur.FindByUsername("test")
				assert.NoError(t, err)
				assert.Equalf(t, "test", user.Username, "got = %v, want %v", user.Username, "test")
			},
		},
		{
			name: "not exists",
			fn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository) {
				_, err := ur.FindByUsername("test")
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f.expectGiven(tt.givenFn)
			f.run(tt.fn)
		})
	}
}

func Test_sqliteRepository_FindProfile(t *testing.T) {
	f := newSqliteFixture(t)

	tests := []struct {
		name    string
		givenFn func(tx *gorm.DB) error
		fn      func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository)
	}{
		{
			name: "find following profile",
			givenFn: func(tx *gorm.DB) error {
				follower := domain.User{Email: "test1@example.com", Username: "test1"}
				tx.Create(&follower)
				following := domain.User{Email: "test2@example.com", Username: "test2"}
				tx.Create(&following)
				return tx.Create(&domain.Follow{
					FollowerID:  follower.ID,
					FollowingID: following.ID,
				}).Error
			},
			fn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository) {
				profile, err := ur.FindProfile(1, 2)
				assert.NoError(t, err)
				assert.Equalf(t, "test2", profile.Username, "got = %v, want %v", profile.Username, "test2")
				assert.True(t, profile.Following)
			},
		},
		{
			name: "find unfollowing profile",
			givenFn: func(tx *gorm.DB) error {
				follower := domain.User{Email: "test1@example.com", Username: "test1"}
				tx.Create(&follower)
				following := domain.User{Email: "test2@example.com", Username: "test2"}
				tx.Create(&following)
				return tx.Create(&domain.Follow{
					FollowerID:  follower.ID,
					FollowingID: following.ID,
				}).Error
			},
			fn: func(t *testing.T, ur ports.UserRepository, ar ports.ArticleRepository) {
				profile, err := ur.FindProfile(2, 1)
				assert.NoError(t, err)
				assert.Equalf(t, "test1", profile.Username, "got = %v, want %v", profile.Username, "test1")
				assert.False(t, profile.Following)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f.expectGiven(tt.givenFn)
			f.run(tt.fn)
		})
	}
}
