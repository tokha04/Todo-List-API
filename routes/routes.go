package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tokha04/todo-list-api/controllers"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/register", controllers.Registration())
	incomingRoutes.POST("/login", controllers.Login())
}

func TodoRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/todos", controllers.CreateItem())
	incomingRoutes.PATCH("/todos/:id", controllers.UpdateItem())
	incomingRoutes.DELETE("/todos/:id", controllers.DeleteItem())
	incomingRoutes.GET("/todos", controllers.GetItems())
}
