package handler

import (
	"archive/pkg/service"

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
		auth.POST("/refresh-token", h.refreshToken)
	}

	ref := router.Group("/api")
	{
		// document types
		ref.POST("/document_types", h.createDocumentType)
		ref.GET("/document_types", h.getAllDocumentTypes)
		ref.GET("/document_types/:id", h.getDocumentTypeByID)
		ref.PUT("/document_types/:id", h.updateDocumentType)
		ref.DELETE("/document_types/:id", h.deleteDocumentType)

		// tags
		ref.POST("/tags", h.createTag)
		ref.GET("/tags", h.getAllTags)
		ref.GET("/tags/:id", h.getTagByID)
		ref.PUT("/tags/:id", h.updateTag)
		ref.DELETE("/tags/:id", h.deleteTag)
	}

	// document endpoints (protected)
	docs := router.Group("/api/documents")
	docs.Use(h.userIdentityMiddleware)
	{
		docs.POST("", h.createDocument)
		docs.GET("", h.searchDocumentsByTag) // query params: tag=..., limit, offset, author, type, date_from, date_to
		docs.GET("/:id", h.getDocumentByID)
		docs.PUT("/:id", h.updateDocument)
		docs.DELETE("/:id", h.deleteDocument)

		// permission management (admin)
		docs.POST("/:id/permissions", h.setDocumentPermission)      // body: target_user_id, can_view, can_edit
		docs.DELETE("/:id/permissions", h.removeDocumentPermission) // body: target_user_id
	}

	// logs endpoints for admin
	logs := router.Group("/api/logs")
	logs.Use(h.userIdentityMiddleware)
	{
		logs.GET("/by_user", h.getLogsByUser)   // admin only
		logs.GET("/by_table", h.getLogsByTable) // admin only
		logs.GET("/by_date", h.getLogsByDate)   // admin only
	}

	return router
}
