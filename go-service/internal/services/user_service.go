package services

import (
	"context"
	"fmt"
	"time"

	"github.com/code-zt/vidnotes/internal/models"
	"github.com/code-zt/vidnotes/internal/repository"
	"github.com/code-zt/vidnotes/pkg/password"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService interface {
	Register(ctx context.Context, req models.CreateUserRequest) (primitive.ObjectID, error)
	Login(ctx context.Context, req models.LoginRequest) (primitive.ObjectID, error)
	GetProfile(ctx context.Context, userID primitive.ObjectID) (*models.User, error)
	UpdateProfile(ctx context.Context, userID primitive.ObjectID, req models.UpdateUserRequest) error
	ChangePassword(ctx context.Context, userID primitive.ObjectID, oldPassword, newPassword string) error
	Delete(ctx context.Context, userID primitive.ObjectID) error
	CanPerformAnalysis(ctx context.Context, userID primitive.ObjectID) error
	RecordAnalysis(ctx context.Context, userID primitive.ObjectID) error
	GetAnalyticsInfo(ctx context.Context, userID primitive.ObjectID) (*models.AnalyticsInfo, error)
	ChangeSubscription(ctx context.Context, userID primitive.ObjectID, subscription string) error
	GetSubscriptionLimits(subscription string) models.SubscriptionConfig
	IncrementAnalysesCount(ctx context.Context, userID primitive.ObjectID) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) Register(ctx context.Context, req models.CreateUserRequest) (primitive.ObjectID, error) {
	existingUser, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return primitive.NilObjectID, err
	}
	if existingUser != nil {
		return primitive.NilObjectID, models.ErrUserAlreadyExists
	}

	hashedPassword, err := password.HashPassword(req.Password)
	if err != nil {
		return primitive.NilObjectID, err
	}
	now := time.Now()
	user := &models.User{
		Email:               req.Email,
		Password:            hashedPassword,
		Name:                req.Name,
		AvatarURL:           req.AvatarURL,
		Subscription:        "free",
		AnalysesCount:       0,
		MonthlyAnalysesUsed: 0,
		LastResetMonth:      int(now.Month()),
		LastResetYear:       now.Year(),
		CreatedAt:           now,
	}

	createdUserID, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return createdUserID, nil
}

func (s *userService) Login(ctx context.Context, req models.LoginRequest) (primitive.ObjectID, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return primitive.NilObjectID, err
	}
	if user == nil {
		return primitive.NilObjectID, models.ErrUserNotFound
	}

	if !password.CheckPassword(req.Password, user.Password) {
		return primitive.NilObjectID, models.ErrInvalidPassword
	}

	return user.ID, nil
}

func (s *userService) GetProfile(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.Password = ""
	return user, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID primitive.ObjectID, req models.UpdateUserRequest) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (s *userService) ChangePassword(ctx context.Context, userID primitive.ObjectID, oldPassword string, newPassword string) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if !password.CheckPassword(oldPassword, user.Password) {
		return models.ErrInvalidPassword
	}
	hashedPassword, err := password.HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	return s.userRepo.UpdateUser(ctx, user)
}

func (s *userService) Delete(ctx context.Context, userID primitive.ObjectID) error {
	return s.userRepo.DeleteUser(ctx, userID)
}

func (s *userService) checkAndResetMonthlyLimits(user *models.User) bool {
	now := time.Now()
	currentMonth := int(now.Month())
	currentYear := now.Year()

	if user.LastResetMonth != currentMonth || user.LastResetYear != currentYear {
		user.MonthlyAnalysesUsed = 0
		user.LastResetMonth = currentMonth
		user.LastResetYear = currentYear
		return true
	}
	return false
}

func (s *userService) CanPerformAnalysis(ctx context.Context, userID primitive.ObjectID) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	s.checkAndResetMonthlyLimits(user)

	limits, exists := models.SubscriptionLimits[user.Subscription]
	if !exists {
		limits = models.SubscriptionLimits["free"]
	}

	if user.MonthlyAnalysesUsed >= limits.MonthlyAnalyses {
		return models.ErrMonthlyAnalysesLimitExceeded
	}

	return nil
}

func (s *userService) RecordAnalysis(ctx context.Context, userID primitive.ObjectID) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	user.MonthlyAnalysesUsed++
	user.AnalysesCount++
	user.LastAnalysisDate = time.Now()

	return s.userRepo.UpdateUser(ctx, user)
}

func (s *userService) GetAnalyticsInfo(ctx context.Context, userID primitive.ObjectID) (*models.AnalyticsInfo, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	needsUpdate := s.checkAndResetMonthlyLimits(user)
	if needsUpdate {
		if err := s.userRepo.UpdateUser(ctx, user); err != nil {
			return nil, err
		}
	}

	limits, exists := models.SubscriptionLimits[user.Subscription]
	if !exists {
		limits = models.SubscriptionLimits["free"]
	}

	now := time.Now()
	nextReset := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	usagePercentage := float64(user.MonthlyAnalysesUsed) / float64(limits.MonthlyAnalyses) * 100

	return &models.AnalyticsInfo{
		Subscription:       user.Subscription,
		MonthlyLimit:       limits.MonthlyAnalyses,
		MonthlyUsed:        user.MonthlyAnalysesUsed,
		MonthlyRemaining:   limits.MonthlyAnalyses - user.MonthlyAnalysesUsed,
		TotalAnalyses:      user.AnalysesCount,
		CanPerformAnalysis: user.MonthlyAnalysesUsed < limits.MonthlyAnalyses,
		NextReset:          nextReset,
		UsagePercentage:    usagePercentage,
		CurrentMonth:       now.Format("January 2006"),
	}, nil
}

func (s *userService) ChangeSubscription(ctx context.Context, userID primitive.ObjectID, subscription string) error {
	if _, exists := models.SubscriptionLimits[subscription]; !exists {
		return models.ErrInvalidSubscription
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Subscription = subscription
	return s.userRepo.UpdateUser(ctx, user)
}

func (s *userService) GetSubscriptionLimits(subscription string) models.SubscriptionConfig {
	limits, exists := models.SubscriptionLimits[subscription]
	if !exists {
		return models.SubscriptionLimits["free"]
	}
	return limits
}

func (s *userService) IncrementAnalysesCount(ctx context.Context, userID primitive.ObjectID) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.AnalysesCount++
	return s.userRepo.UpdateUser(ctx, user)
}
