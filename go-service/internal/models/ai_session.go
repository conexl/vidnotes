package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AISession struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	VideoID   primitive.ObjectID `bson:"video_id" json:"video_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	Messages  []AIMessage        `bson:"messages" json:"messages"`
	Title     string             `bson:"title" json:"title"`
	Summary   string             `bson:"summary,omitempty" json:"summary,omitempty"`
}

type AIMessage struct {
	Role    string    `bson:"role" json:"role"`
	Content string    `bson:"content" json:"content"`
	Time    time.Time `bson:"time" json:"time"`
}

type AIRequest struct {
	SessionID primitive.ObjectID `json:"session_id"`
	Message   string             `json:"message"`
	VideoID   primitive.ObjectID `json:"video_id,omitempty"`
}

type AIResponse struct {
	Message   string    `json:"message"`
	SessionID string    `json:"session_id"`
	Time      time.Time `json:"time"`
}
