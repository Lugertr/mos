package handler

import (
	"archive"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) createDocumentType(c *gin.Context) {
	var input archive.DocumentTypeCreate
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input")
		return
	}
	if input.Name == "" {
		newErrorResponse(c, http.StatusBadRequest, "name required")
		return
	}
	id, err := h.services.DocumentTypes.CreateDocumentType(c.Request.Context(), input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"id": id})
}

func (h *Handler) getAllDocumentTypes(c *gin.Context) {
	items, err := h.services.DocumentTypes.GetAllDocumentTypes(c.Request.Context())
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) getDocumentTypeByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.services.DocumentTypes.GetDocumentType(c.Request.Context(), id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) updateDocumentType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}
	var input archive.DocumentType
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input")
		return
	}
	if input.Name == "" {
		newErrorResponse(c, http.StatusBadRequest, "name required")
		return
	}
	if err := h.services.DocumentTypes.UpdateDocumentType(c.Request.Context(), id, input); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) deleteDocumentType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.services.DocumentTypes.DeleteDocumentType(c.Request.Context(), id); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}
