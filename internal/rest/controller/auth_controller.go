package controller

import (
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"github.com/KumKeeHyun/gin-realworld/internal/rest/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthController struct {
	authService ports.AuthService
}

func NewAuthController(authService ports.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

type AuthenticateUserRequest struct {
	User struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	} `json:"user" binding:"required"`
}

func (c *AuthController) AuthenticateUser(ctx *gin.Context) {
	request := AuthenticateUserRequest{}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		return
	}
	user, err := c.authService.Login(request.User.Email, request.User.Password)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, UserToResp(user))
}

type RegisterUserRequest struct {
	User struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	} `json:"user" binding:"required"`
}

func (c *AuthController) RegisterUser(ctx *gin.Context) {
	request := RegisterUserRequest{}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		return
	}

	user, err := c.authService.Register(request.User.Email, request.User.Username, request.User.Password)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, UserToResp(user))
}

func (c *AuthController) GetCurrentUser(ctx *gin.Context) {
	claim, err := middleware.GetAccessClaim(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	token, err := middleware.GetToken(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, ClaimToUserResp(claim, token))
}

type UpdateUserRequest struct {
	User struct {
		Email    *string `json:"email"`
		Username *string `json:"username"`
		Password *string `json:"password"`
		Bio      *string `json:"bio"`
		Image    *string `json:"image"`
	} `json:"user" binding:"required"`
}

func (c *AuthController) UpdateUser(ctx *gin.Context) {
	claim, err := middleware.GetAccessClaim(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	tx, err := middleware.GetTransaction(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	request := UpdateUserRequest{}
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		return
	}
	fields := ports.UserUpdateFields{
		Email:    request.User.Email,
		Username: request.User.Username,
		Password: request.User.Password,
		Bio:      request.User.Bio,
		Image:    request.User.Image,
	}

	user, err := c.authService.WithTx(tx).Update(claim.UID, fields)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, UserToResp(user))
}
