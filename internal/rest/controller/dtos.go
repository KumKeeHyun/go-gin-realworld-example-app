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

type ArticleResponse struct {
	Article struct {
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
	} `json:"article"`
}

func ArticleViewToResponse(article domain.ArticleView) ArticleResponse {
	var resp ArticleResponse
	resp.Article.Slug = article.Slug
	resp.Article.Title = article.Title
	resp.Article.Description = article.Description
	resp.Article.Body = article.Body
	resp.Article.TagList = article.Tags
	resp.Article.CreatedAt = JSONTime(article.CreatedAt)
	resp.Article.UpdatedAt = JSONTime(article.UpdatedAt)
	resp.Article.Favorited = article.Favorited
	resp.Article.FavoritesCount = article.FavoritesCount
	resp.Article.Author.Username = article.AuthorUsername
	resp.Article.Author.Bio = article.AuthorBio
	if article.AuthorImage.Valid {
		resp.Article.Author.Image = &article.AuthorImage.String
	}
	resp.Article.Author.Following = article.AuthorFollowing
	return resp
}

type MultipleArticlesResponse struct {
	Articles      []ArticleResponse `json:"articles"`
	ArticlesCount int               `json:"articlesCount"`
}

func ArticlesToResponse(articles []domain.ArticleView) MultipleArticlesResponse {
	var resp MultipleArticlesResponse
	resp.Articles = lo.Map(articles, func(article domain.ArticleView, index int) ArticleResponse {
		return ArticleViewToResponse(article)
	})
	resp.ArticlesCount = len(articles)
	return resp
}

type CommentResponse struct {
	Comment struct {
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
	} `json:"comment"`
}

func CommentViewToResponse(comment domain.CommentView) CommentResponse {
	var resp CommentResponse
	resp.Comment.Id = comment.ID
	resp.Comment.CreatedAt = JSONTime(comment.CreatedAt)
	resp.Comment.UpdatedAt = JSONTime(comment.UpdatedAt)
	resp.Comment.Body = comment.Body
	resp.Comment.Author.Username = comment.AuthorUsername
	resp.Comment.Author.Bio = comment.AuthorBio
	if comment.AuthorImage.Valid {
		resp.Comment.Author.Image = &comment.AuthorImage.String
	}
	resp.Comment.Author.Following = comment.AuthorFollowing
	return resp
}

type MultipleCommentsResponse struct {
	Comments []CommentResponse `json:"comments"`
}

func CommentsToResponse(comments []domain.CommentView) MultipleCommentsResponse {
	var resp MultipleCommentsResponse
	resp.Comments = lo.Map(comments, func(comment domain.CommentView, index int) CommentResponse {
		return CommentViewToResponse(comment)
	})
	return resp
}

type TagsResponse struct {
	Tags []string `json:"tags"`
}
