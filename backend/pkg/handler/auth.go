package handler

import (
	"archive"
	"net/http"

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
