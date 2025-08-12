package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "backend/docs"
	"backend/internal/handlers"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/routes"
	"backend/internal/services"
	"backend/pkg/database"
)

// @title User Management API
// @version 1.0
// @description A user management REST API with authentication and role-based access control
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Connect to database
	database.Connect()
	defer database.Close()

	// Run database migrations
	database.Migrate()

	// Initialize repositories
	userRepo := repository.NewUserRepository()
	packageRepo := repository.NewPackageRepository()
	businessRepo := repository.NewBusinessRepository()

	// Initialize services
	userService := services.NewUserService(userRepo)
	packageService := services.NewPackageService(packageRepo)
	businessService := services.NewBusinessService(businessRepo, userRepo, packageRepo)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	packageHandler := handlers.NewPackageHandler(packageService)
	businessHandler := handlers.NewBusinessHandler(businessService)

	r := gin.Default()

	// Add CORS middleware
	r.Use(middleware.CORSMiddleware())

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		// Setup user routes
		routes.SetupUserRoutes(api, userHandler)
		routes.SetupPackageRoutes(api, packageHandler)
		routes.SetupBusinessRoutes(api, businessHandler) // Add this line
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on port %s", port)
		log.Printf("Swagger docs available at: http://localhost:%s/swagger/index.html", port)
		if err := r.Run(":" + port); err != nil {
			log.Fatal("Server failed to start:", err)
		}
	}()

	<-c
	log.Println("Shutting down server...")
}
