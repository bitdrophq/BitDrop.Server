package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/richiethie/BitDrop.Server/internal/handlers"
	"github.com/richiethie/BitDrop.Server/internal/middleware"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")

	// Public routes
	api.POST("/signup", handlers.SignUp)
	api.POST("/login", handlers.Login)
	api.POST("/logout", handlers.Logout)
	api.GET("/check-availability", handlers.CheckAvailability)

	// Protected routes
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())

	protected.GET("/profile", handlers.GetProfile)
	protected.POST("/drops/upload", handlers.UploadDropHandler)
    protected.GET("/drops/user", handlers.GetUserDropsHandler)
	protected.GET("/drops/:id/details", handlers.GetDropDetailsHandler)
    protected.DELETE("/drops/:id", handlers.DeleteDropHandler)
}
