package handler

import (
	"center"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) createInsuranceCompany(c *gin.Context) {
	var input center.InsuranceCompany
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.services.InsuranceCompany.Create(input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})
}

type getAllInsuranceCompaniesResponse struct {
	Data []center.InsuranceCompany `json:"data"`
}

func (h *Handler) getAllInsuranceCompanies(c *gin.Context) {
	InsuranceCompanies, err := h.services.InsuranceCompany.GetAll()
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, getAllInsuranceCompaniesResponse{
		Data: InsuranceCompanies,
	})
}

func (h *Handler) getInsuranceCompanyById(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	InsuranceCompanies, err := h.services.InsuranceCompany.GetById(id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, InsuranceCompanies)
}

func (h *Handler) updateInsuranceCompany(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	var input center.InsuranceCompanyUpdate
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.services.InsuranceCompany.Update(id, input); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{"ok"})
}

func (h *Handler) deleteInsuranceCompany(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	err = h.services.InsuranceCompany.Delete(id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{
		Status: "ok",
	})
}
