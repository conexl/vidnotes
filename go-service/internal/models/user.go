// models/user.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"-"`
	Name      string             `bson:"name" json:"name"`
	AvatarURL string             `bson:"avatar_url,omitempty" json:"avatar_url,omitempty"`

	// Система анализов
	AnalysesCount       int       `bson:"analyses_count" json:"analyses_count"`
	MonthlyAnalysesUsed int       `bson:"monthly_analyses_used" json:"monthly_analyses_used"`
	LastAnalysisDate    time.Time `bson:"last_analysis_date,omitempty" json:"last_analysis_date,omitempty"`
	LastResetMonth      int       `bson:"last_reset_month" json:"last_reset_month"`
	LastResetYear       int       `bson:"last_reset_year" json:"last_reset_year"`

	// Подписка (только free/premium)
	Subscription string `bson:"subscription" json:"subscription"`
	Role         string `bson:"role" json:"role"`

	// Таймстампы
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}
