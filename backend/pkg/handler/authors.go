package handler

import (
	"archive"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) createAuthor(c *gin.Context) {
	var input archive.AuthorCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input")
		return
	}
	input.FullName = strings.TrimSpace(input.FullName)
	if input.FullName == "" {
		newErrorResponse(c, http.StatusBadRequest, "full_name is required")
		return
	}

	id, err := h.services.Authors.CreateAuthor(c.Request.Context(), input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, map[string]interface{}{"id": id})
}

func (h *Handler) getAllAuthors(c *gin.Context) {
	items, err := h.services.Authors.GetAllAuthors(c.Request.Context())
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) getAuthorByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.services.Authors.GetAuthor(c.Request.Context(), id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) updateAuthor(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}
	var input archive.Author
	if err := c.ShouldBindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input")
		return
	}
	input.FullName = strings.TrimSpace(input.FullName)
	if input.FullName == "" {
		newErrorResponse(c, http.StatusBadRequest, "full_name is required")
		return
	}
	if err := h.services.Authors.UpdateAuthor(c.Request.Context(), id, input); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) deleteAuthor(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.services.Authors.DeleteAuthor(c.Request.Context(), id); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}
