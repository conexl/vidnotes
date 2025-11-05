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

type AISessionRepository interface {
	Create(ctx context.Context, session *models.AISession) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.AISession, error)
	GetByVideoID(ctx context.Context, videoID primitive.ObjectID) ([]*models.AISession, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.AISession, error)
	AddMessage(ctx context.Context, sessionID primitive.ObjectID, message models.AIMessage) error
	UpdateTitle(ctx context.Context, sessionID primitive.ObjectID, title string) error
	//	UpdateSummary(ctx context.Context, sessionID primitive.ObjectID, summary string) error
	Delete(ctx context.Context, sessionID primitive.ObjectID) error
}

type aiSessionRepository struct {
	collection *mongo.Collection
}

func NewAISessionRepository(db *mongo.Database) AISessionRepository {
	return &aiSessionRepository{
		collection: db.Collection("ai_sessions"),
	}
}

func (r *aiSessionRepository) Create(ctx context.Context, session *models.AISession) (primitive.ObjectID, error) {
	session.CreatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, session)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("%w: %v", models.ErrSessionCreateFailed, err)
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("%w: failed to convert inserted ID", models.ErrSessionCreateFailed)
	}

	session.ID = insertedID
	return insertedID, nil
}

func (r *aiSessionRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.AISession, error) {
	var session models.AISession

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, models.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

func (r *aiSessionRepository) GetByVideoID(ctx context.Context, videoID primitive.ObjectID) ([]*models.AISession, error) {
	var sessions []*models.AISession

	cursor, err := r.collection.Find(ctx, bson.M{"video_id": videoID})
	if err != nil {
		return nil, fmt.Errorf("failed to find sessions: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, fmt.Errorf("failed to decode sessions: %w", err)
	}

	return sessions, nil
}

func (r *aiSessionRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.AISession, error) {
	var sessions []*models.AISession

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to find sessions: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, fmt.Errorf("failed to decode sessions: %w", err)
	}

	return sessions, nil
}

func (r *aiSessionRepository) AddMessage(ctx context.Context, sessionID primitive.ObjectID, message models.AIMessage) error {
	message.Time = time.Now()

	if message.Role == "" {
		return models.ErrInvalidMessageRole
	}

	update := bson.M{
		"$push": bson.M{
			"messages": message,
		},
	}

	result, err := r.collection.UpdateByID(ctx, sessionID, update)
	if err != nil {
		return fmt.Errorf("%w: %v", models.ErrSessionUpdateFailed, err)
	}

	if result.MatchedCount == 0 {
		return models.ErrSessionNotFound
	}

	return nil
}

func (r *aiSessionRepository) UpdateTitle(ctx context.Context, sessionID primitive.ObjectID, title string) error {
	update := bson.M{
		"$set": bson.M{
			"title": title,
		},
	}

	result, err := r.collection.UpdateByID(ctx, sessionID, update)
	if err != nil {
		return fmt.Errorf("%w: %v", models.ErrSessionUpdateFailed, err)
	}

	if result.MatchedCount == 0 {
		return models.ErrSessionNotFound
	}

	return nil
}

func (r *aiSessionRepository) Delete(ctx context.Context, sessionID primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": sessionID})
	if err != nil {
		return fmt.Errorf("%w: %v", models.ErrSessionDeleteFailed, err)
	}

	if result.DeletedCount == 0 {
		return models.ErrSessionNotFound
	}

	return nil
}
