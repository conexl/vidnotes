// repository/video_repository.go
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

type VideoRepository interface {
	Create(ctx context.Context, video *models.Video) (primitive.ObjectID, error)
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error
	UpdateSummary(ctx context.Context, id primitive.ObjectID, summary string) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Video, error)
	GetByUser(ctx context.Context, userID primitive.ObjectID) ([]*models.Video, error)
	Delete(ctx context.Context, videoID primitive.ObjectID) error
}

type videoRepository struct {
	collection *mongo.Collection
}

func NewVideoRepository(db *mongo.Database) VideoRepository {
	return &videoRepository{
		collection: db.Collection("videos"),
	}
}

func (r *videoRepository) Create(ctx context.Context, video *models.Video) (primitive.ObjectID, error) {
	video.CreatedAt = time.Now()
	video.UpdatedAt = time.Now()
	video.Status = "uploaded"

	result, err := r.collection.InsertOne(ctx, video)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("%w: %v", models.ErrVideoCreateFailed, err)
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("%w: failed to convert inserted ID", models.ErrVideoCreateFailed)
	}

	video.ID = insertedID
	return insertedID, nil
}

func (r *videoRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateByID(ctx, id, update)
	if err != nil {
		return fmt.Errorf("%w: %v", models.ErrVideoUpdateFailed, err)
	}

	if result.MatchedCount == 0 {
		return models.ErrVideoNotFound
	}

	return nil
}

func (r *videoRepository) UpdateSummary(ctx context.Context, id primitive.ObjectID, summary string) error {
	update := bson.M{
		"$set": bson.M{
			"summary":    summary,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateByID(ctx, id, update)
	if err != nil {
		return fmt.Errorf("%w: %v", models.ErrVideoUpdateFailed, err)
	}

	if result.MatchedCount == 0 {
		return models.ErrVideoNotFound
	}

	return nil
}

func (r *videoRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Video, error) {
	var video models.Video

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&video)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, models.ErrVideoNotFound
		}
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	return &video, nil
}

func (r *videoRepository) GetByUser(ctx context.Context, userID primitive.ObjectID) ([]*models.Video, error) {
	var videos []*models.Video

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to find videos: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &videos); err != nil {
		return nil, fmt.Errorf("failed to decode videos: %w", err)
	}

	return videos, nil
}

func (r *videoRepository) Delete(ctx context.Context, videoID primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": videoID})
	if err != nil {
		return fmt.Errorf("%w: %v", models.ErrVideoDeleteFailed, err)
	}

	if result.DeletedCount == 0 {
		return models.ErrVideoNotFound
	}

	return nil
}
