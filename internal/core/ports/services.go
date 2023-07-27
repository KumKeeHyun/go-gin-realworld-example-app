package ports

import (
	"errors"
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
)

var (
	ErrInternal                  = errors.New("internal error")
	ErrResourceNotFound          = errors.New("resource not found")
	ErrInvalidPassword           = errors.New("invalid password")
	ErrSelfFollowing             = errors.New("can not follow oneself")
	ErrDuplicatedEmailOrUsername = errors.New("duplicated email or username")
	ErrNonOwnedContent           = errors.New("user is not author of article")
)

type UserUpdateFields struct {
	Email    *string
	Username *string
	Password *string
	Bio      *string
	Image    *string
}

type AuthService interface {
	Transactional[AuthService]
	Register(email, username, password string) (domain.User, error)
	Login(email, password string) (domain.User, error)
	Update(userID uint, fields UserUpdateFields) (domain.User, error)
}

type ProfileService interface {
	Transactional[ProfileService]
	Find(curUserID uint, profileUsername string) (domain.Profile, error)
	Follow(curUserID uint, followingName string) (domain.Profile, error)
	Unfollow(curUserID uint, followingName string) (domain.Profile, error)
}

type ArticleUpdateFields struct {
	Title       *string
	Description *string
	Body        *string
}

type ArticleService interface {
	Transactional[ArticleService]
	Create(authorID uint, title, description, body string, tags []string) (domain.ArticleView, error)
	Find(readerID uint, slug string) (domain.ArticleView, error)
	ListByConditions(readerID uint, conditions ArticleSearchConditions) ([]domain.ArticleView, error)
	ListFeed(readerID uint, pageable Pageable) ([]domain.ArticleView, error)
	Update(authorID uint, slug string, fields ArticleUpdateFields) (domain.ArticleView, error)
	Delete(authorID uint, slug string) error
	Favorite(userID uint, slug string) (domain.ArticleView, error)
	Unfavorite(userID uint, slug string) (domain.ArticleView, error)
	ListTags() ([]string, error)
}

type CommentService interface {
	Transactional[CommentService]
	Create(authorID uint, slug string, body string) (domain.CommentView, error)
	GetFromArticle(readerID uint, slug string) ([]domain.CommentView, error)
	Delete(authorID, commentID uint) error
}
