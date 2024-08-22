package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/tokha04/todo-list-api/middleware"
	"github.com/tokha04/todo-list-api/routes"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("could not load .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	router := gin.Default()
	routes.UserRoutes(router)
	router.Use(middleware.Authentication())
	routes.TodoRoutes(router)

	log.Fatal(router.Run(":" + port))
}
