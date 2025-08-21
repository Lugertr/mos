package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	authorizationHeader = "Authorization"
	userCtx             = "userId"
)

func (h *Handler) userIdentityMiddleware(c *gin.Context) {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		newErrorResponse(c, http.StatusUnauthorized, "empty auth header")
		c.Abort()
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		newErrorResponse(c, http.StatusUnauthorized, "invalid auth header")
		c.Abort()
		return
	}

	if len(headerParts[1]) == 0 {
		newErrorResponse(c, http.StatusUnauthorized, "token is empty")
		c.Abort()
		return
	}

	userId, err := h.services.Authorization.ParseToken(c, headerParts[1])
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		c.Abort()
		return
	}

	c.Set(userCtx, userId)
	c.Next()
}

func getUserId(c *gin.Context) (int64, error) {
	id, ok := c.Get(userCtx)
	if !ok {
		return 0, ErrUserNotFound
	}

	switch t := id.(type) {
	case int64:
		return t, nil
	case int:
		return int64(t), nil
	case *int64:
		if t == nil {
			return 0, ErrUserNotFound
		}
		return *t, nil
	default:
		return 0, ErrUserNotFound
	}
}
