package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	userv1 "github.com/azarenkov/ap2-final-gen/user/v1"
)

type UserHandler struct {
	client userv1.UserServiceClient
}

func NewUserHandler(c userv1.UserServiceClient) *UserHandler {
	return &UserHandler{client: c}
}

func (h *UserHandler) RegisterPublic(rg *gin.RouterGroup) {
	rg.POST("/users/register", h.register)
	rg.POST("/users/login", h.login)
	rg.POST("/users/reset-password", h.reset)
}

func (h *UserHandler) RegisterAuthenticated(rg *gin.RouterGroup) {
	rg.GET("/users/me", h.me)
	rg.PATCH("/users/me", h.update)
	rg.POST("/users/me/change-password", h.changePassword)
}

type registerBody struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
}

func (h *UserHandler) register(c *gin.Context) {
	var body registerBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := withTimeout(c, 5*time.Second)
	defer cancel()
	out, err := h.client.CreateUser(ctx, &userv1.CreateUserRequest{
		Email:    body.Email,
		Password: body.Password,
		FullName: body.FullName,
	})
	respond(c, out, err, http.StatusCreated)
}

type loginBody struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) login(c *gin.Context) {
	var body loginBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := withTimeout(c, 5*time.Second)
	defer cancel()
	out, err := h.client.LoginUser(ctx, &userv1.LoginUserRequest{Email: body.Email, Password: body.Password})
	respond(c, out, err, http.StatusOK)
}

func (h *UserHandler) reset(c *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := withTimeout(c, 5*time.Second)
	defer cancel()
	_, err := h.client.ResetPassword(ctx, &userv1.ResetPasswordRequest{Email: body.Email})
	respond(c, gin.H{"ok": true}, err, http.StatusAccepted)
}

func (h *UserHandler) me(c *gin.Context) {
	token := bearerToken(c)
	ctx, cancel := withTimeout(c, 5*time.Second)
	defer cancel()
	out, err := h.client.GetUserProfile(ctx, &userv1.GetUserProfileRequest{AccessToken: token})
	respond(c, out, err, http.StatusOK)
}

func (h *UserHandler) update(c *gin.Context) {
	var body struct {
		FullName string `json:"full_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token := bearerToken(c)
	ctx, cancel := withTimeout(c, 5*time.Second)
	defer cancel()
	out, err := h.client.UpdateUserProfile(ctx, &userv1.UpdateUserProfileRequest{
		AccessToken: token,
		FullName:    body.FullName,
	})
	respond(c, out, err, http.StatusOK)
}

func (h *UserHandler) changePassword(c *gin.Context) {
	var body struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token := bearerToken(c)
	ctx, cancel := withTimeout(c, 5*time.Second)
	defer cancel()
	_, err := h.client.ChangePassword(ctx, &userv1.ChangePasswordRequest{
		AccessToken: token,
		OldPassword: body.OldPassword,
		NewPassword: body.NewPassword,
	})
	respond(c, gin.H{"ok": true}, err, http.StatusOK)
}
