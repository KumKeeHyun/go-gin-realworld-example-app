package controller

import (
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"github.com/KumKeeHyun/gin-realworld/internal/rest/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ArticleController struct {
	articleService ports.ArticleService
}

func NewArticleController(articleService ports.ArticleService) *ArticleController {
	return &ArticleController{
		articleService: articleService,
	}
}

type CreateArticleRequest struct {
	Article struct {
		Title       string   `json:"title" binding:"required"`
		Description string   `json:"description" binding:"required"`
		Body        string   `json:"body" binding:"required"`
		TagList     []string `json:"tagList"`
	} `json:"article" binding:"required"`
}

func (c *ArticleController) CreateArticle(ctx *gin.Context) {
	claim, err := middleware.GetAccessClaim(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	request := CreateArticleRequest{}
	if err := ctx.ShouldBind(&request); err != nil {
		ctx.Error(err)
		return
	}

	created, err := c.articleService.Create(
		claim.UID,
		request.Article.Title,
		request.Article.Description,
		request.Article.Body,
		request.Article.TagList,
	)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusCreated, ArticleViewToResponse(created))
}

type ListArticlesQuery struct {
	Tag       *string `form:"tag"`
	Author    *string `form:"author"`
	Favorited *string `form:"favorited"`
	Limit     int     `form:"limit,default=20"`
	Offset    int     `form:"offset,default=0"`
}

func (q ListArticlesQuery) ToSearchConditions() ports.ArticleSearchConditions {
	return ports.ArticleSearchConditions{
		Tag:       q.Tag,
		Author:    q.Author,
		Favorited: q.Favorited,
		Pageable: ports.Pageable{
			Limit:  q.Limit,
			Offset: q.Offset,
		},
	}
}

func (c *ArticleController) ListArticles(ctx *gin.Context) {
	claim, _ := middleware.GetAccessClaim(ctx)

	request := ListArticlesQuery{}
	if err := ctx.ShouldBindQuery(&request); err != nil {
		ctx.Error(err)
		return
	}

	articles, err := c.articleService.ListByConditions(claim.UID, request.ToSearchConditions())
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, ArticlesToResponse(articles))
}

type FeedArticlesQuery struct {
	Limit  int `form:"limit,default=20"`
	Offset int `form:"offset,default=0"`
}

func (q FeedArticlesQuery) ToPageable() ports.Pageable {
	return ports.Pageable{
		Limit:  q.Limit,
		Offset: q.Offset,
	}
}

func (c *ArticleController) FeedArticles(ctx *gin.Context) {
	claim, err := middleware.GetAccessClaim(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	request := FeedArticlesQuery{}
	if err := ctx.ShouldBindQuery(&request); err != nil {
		ctx.Error(err)
		return
	}

	articles, err := c.articleService.ListFeed(claim.UID, request.ToPageable())
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, ArticlesToResponse(articles))
}

type ArticleUri struct {
	Slug string `uri:"slug" binding:"required"`
}

func (c *ArticleController) GetArticle(ctx *gin.Context) {
	claim, _ := middleware.GetAccessClaim(ctx)

	var requestUri ArticleUri
	if err := ctx.ShouldBindUri(&requestUri); err != nil {
		ctx.Error(err)
		return
	}

	article, err := c.articleService.Find(claim.UID, requestUri.Slug)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusCreated, ArticleViewToResponse(article))
}

type UpdateArticleRequest struct {
	Article struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Body        *string `json:"body"`
	} `json:"article" binding:"required"`
}

func (r UpdateArticleRequest) ToArticleUpdateFields() ports.ArticleUpdateFields {
	return ports.ArticleUpdateFields{
		Title:       r.Article.Title,
		Description: r.Article.Description,
		Body:        r.Article.Body,
	}
}

func (c *ArticleController) UpdateArticle(ctx *gin.Context) {
	claim, err := middleware.GetAccessClaim(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var requestUri ArticleUri
	if err := ctx.ShouldBindUri(&requestUri); err != nil {
		ctx.Error(err)
		return
	}
	var request UpdateArticleRequest
	if err := ctx.ShouldBind(&request); err != nil {
		ctx.Error(err)
		return
	}

	updated, err := c.articleService.Update(claim.UID, requestUri.Slug, request.ToArticleUpdateFields())
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, ArticleViewToResponse(updated))
}

func (c *ArticleController) DeleteArticle(ctx *gin.Context) {
	claim, err := middleware.GetAccessClaim(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var requestUri ArticleUri
	if err := ctx.ShouldBindUri(&requestUri); err != nil {
		ctx.Error(err)
		return
	}

	err = c.articleService.Delete(claim.UID, requestUri.Slug)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.Status(http.StatusOK)
}

func (c *ArticleController) FavoriteArticle(ctx *gin.Context) {
	claim, err := middleware.GetAccessClaim(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var requestUri ArticleUri
	if err := ctx.ShouldBindUri(&requestUri); err != nil {
		ctx.Error(err)
		return
	}

	article, err := c.articleService.Favorite(claim.UID, requestUri.Slug)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, ArticleViewToResponse(article))
}

func (c *ArticleController) UnfavoriteArticle(ctx *gin.Context) {
	claim, err := middleware.GetAccessClaim(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var requestUri ArticleUri
	if err := ctx.ShouldBindUri(&requestUri); err != nil {
		ctx.Error(err)
		return
	}

	article, err := c.articleService.Unfavorite(claim.UID, requestUri.Slug)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, ArticleViewToResponse(article))
}

func (c *ArticleController) GetTags(ctx *gin.Context) {
	tags, err := c.articleService.ListTags()
	if err != nil {
		ctx.Error(err)
		return
	}

	var tagResponse TagsResponse
	tagResponse.Tags = tags
	ctx.JSON(http.StatusOK, tagResponse)
}
