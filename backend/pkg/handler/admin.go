package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// getLogsByUser ?target_user_id=&start=&end=
func (h *Handler) getLogsByUser(c *gin.Context) {
	adminID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "user not found")
		return
	}

	targetStr := c.Query("target_user_id")
	if targetStr == "" {
		newErrorResponse(c, http.StatusBadRequest, "target_user_id required")
		return
	}
	targetID, err := strconv.ParseInt(targetStr, 10, 64)
	if err != nil || targetID <= 0 {
		newErrorResponse(c, http.StatusBadRequest, "invalid target_user_id")
		return
	}

	var startPtr *time.Time
	var endPtr *time.Time
	if s := c.Query("start"); s != "" {
		if t, err := parseDateFlexible(s); err == nil {
			startPtr = &t
		} else {
			newErrorResponse(c, http.StatusBadRequest, "invalid start date")
			return
		}
	}
	if s := c.Query("end"); s != "" {
		if t, err := parseDateFlexible(s); err == nil {
			endPtr = &t
		} else {
			newErrorResponse(c, http.StatusBadRequest, "invalid end date")
			return
		}
	}

	logs, err := h.services.Admin.GetLogsByUser(c.Request.Context(), adminID, targetID, startPtr, endPtr)
	if err != nil {
		newErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}

	c.JSON(http.StatusOK, logs)
}

// getLogsByTable ?table=&start=&end=
func (h *Handler) getLogsByTable(c *gin.Context) {
	adminID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "user not found")
		return
	}

	table := c.Query("table")
	if table == "" {
		newErrorResponse(c, http.StatusBadRequest, "table param required")
		return
	}

	var startPtr *time.Time
	var endPtr *time.Time
	if s := c.Query("start"); s != "" {
		if t, err := parseDateFlexible(s); err == nil {
			startPtr = &t
		} else {
			newErrorResponse(c, http.StatusBadRequest, "invalid start date")
			return
		}
	}
	if s := c.Query("end"); s != "" {
		if t, err := parseDateFlexible(s); err == nil {
			endPtr = &t
		} else {
			newErrorResponse(c, http.StatusBadRequest, "invalid end date")
			return
		}
	}

	logs, err := h.services.Admin.GetLogsByTable(c.Request.Context(), adminID, table, startPtr, endPtr)
	if err != nil {
		newErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}
	c.JSON(http.StatusOK, logs)
}

// getLogsByDate ?start=&end=
func (h *Handler) getLogsByDate(c *gin.Context) {
	adminID, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "user not found")
		return
	}

	startStr := c.Query("start")
	endStr := c.Query("end")
	if startStr == "" || endStr == "" {
		newErrorResponse(c, http.StatusBadRequest, "start and end required")
		return
	}
	start, err := parseDateFlexible(startStr)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid start")
		return
	}
	end, err := parseDateFlexible(endStr)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid end")
		return
	}

	logs, err := h.services.Admin.GetLogsByDate(c.Request.Context(), adminID, start, end)
	if err != nil {
		newErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}
	c.JSON(http.StatusOK, logs)
}
