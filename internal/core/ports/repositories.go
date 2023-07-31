package ports

//go:generate mockgen -destination=./mock_ports/mock_repositories.go -package=mock_ports github.com/KumKeeHyun/gin-realworld/internal/core/ports UserRepository,ArticleRepository,CommentRepository

import (
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
)

type UserRepository interface {
	Transactional[UserRepository]
	Save(user domain.User) (domain.User, error)
	FindByID(id uint) (domain.User, error)
	FindByEmail(email string) (domain.User, error)
	FindByUsername(username string) (domain.User, error)
	FindByEmailOrUsername(email, username string) (domain.User, error)
	FindProfile(curUserID, profileUserID uint) (domain.Profile, error)
	CreateFollow(followerID, followingID uint) (domain.Follow, error)
	FindFollow(followerID, followingID uint) (domain.Follow, error)
	FindFollows(followerID uint, followingIDs []uint) ([]domain.Follow, error)
	DeleteFollow(followerID, followingID uint) error
}

type Pageable struct {
	Limit  int
	Offset int
}

type ArticleSearchConditions struct {
	Tag       *string
	Author    *string
	Favorited *string
	Pageable
}

type ArticleRepository interface {
	Transactional[ArticleRepository]
	Save(article domain.Article) (domain.Article, error)
	FindBySlug(slug string) (domain.Article, error)
	FindBySearchConditions(cond ArticleSearchConditions) ([]domain.Article, error)
	FindFeed(userID uint, pageable Pageable) ([]domain.Article, error)
	DeleteBySlug(slug string) error
	UpdateAuthorInfo(user domain.User) error
	CreateFavorite(userID, articleID uint) (domain.Favorite, error)
	FindFavorite(userID uint, articleID uint) (domain.Favorite, error)
	FindFavorites(userID uint, articleIDs []uint) ([]domain.Favorite, error)
	DeleteFavorite(userID, articleID uint) error
	FindTags() ([]string, error)
}

type CommentRepository interface {
	Transactional[CommentRepository]
	Save(comment domain.Comment) (domain.Comment, error)
	FindFromArticle(slug string) ([]domain.Comment, error)
	Delete(id, authorID uint) error
}
