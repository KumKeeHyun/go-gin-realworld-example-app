package controller

import (
	"errors"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"github.com/KumKeeHyun/gin-realworld/internal/rest/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ProfileController struct {
	profileService ports.ProfileService
}

func NewProfileController(profileService ports.ProfileService) *ProfileController {
	return &ProfileController{profileService: profileService}
}

type ProfileUri struct {
	Username string `uri:"username" binding:"required"`
}

func (c *ProfileController) GetProfile(ctx *gin.Context) {
	claim, err := middleware.GetAccessClaim(ctx)
	if err != nil && !errors.Is(err, middleware.ErrClaimNotExists) {
		ctx.Error(err)
		return
	}

	var requestUri ProfileUri
	if err := ctx.ShouldBindUri(&requestUri); err != nil {
		ctx.Error(err)
		return
	}

	profile, err := c.profileService.Find(claim.UID, requestUri.Username)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, ProfileToResp(profile))
}

func (c *ProfileController) FollowUser(ctx *gin.Context) {
	claim, err := middleware.GetAccessClaim(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var requestUri ProfileUri
	if err := ctx.ShouldBindUri(&requestUri); err != nil {
		ctx.Error(err)
		return
	}

	profile, err := c.profileService.Follow(claim.UID, requestUri.Username)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, ProfileToResp(profile))
}

func (c *ProfileController) UnfollowUser(ctx *gin.Context) {
	claim, err := middleware.GetAccessClaim(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var requestUri ProfileUri
	if err := ctx.ShouldBindUri(&requestUri); err != nil {
		ctx.Error(err)
		return
	}

	profile, err := c.profileService.Unfollow(claim.UID, requestUri.Username)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, ProfileToResp(profile))
}
