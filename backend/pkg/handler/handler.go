package handler

import (
	"hotel/pkg/service"

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
		clients := api.Group("/clients")
		{
			clients.POST("/", h.createClient)
			clients.GET("/", h.getAllClients)
			clients.GET("/:id", h.getClientById)
			clients.PUT("/:id", h.updateClient)
			clients.DELETE("/:id", h.deleteClient)
		}

		app := api.Group("/app")
		{
			app.POST("/", h.createApp)
			app.GET("/", h.getAllApps)
			app.GET("/:id", h.getAppById)
			app.PUT("/:id", h.updateApp)
			app.DELETE("/:id", h.deleteApp)
		}

		appType := api.Group("/appType")
		{
			appType.POST("/", h.createAppType)
			appType.GET("/", h.getAllAppTypes)
			appType.GET("/:id", h.getAppTypeById)
			appType.PUT("/:id", h.updateAppType)
			appType.DELETE("/:id", h.deleteAppType)
		}

		appService := api.Group("/appService")
		{
			appService.POST("/", h.createAppService)
			appService.GET("/", h.getAllAppServices)
			appService.GET("/:id", h.getAppServiceById)
			appService.PUT("/:id", h.updateAppService)
			appService.DELETE("/:id", h.deleteAppService)
		}

		appServiceType := api.Group("/appServiceType")
		{
			appServiceType.POST("/", h.createAppServiceType)
			appServiceType.GET("/", h.getAllAppServiceTypes)
			appServiceType.GET("/:id", h.getAppServiceTypeById)
			appServiceType.PUT("/:id", h.updateAppServiceType)
			appServiceType.DELETE("/:id", h.deleteAppServiceType)
		}

	}

	return router
}
