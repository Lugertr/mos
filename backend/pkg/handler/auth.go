package handler

import (
	"net/http"
	"strconv"
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

type updateFullNameInput struct {
	FullName string `json:"full_name" binding:"required"`
}

// PUT /api/users/full_name
// current user is taken from token (getUserId)
func (h *Handler) updateUserFullName(c *gin.Context) {
	requesterID, err := getUserId(c)
	if err != nil || requesterID == 0 {
		newErrorResponse(c, http.StatusUnauthorized, "user not authorized")
		return
	}

	var in updateFullNameInput
	if err := c.ShouldBindJSON(&in); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input")
		return
	}

	// call service: requester changes own full_name
	if err := h.services.Authorization.UpdateUserFullName(c.Request.Context(), requesterID, requesterID, in.FullName); err != nil {
		// можно расширить маппинг ошибок по тексту
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// body: { "old_password": "...", "new_password": "..." }
// old_password required for non-admins (enforced in DB logic)
type changePasswordInput struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password" binding:"required"`
}

// PUT /api/users/password
// current user is taken from token (getUserId)
func (h *Handler) changeUserPassword(c *gin.Context) {
	requesterID, err := getUserId(c)
	if err != nil || requesterID == 0 {
		newErrorResponse(c, http.StatusUnauthorized, "user not authorized")
		return
	}

	var in changePasswordInput
	if err := c.ShouldBindJSON(&in); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input")
		return
	}

	// service will hash passwords and call DB fn_change_user_password
	if err := h.services.Authorization.ChangeUserPassword(c.Request.Context(), requesterID, requesterID, in.OldPassword, in.NewPassword); err != nil {
		// mismatch old password -> возвращаем 400
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) getUsers(c *gin.Context) {
	// require authentication (middleware sets user id in context via getUserId)
	_, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "user not authorized")
		return
	}

	idsParam := c.Query("ids")
	var ids []int64
	if idsParam != "" {
		parts := strings.Split(idsParam, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			id, err := strconv.ParseInt(p, 10, 64)
			if err != nil {
				newErrorResponse(c, http.StatusBadRequest, "invalid ids param")
				return
			}
			ids = append(ids, id)
		}
	} else {
		// ids == nil -> repository will fetch all
		ids = nil
	}

	users, err := h.services.Authorization.GetUsersByIDs(c.Request.Context(), ids)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// response format: array of { id, full_name }
	out := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		m := map[string]interface{}{"id": u.ID}
		if u.FullName != nil {
			m["full_name"] = *u.FullName
		} else {
			m["full_name"] = nil
		}
		out = append(out, m)
	}

	c.JSON(http.StatusOK, out)
}
