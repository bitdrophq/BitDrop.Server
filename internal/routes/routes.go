package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/richiethie/BitDrop.Server/internal/handlers"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")

	//Auth routes
	api.POST("/signup", handlers.SignUp)
	api.POST("/login", handlers.Login)
	api.POST("/logout", handlers.Logout)

	//Profile routes
	api.GET("/profile", handlers.GetProfile)

	// Availability check for username and email
	api.GET("/check-availability", handlers.CheckAvailability)

}
