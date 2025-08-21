package handler

import (
	"archive"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// createDocument — поддерживает multipart/form-data с GeoJSON
func (h *Handler) createDocument(c *gin.Context) {
	creatorID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "user not found")
		return
	}

	// собираем входные поля в DocumentCreateInput
	var in archive.DocumentCreateInput

	in.Title = c.PostForm("title")

	// privacy (optional)
	if p := c.PostForm("privacy"); p != "" {
		in.Privacy = archive.PrivacyType(p)
	}

	// document_date (optional)
	if v := c.PostForm("document_date"); v != "" {
		if t, err := parseDateFlexible(v); err == nil {
			in.DocumentDate = &t
		} else {
			newErrorResponse(c, http.StatusBadRequest, "invalid document_date")
			return
		}
	}

	// author_id -> AuthorID pointer
	if v := c.PostForm("author_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			in.AuthorID = &id
		}
	}
	// author_name
	if v := c.PostForm("author_name"); v != "" {
		in.AuthorName = &v
	}

	// document_type_id -> TypeID
	if v := c.PostForm("document_type_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			in.TypeID = &id
		}
	}

	// geojson (optional) — validate JSON
	if v := c.PostForm("geojson"); v != "" {
		raw := json.RawMessage([]byte(v))
		if !json.Valid(raw) {
			newErrorResponse(c, http.StatusBadRequest, "invalid geojson")
			return
		}
		in.GeoJSON = raw
	}

	// geom (optional) — attempt to set GeoJSON if valid
	if v := c.PostForm("geom"); v != "" {
		if json.Valid([]byte(v)) {
			raw := json.RawMessage([]byte(v))
			in.GeoJSON = raw
		}
	}

	// tags (optional) — comma separated
	if v := c.PostForm("tags"); v != "" {
		parts := strings.Split(v, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		in.Tags = parts
	}

	// file upload (optional) -> read bytes
	if fileHdr, err := c.FormFile("file"); err == nil {
		f, err := fileHdr.Open()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, "failed to open uploaded file")
			return
		}
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, "failed to read uploaded file")
			return
		}
		in.File = b
	}

	// creator id
	in.CreatorID = creatorID

	// Создаём через сервис
	id, err := h.services.Document.CreateDocument(c.Request.Context(), in)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, map[string]interface{}{"id": id})
}

// searchDocumentsByTag — поиск с фильтрами
func (h *Handler) searchDocumentsByTag(c *gin.Context) {
	filter := archive.DocumentSearchFilter{
		Tag:      c.Query("tag"),
		Author:   c.Query("author"),
		Type:     c.Query("type"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
	}
	if v := c.Query("limit"); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			filter.Limit = val
		}
	}
	if v := c.Query("offset"); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			filter.Offset = val
		}
	}

	items, err := h.services.Document.SearchDocumentsByTag(c.Request.Context(), filter)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, items)
}

// getDocumentByID
func (h *Handler) getDocumentByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}

	item, err := h.services.Document.GetDocumentByID(c.Request.Context(), id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, item)
}

// updateDocument — поддерживает multipart/form-data с GeoJSON
func (h *Handler) updateDocument(c *gin.Context) {
	updaterID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "user not found")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}

	var in archive.DocumentUpdateInput
	in.DocumentID = id
	in.UpdaterID = updaterID

	// optional fields — if present, set pointers
	if v := c.PostForm("title"); v != "" {
		in.Title = &v
	}

	if v := c.PostForm("privacy"); v != "" {
		p := archive.PrivacyType(v)
		in.Privacy = &p
	}

	if v := c.PostForm("document_date"); v != "" {
		if t, err := parseDateFlexible(v); err == nil {
			in.DocumentDate = &t
		} else {
			newErrorResponse(c, http.StatusBadRequest, "invalid document_date")
			return
		}
	}

	if v := c.PostForm("author_id"); v != "" {
		if idv, err := strconv.ParseInt(v, 10, 64); err == nil {
			in.AuthorID = &idv
		}
	}
	if v := c.PostForm("author_name"); v != "" {
		in.AuthorName = &v
	}

	if v := c.PostForm("document_type_id"); v != "" {
		if idv, err := strconv.ParseInt(v, 10, 64); err == nil {
			in.TypeID = &idv
		}
	}

	if v := c.PostForm("geojson"); v != "" {
		raw := json.RawMessage([]byte(v))
		if !json.Valid(raw) {
			newErrorResponse(c, http.StatusBadRequest, "invalid geojson")
			return
		}
		in.GeoJSON = &raw
	}

	if v := c.PostForm("tags"); v != "" {
		parts := strings.Split(v, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		in.Tags = &parts
	}

	// file upload
	if fileHdr, err := c.FormFile("file"); err == nil {
		f, err := fileHdr.Open()
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, "failed to open uploaded file")
			return
		}
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, "failed to read uploaded file")
			return
		}
		in.File = &b
	}

	if err := h.services.Document.UpdateDocument(c.Request.Context(), in.DocumentID, in); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

// deleteDocument
func (h *Handler) deleteDocument(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.services.Document.DeleteDocument(c.Request.Context(), id); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

// setDocumentPermission
func (h *Handler) setDocumentPermission(c *gin.Context) {
	docID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || docID <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "invalid document id")
		return
	}

	var input archive.DocumentPermission
	if err := c.ShouldBindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input")
		return
	}
	// Ensure document id correctness (overwrite if needed)
	input.DocumentID = docID

	if err := h.services.Document.SetDocumentPermission(c.Request.Context(), docID, input); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

// removeDocumentPermission
func (h *Handler) removeDocumentPermission(c *gin.Context) {
	docID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || docID <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "invalid document id")
		return
	}

	var input struct {
		TargetUserID int64 `json:"target_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input")
		return
	}

	if err := h.services.Document.RemoveDocumentPermission(c.Request.Context(), docID, input.TargetUserID); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}
