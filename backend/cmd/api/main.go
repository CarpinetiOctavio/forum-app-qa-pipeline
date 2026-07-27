package main

import (
	"log"
	"net/http"

	"forum-app-ci-testing/internal/database"
	"forum-app-ci-testing/internal/handlers"
	"forum-app-ci-testing/internal/repository"
	"forum-app-ci-testing/internal/router"
	"forum-app-ci-testing/internal/services"
)

func main() {
	// Initialize database
	db, err := database.InitDB("./database.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Create repositories
	userRepo := repository.NewSQLiteUserRepository(db)
	postRepo := repository.NewSQLitePostRepository(db)

	// Create services
	authService := services.NewAuthService(userRepo)
	postService := services.NewPostService(postRepo, userRepo)

	// Create handlers
	authHandler := handlers.NewAuthHandler(authService)
	postHandler := handlers.NewPostHandler(postService)

	// Configure routes
	r := router.Setup(authHandler, postHandler)

	// Start server
	log.Println("🚀 Server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
