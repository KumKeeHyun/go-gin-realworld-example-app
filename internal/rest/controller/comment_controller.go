package controller

import (
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"github.com/KumKeeHyun/gin-realworld/internal/rest/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

type CommentController struct {
	commentService ports.CommentService
}

func NewCommentController(commentService ports.CommentService) *CommentController {
	return &CommentController{
		commentService: commentService,
	}
}

type AddCommentRequest struct {
	Comment struct {
		Body string `json:"body" binding:"required"`
	} `json:"comment" binding:"required"`
}

func (c *CommentController) AddCommentToArticle(ctx *gin.Context) {
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
	var request AddCommentRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		return
	}

	created, err := c.commentService.Create(claim.UID, requestUri.Slug, request.Comment.Body)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusCreated, CommentViewToResponse(created))
}

func (c *CommentController) GetCommentsFromArticle(ctx *gin.Context) {
	claim, _ := middleware.GetAccessClaim(ctx)

	var requestUri ArticleUri
	if err := ctx.ShouldBindUri(&requestUri); err != nil {
		ctx.Error(err)
		return
	}
	comments, err := c.commentService.GetFromArticle(claim.UID, requestUri.Slug)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, CommentsToResponse(comments))
}

type CommentUri struct {
	Slug string `uri:"slug" binding:"required"`
	ID   uint   `uri:"id" binding:"required"`
}

func (c *CommentController) DeleteComment(ctx *gin.Context) {
	claim, err := middleware.GetAccessClaim(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var requestUri CommentUri
	if err := ctx.ShouldBindUri(&requestUri); err != nil {
		ctx.Error(err)
		return
	}

	err = c.commentService.Delete(claim.UID, requestUri.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.Status(http.StatusOK)
}
