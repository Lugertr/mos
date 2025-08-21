package handler

import (
	"archive"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) createTag(c *gin.Context) {
	var input archive.TagCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input")
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		newErrorResponse(c, http.StatusBadRequest, "name required")
		return
	}

	id, err := h.services.Tags.CreateTag(c.Request.Context(), input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, map[string]interface{}{"id": id})
}

func (h *Handler) getAllTags(c *gin.Context) {
	items, err := h.services.Tags.GetAllTags(c.Request.Context())
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) getTagByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.services.Tags.GetTag(c.Request.Context(), id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) updateTag(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}
	var input archive.Tag
	if err := c.ShouldBindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input")
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		newErrorResponse(c, http.StatusBadRequest, "name required")
		return
	}
	if err := h.services.Tags.UpdateTag(c.Request.Context(), id, input); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) deleteTag(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.services.Tags.DeleteTag(c.Request.Context(), id); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}
