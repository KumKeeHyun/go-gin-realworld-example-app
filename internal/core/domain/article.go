package domain

import (
	"database/sql"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"time"
)

type Author struct {
	ID       uint
	Username string
	Bio      string
	Image    sql.NullString
}

type Article struct {
	gorm.Model
	Slug           string `gorm:"unique;index"`
	Title          string
	Description    string
	Body           string
	Tags           pq.StringArray `gorm:"type:text[]"`
	FavoritesCount int
	// Denormalize Article <-> User
	Author Author `gorm:"embedded;embeddedPrefix:author_"`
}

type Favorite struct {
	gorm.Model
	UserID    uint `gorm:"index:idx_user_article"`
	User      User
	ArticleID uint `gorm:"index:idx_user_article"`
	Article   Article
}

type Comment struct {
	gorm.Model
	Body      string
	ArticleID uint
	Article   Article
	// Denormalize Comment <-> User
	Author Author `gorm:"embedded;embeddedPrefix:author_"`
}

type ArticleView struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt

	Slug           string
	Title          string
	Description    string
	Body           string
	Tags           pq.StringArray
	Favorited      bool
	FavoritesCount int

	AuthorID        uint
	AuthorUsername  string
	AuthorBio       string
	AuthorImage     sql.NullString
	AuthorFollowing bool
}

func NewArticleView(article Article, favorited, following bool) ArticleView {
	return ArticleView{
		ID:              article.ID,
		CreatedAt:       article.CreatedAt,
		UpdatedAt:       article.UpdatedAt,
		DeletedAt:       article.DeletedAt,
		Slug:            article.Slug,
		Title:           article.Title,
		Description:     article.Description,
		Body:            article.Body,
		Tags:            article.Tags,
		Favorited:       favorited,
		FavoritesCount:  article.FavoritesCount,
		AuthorID:        article.Author.ID,
		AuthorUsername:  article.Author.Username,
		AuthorBio:       article.Author.Bio,
		AuthorImage:     article.Author.Image,
		AuthorFollowing: following,
	}
}

type CommentView struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
	Body      string

	ArticleID       uint
	AuthorID        uint
	AuthorUsername  string
	AuthorBio       string
	AuthorImage     sql.NullString
	AuthorFollowing bool
}

func NewCommentView(comment Comment, following bool) CommentView {
	return CommentView{
		ID:              comment.ID,
		CreatedAt:       comment.CreatedAt,
		UpdatedAt:       comment.UpdatedAt,
		DeletedAt:       comment.DeletedAt,
		Body:            comment.Body,
		ArticleID:       comment.ArticleID,
		AuthorID:        comment.Author.ID,
		AuthorUsername:  comment.Author.Username,
		AuthorBio:       comment.Author.Bio,
		AuthorImage:     comment.Author.Image,
		AuthorFollowing: following,
	}
}
