package handler

import (
	"net/http"
	"strings"

	"archive"

	"github.com/gin-gonic/gin"
)

type signUpInput struct {
	Login    string  `json:"login" binding:"required"`
	Password string  `json:"password" binding:"required"`
	RoleID   *int64  `json:"role_id,omitempty"`
	FullName *string `json:"full_name,omitempty"`
}

func (h *Handler) signUp(c *gin.Context) {
	var input signUpInput
	if err := c.ShouldBindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	user := archive.User{
		Login:        input.Login,
		PasswordHash: input.Password, // сервис решит хешировать или вызвать fn_register_user
	}

	if input.RoleID != nil {
		user.RoleID = *input.RoleID
	}
	if input.FullName != nil {
		user.FullName = input.FullName
	}

	id, err := h.services.Authorization.CreateUser(c.Request.Context(), user)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, map[string]interface{}{"id": id})
}

type signInInput struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) signIn(c *gin.Context) {
	var input signInInput
	if err := c.ShouldBindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input")
		return
	}

	token, err := h.services.Authorization.GenerateToken(c.Request.Context(), input.Login, input.Password)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"token": token})
}

func (h *Handler) refreshToken(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if header == "" {
		newErrorResponse(c, http.StatusBadRequest, "empty auth header")
		return
	}
	parts := strings.Split(header, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		newErrorResponse(c, http.StatusBadRequest, "invalid auth header")
		return
	}
	oldToken := parts[1]
	if oldToken == "" {
		newErrorResponse(c, http.StatusBadRequest, "token is empty")
		return
	}

	newToken, err := h.services.Authorization.RefreshToken(c.Request.Context(), oldToken)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{"token": newToken})
}
