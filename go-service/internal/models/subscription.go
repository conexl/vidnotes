// models/subscription.go
package models

import "time"

type SubscriptionConfig struct {
	MonthlyAnalyses int `json:"monthly_analyses"` // Месячный лимит анализов
}

type AnalyticsInfo struct {
	Subscription       string    `json:"subscription"`         // "free" или "premium"
	MonthlyLimit       int       `json:"monthly_limit"`        // Месячный лимит
	MonthlyUsed        int       `json:"monthly_used"`         // Использовано в этом месяце
	MonthlyRemaining   int       `json:"monthly_remaining"`    // Осталось в этом месяце
	TotalAnalyses      int       `json:"total_analyses"`       // Всего анализов за все время
	CanPerformAnalysis bool      `json:"can_perform_analysis"` // Может ли выполнить анализ
	NextReset          time.Time `json:"next_reset"`           // Время следующего сброса
	UsagePercentage    float64   `json:"usage_percentage"`     // Процент использования
	CurrentMonth       string    `json:"current_month"`        // Текущий месяц
}

var SubscriptionLimits = map[string]SubscriptionConfig{
	"free": {
		MonthlyAnalyses: 50, // 50 анализов в месяц
	},
	"premium": {
		MonthlyAnalyses: 500, // 500 анализов в месяц
	},
}
