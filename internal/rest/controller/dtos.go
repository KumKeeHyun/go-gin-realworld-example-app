package controller

import (
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
	"github.com/samber/lo"
	"time"
)

const (
	ISO8601      = "2006-01-02T15:04:05"
	JSON_ISO8601 = `"` + ISO8601 + `"`
)

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(time.Time(t).Format(JSON_ISO8601)), nil
}

func (t *JSONTime) UnmarshalJSON(data []byte) error {
	p, err := time.Parse(JSON_ISO8601, string(data))
	if err != nil {
		return err
	}
	*t = JSONTime(p)
	return nil
}

type UserResponse struct {
	User struct {
		Email    string  `json:"email"`
		Token    string  `json:"token"`
		Username string  `json:"username"`
		Bio      string  `json:"bio"`
		Image    *string `json:"image"`
	} `json:"user"`
}

func UserToResp(user domain.User) UserResponse {
	var resp UserResponse
	resp.User.Email = user.Email
	resp.User.Token = user.Token
	resp.User.Username = user.Username
	resp.User.Bio = user.Bio
	resp.User.Image = lo.If(user.Image.Valid, &user.Image.String).Else(nil)
	return resp
}

func ClaimToUserResp(claim domain.AccessClaim, token string) UserResponse {
	var resp UserResponse
	resp.User.Email = claim.Email
	resp.User.Token = token
	resp.User.Username = claim.Username
	resp.User.Bio = claim.Bio
	resp.User.Image = claim.Image
	return resp
}

type ProfileResponse struct {
	Profile struct {
		Username  string  `json:"username"`
		Bio       string  `json:"bio"`
		Image     *string `json:"image"`
		Following bool    `json:"following"`
	} `json:"profile"`
}

func ProfileToResp(profile domain.Profile) ProfileResponse {
	var resp ProfileResponse
	resp.Profile.Username = profile.Username
	resp.Profile.Bio = profile.Bio
	if profile.Image.Valid {
		resp.Profile.Image = &profile.Image.String
	}
	resp.Profile.Following = profile.Following
	return resp
}

type Article struct {
	Slug           string   `json:"slug"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Body           string   `json:"body"`
	TagList        []string `json:"tagList"`
	CreatedAt      JSONTime `json:"createdAt"`
	UpdatedAt      JSONTime `json:"updatedAt"`
	Favorited      bool     `json:"favorited"`
	FavoritesCount int      `json:"favoritesCount"`
	Author         struct {
		Username  string  `json:"username"`
		Bio       string  `json:"bio"`
		Image     *string `json:"image"`
		Following bool    `json:"following"`
	} `json:"author"`
}

func ArticleViewToDto(article domain.ArticleView) Article {
	var a Article
	a.Slug = article.Slug
	a.Title = article.Title
	a.Description = article.Description
	a.Body = article.Body
	a.TagList = article.Tags
	a.CreatedAt = JSONTime(article.CreatedAt)
	a.UpdatedAt = JSONTime(article.UpdatedAt)
	a.Favorited = article.Favorited
	a.FavoritesCount = article.FavoritesCount
	a.Author.Username = article.AuthorUsername
	a.Author.Bio = article.AuthorBio
	if article.AuthorImage.Valid {
		a.Author.Image = &article.AuthorImage.String
	}
	a.Author.Following = article.AuthorFollowing
	return a
}

type ArticleResponse struct {
	Article Article `json:"article"`
}

func ArticleViewToResponse(article domain.ArticleView) ArticleResponse {
	var resp ArticleResponse
	resp.Article = ArticleViewToDto(article)
	return resp
}

type MultipleArticlesResponse struct {
	Articles      []Article `json:"articles"`
	ArticlesCount int       `json:"articlesCount"`
}

func ArticlesToResponse(articles []domain.ArticleView) MultipleArticlesResponse {
	var resp MultipleArticlesResponse
	resp.Articles = lo.Map(articles, func(article domain.ArticleView, index int) Article {
		return ArticleViewToDto(article)
	})
	resp.ArticlesCount = len(articles)
	return resp
}

type Comment struct {
	Id        uint     `json:"id"`
	CreatedAt JSONTime `json:"createdAt"`
	UpdatedAt JSONTime `json:"updatedAt"`
	Body      string   `json:"body"`
	Author    struct {
		Username  string  `json:"username"`
		Bio       string  `json:"bio"`
		Image     *string `json:"image"`
		Following bool    `json:"following"`
	} `json:"author"`
}

func CommentViewToDto(comment domain.CommentView) Comment {
	var c Comment
	c.Id = comment.ID
	c.CreatedAt = JSONTime(comment.CreatedAt)
	c.UpdatedAt = JSONTime(comment.UpdatedAt)
	c.Body = comment.Body
	c.Author.Username = comment.AuthorUsername
	c.Author.Bio = comment.AuthorBio
	if comment.AuthorImage.Valid {
		c.Author.Image = &comment.AuthorImage.String
	}
	c.Author.Following = comment.AuthorFollowing
	return c
}

type CommentResponse struct {
	Comment Comment `json:"comment"`
}

func CommentViewToResponse(comment domain.CommentView) CommentResponse {
	var resp CommentResponse
	resp.Comment = CommentViewToDto(comment)
	return resp
}

type MultipleCommentsResponse struct {
	Comments []Comment `json:"comments"`
}

func CommentsToResponse(comments []domain.CommentView) MultipleCommentsResponse {
	var resp MultipleCommentsResponse
	resp.Comments = lo.Map(comments, func(comment domain.CommentView, index int) Comment {
		return CommentViewToDto(comment)
	})
	return resp
}

type TagsResponse struct {
	Tags []string `json:"tags"`
}
