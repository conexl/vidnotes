package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/code-zt/vidnotes/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) (primitive.ObjectID, error)
	GetUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id primitive.ObjectID) error
	UserExists(ctx context.Context, email string) (bool, error)
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
	}
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User) (primitive.ObjectID, error) {
	user.CreatedAt = time.Now()
	if user.Subscription == "" {
		user.Subscription = "free"
	}
	if user.Role == "" {
		user.Role = "user"
	}

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return primitive.NilObjectID, models.ErrUserAlreadyExists
		}
		return primitive.NilObjectID, fmt.Errorf("%w: %v", models.ErrUserCreateFailed, err)
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("%w: failed to convert inserted ID", models.ErrUserCreateFailed)
	}

	user.ID = insertedID
	return insertedID, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user *models.User) error {
	update := bson.M{
		"$set": bson.M{
			"name":           user.Name,
			"email":          user.Email,
			"avatar_url":     user.AvatarURL,
			"analyses_count": user.AnalysesCount,
			"subscription":   user.Subscription,
			"role":           user.Role,
			"last_active_at": time.Now(),
		},
	}
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		update,
	)
	if err != nil {
		return fmt.Errorf("%w: %v", models.ErrUserUpdateFailed, err)
	}

	if result.MatchedCount == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("%w: %v", models.ErrUserDeleteFailed, err)
	}

	if result.DeletedCount == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) UserExists(ctx context.Context, email string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return count > 0, nil
}
