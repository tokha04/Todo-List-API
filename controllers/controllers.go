package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/tokha04/todo-list-api/database"
	"github.com/tokha04/todo-list-api/helpers"
	"github.com/tokha04/todo-list-api/models"
	"github.com/tokha04/todo-list-api/tokens"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.Client.Database("todo-list").Collection("users")
var todoCollection *mongo.Collection = database.Client.Database("todo-list").Collection("todos")
var validate = validator.New()

func Registration() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var c, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var user models.User
		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "could not bind json"})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "could not validate struct"})
			return
		}

		count, err := userCollection.CountDocuments(c, bson.M{"email": user.Email})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not count documents"})
			return
		}
		if count > 0 {
			ctx.JSON(http.StatusFound, gin.H{"error": "email already exists"})
			return
		}

		user.ID = primitive.NewObjectID()
		password := helpers.HashPassword(user.Password)
		user.Password = password

		token, refreshToken, _ := tokens.GeneratTokens(user.Name, user.Email, user.ID)
		user.Token = &token
		user.Refresh_Token = &refreshToken

		_, insertionErr := userCollection.InsertOne(c, user)
		if insertionErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not insert a user"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"token": *user.Token})
	}
}

func Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var c, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var user models.User
		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "could not bind json"})
			return
		}

		var foundUser models.User
		err := userCollection.FindOne(c, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "email is incorrect"})
			return
		}

		isPasswordValid := helpers.VerifyPassword(user.Password, foundUser.Password)
		if !isPasswordValid {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "password is incorrect"})
			return
		}

		token, refreshToken, _ := tokens.GeneratTokens(foundUser.Name, foundUser.Email, foundUser.ID)
		tokens.UpdateTokens(token, refreshToken, foundUser.ID)

		ctx.JSON(http.StatusOK, gin.H{"token": *foundUser.Token})
	}
}

func CreateItem() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var c, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var todo models.Todo
		if err := ctx.BindJSON(&todo); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "could not bind json"})
			return
		}

		err := validate.Struct(todo)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "could not validate a struct"})
			return
		}

		userID, ok := ctx.Get("user_id")
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		todo.ID = primitive.NewObjectID()
		todo.User_ID = userID.(primitive.ObjectID)
		todo.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		todo.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		_, err = todoCollection.InsertOne(c, todo)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not create a todo"})
			return
		}

		ctx.JSON(http.StatusOK, todo)
	}
}

func UpdateItem() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var c, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		id := ctx.Param("id")
		todoID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var todo models.Todo
		if err := ctx.BindJSON(&todo); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "could not bind json"})
			return
		}

		var existingTodo models.Todo
		err = todoCollection.FindOne(c, bson.M{"_id": todoID}).Decode(&existingTodo)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "could not fetch a todo"})
			return
		}

		var updateTodo primitive.D

		if todo.Title != "" {
			updateTodo = append(updateTodo, bson.E{Key: "title", Value: todo.Title})
		}
		if todo.Description != "" {
			updateTodo = append(updateTodo, bson.E{Key: "description", Value: todo.Description})
		}
		todo.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateTodo = append(updateTodo, bson.E{Key: "updated_at", Value: todo.Updated_At})

		userID, ok := ctx.Get("user_id")
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		if userID != existingTodo.User_ID {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "forbidden"})
			return
		}

		filter := bson.M{"_id": todoID}
		update := bson.M{"$set": updateTodo}
		res, err := todoCollection.UpdateOne(c, filter, update)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not update"})
			return
		}
		if res.MatchedCount == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "could not find a todo"})
			return
		}

		var updatedTodo models.Todo
		err = todoCollection.FindOne(c, bson.M{"_id": todoID}).Decode(&updatedTodo)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "could not fetch a todo"})
			return
		}

		ctx.JSON(http.StatusOK, updatedTodo)
	}
}

func DeleteItem() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var c, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		id := ctx.Param("id")
		todoID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var todo models.Todo
		err = todoCollection.FindOne(c, bson.M{"_id": todoID}).Decode(&todo)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "could not fetch a todo"})
			return
		}

		userID, ok := ctx.Get("user_id")
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		if userID != todo.User_ID {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "forbidden"})
			return
		}

		filter := bson.M{"_id": todoID}
		res, err := todoCollection.DeleteOne(c, filter)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete"})
			return
		}
		if res.DeletedCount == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "could not find a todo"})
			return
		}

		ctx.JSON(http.StatusNoContent, gin.H{"message": "successfully deleted"})
	}
}

func GetItems() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var c, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		userID, ok := ctx.Get("user_id")
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		res, err := todoCollection.Find(c, bson.M{"user_id": userID})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "todos not found"})
			return
		}

		var allTodos []models.Todo
		if err = res.All(c, &allTodos); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not decode todos"})
			return
		}

		page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
		if page < 1 {
			page = 1
		}

		limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
		if limit < 0 {
			limit = 10
		}

		start := (page - 1) * limit
		end := start + limit

		if start > len(allTodos) {
			start = len(allTodos)
		}
		if end > len(allTodos) {
			end = len(allTodos)
		}

		paginatedTodos := allTodos[start:end]
		response := gin.H{
			"data":  paginatedTodos,
			"page":  page,
			"limit": limit,
			"total": len(paginatedTodos),
		}

		ctx.JSON(http.StatusOK, response)
	}
}
