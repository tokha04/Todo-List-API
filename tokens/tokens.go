package tokens

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/tokha04/todo-list-api/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SignedDetails struct {
	ID    primitive.ObjectID
	Name  string
	Email string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.Client.Database("todo-list").Collection("users")
var SECRET_KEY = os.Getenv("SECRET_KEY")

func GeneratTokens(name string, email string, id primitive.ObjectID) (string, string, error) {
	claims := &SignedDetails{
		ID:    id,
		Name:  name,
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return "", "", err
	}

	return token, refreshToken, err
}

func UpdateTokens(token string, refreshToken string, id primitive.ObjectID) {
	var c, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var updateObj primitive.D
	updateObj = append(updateObj, bson.E{Key: "token", Value: token})
	updateObj = append(updateObj, bson.E{Key: "refresh_token", Value: refreshToken})

	updatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: updatedAt})

	filter := bson.M{"_id": id}
	_, err := userCollection.UpdateOne(c, filter, bson.D{{Key: "$set", Value: updateObj}})
	if err != nil {
		log.Panic(err)
	}
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(signedToken, &SignedDetails{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "token is invalid"
		return
	}

	if claims.ExpiresAt < time.Now().Unix() {
		msg = "token is already expired"
		return
	}

	return claims, msg
}
