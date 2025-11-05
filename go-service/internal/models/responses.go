// models/responses.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserResponse struct {
	ID            primitive.ObjectID `json:"id"`
	Email         string             `json:"email"`
	Name          string             `json:"name"`
	AvatarURL     string             `json:"avatar_url,omitempty"`
	Subscription  string             `json:"subscription"`
	AnalysesCount int                `json:"analyses_count"`
	CreatedAt     time.Time          `json:"created_at"`
}

type AnalyticsInfoResponse struct {
	Subscription       string    `json:"subscription"`
	DailyLimit         int       `json:"daily_limit"`
	DailyUsed          int       `json:"daily_used"`
	DailyRemaining     int       `json:"daily_remaining"`
	TotalAnalyses      int       `json:"total_analyses"`
	CanPerformAnalysis bool      `json:"can_perform_analysis"`
	NextReset          time.Time `json:"next_reset"`
	UsagePercentage    float64   `json:"usage_percentage"`
}
