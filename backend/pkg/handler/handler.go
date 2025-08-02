package handler

import (
	"center/pkg/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowHeaders = []string{"Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"}
	router.Use(cors.New(config))

	auth := router.Group("/auth")
	{
		auth.POST("/sign-up", h.signUp)
		auth.POST("/sign-in", h.signIn)
	}

	api := router.Group("/api", h.userIdentity)
	{
		LabServices := api.Group("/LabServices")
		{
			LabServices.POST("/", h.createLabService)
			LabServices.GET("/", h.getAllLabServices)
			LabServices.GET("/:id", h.getLabServiceById)
			LabServices.PUT("/:id", h.updateLabService)
			LabServices.DELETE("/:id", h.deleteLabService)
		}

		Patients := api.Group("/Patients")
		{
			Patients.POST("/", h.createPatient)
			Patients.GET("/", h.getAllPatients)
			Patients.GET("/:id", h.getPatientById)
			Patients.PUT("/:id", h.updatePatient)
			Patients.DELETE("/:id", h.deletePatient)
		}

		InsuranceCompanies := api.Group("/InsuranceCompanies")
		{
			InsuranceCompanies.POST("/", h.createInsuranceCompany)
			InsuranceCompanies.GET("/", h.getAllInsuranceCompanies)
			InsuranceCompanies.GET("/:id", h.getInsuranceCompanyById)
			InsuranceCompanies.PUT("/:id", h.updateInsuranceCompany)
			InsuranceCompanies.DELETE("/:id", h.deleteInsuranceCompany)
		}

		Orders := api.Group("/Orders")
		{
			Orders.POST("/", h.createOrder)
			Orders.GET("/", h.getAllOrders)
			Orders.GET("/:id", h.getOrderById)
			Orders.PUT("/:id", h.updateOrder)
			Orders.DELETE("/:id", h.deleteOrder)
		}

		ProvidedServices := api.Group("/ProvidedServices")
		{
			ProvidedServices.POST("/", h.createProvidedService)
			ProvidedServices.GET("/", h.getAllProvidedServices)
			ProvidedServices.GET("/:id", h.getProvidedServiceById)
			ProvidedServices.PUT("/:id", h.updateProvidedService)
			ProvidedServices.DELETE("/:id", h.deleteProvidedService)
		}

		Analyzer := api.Group("/Analyzers")
		{
			Analyzer.POST("/", h.createAnalyzer)
			Analyzer.GET("/", h.getAllAnalyzers)
			Analyzer.GET("/:id", h.getAnalyzerById)
			Analyzer.PUT("/:id", h.updateAnalyzer)
			Analyzer.DELETE("/:id", h.deleteAnalyzer)
		}

		/*
			LabTechnicians := api.Group("/LabTechnicians")
			{
				LabTechnicians.POST("/", h.LabTechnicians)
				LabTechnicians.GET("/", h.LabTechnicians)
				LabTechnicians.GET("/:id", h.LabTechnicians)
				LabTechnicians.PUT("/:id", h.updateLabTechnicians)
				LabTechnicians.DELETE("/:id", h.deleteLabTechnicians)
			}

			Accountants := api.Group("/Accountants")
			{
				Accountants.POST("/", h.Accountants)
				Accountants.GET("/", h.Accountants)
				Accountants.GET("/:id", h.Accountants)
				Accountants.PUT("/:id", h.updateAccountants)
				Accountants.DELETE("/:id", h.deleteAccountants)
			}

			Administrators := api.Group("/Administrators")
			{
				Administrators.POST("/", h.Administrators)
				Administrators.GET("/", h.Administrators)
				Administrators.GET("/:id", h.Administrators)
				Administrators.PUT("/:id", h.updateAdministrators)
				Administrators.DELETE("/:id", h.deleteAdministrators)
			}

			ArchivedData := api.Group("/ArchivedData")
			{
				ArchivedData.POST("/", h.ArchivedData)
				ArchivedData.GET("/", h.ArchivedData)
				ArchivedData.GET("/:id", h.ArchivedData)
				ArchivedData.PUT("/:id", h.updateArchivedData)
				ArchivedData.DELETE("/:id", h.deleteArchivedData)
			}

			FailedLoginAttempts := api.Group("/FailedLoginAttempts")
			{
				FailedLoginAttempts.POST("/", h.FailedLoginAttempts)
				FailedLoginAttempts.GET("/", h.FailedLoginAttempts)
				FailedLoginAttempts.GET("/:id", h.FailedLoginAttempts)
				FailedLoginAttempts.PUT("/:id", h.updateFailedLoginAttempts)
				FailedLoginAttempts.DELETE("/:id", h.deleteFailedLoginAttempts)
			}

			LoginHistory := api.Group("/LoginHistory")
			{
				LoginHistory.POST("/", h.LoginHistory)
				LoginHistory.GET("/", h.LoginHistory)
				LoginHistory.GET("/:id", h.LoginHistory)
				LoginHistory.PUT("/:id", h.updateLoginHistory)
				LoginHistory.DELETE("/:id", h.deleteLoginHistory)
			}
		*/
	}

	return router
}
