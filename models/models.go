package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID            primitive.ObjectID `json:"_id" bson:"_id"`
	Name          string             `json:"name" validate:"required,min=2,max=30"`
	Email         string             `json:"email" validate:"email,required"`
	Password      string             `json:"password" validate:"required,min=6"`
	Token         *string            `json:"token"`
	Refresh_Token *string            `json:"refresh_token"`
}

type Todo struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	User_ID     primitive.ObjectID `json:"user_id"`
	Title       string             `json:"title" validate:"required"`
	Description string             `json:"description"`
	Created_At  time.Time          `json:"created_at"`
	Updated_At  time.Time          `json:"updated_at"`
}
