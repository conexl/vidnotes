package handlers

import (
	"time"

	"github.com/code-zt/vidnotes/internal/models"
	"github.com/code-zt/vidnotes/internal/repository"
	"github.com/code-zt/vidnotes/internal/services"
	"github.com/code-zt/vidnotes/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AIHandlers struct {
	aiService   services.AIService
	sessionRepo repository.AISessionRepository
	videoRepo   repository.VideoRepository
}

func NewAIHandlers(aiService services.AIService, sessionRepo repository.AISessionRepository, videoRepo repository.VideoRepository) *AIHandlers {
	return &AIHandlers{
		aiService:   aiService,
		sessionRepo: sessionRepo,
		videoRepo:   videoRepo,
	}
}

type CreateSessionRequest struct {
	VideoID string `json:"video_id"`
	Title   string `json:"title"`
}

func (h *AIHandlers) CreateSession(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	var req CreateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	videoID, err := primitive.ObjectIDFromHex(req.VideoID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid video ID format")
	}

	video, err := h.videoRepo.GetByID(c.Context(), videoID)
	if err != nil {
		return utils.Error(c, fiber.StatusNotFound, "Video not found")
	}

	if video.UserID != userObjectID {
		return utils.Error(c, fiber.StatusForbidden, "Access denied")
	}
	processedSummary, err := h.aiService.ImproveSummary(c.Context(), video.Summary, []string{"Delete all non-essential content: irrelevant text, personal info, or details that don’t support the main context. Keep only key facts, direct context, and critical info. Return filtered content concisely."})
	if err != nil {
		processedSummary = video.Summary
	}

	session := &models.AISession{
		UserID:  userObjectID,
		VideoID: videoID,
		Title:   req.Title,
		Summary: processedSummary,
		Messages: []models.AIMessage{
			{
				Role:    "system",
				Content: "Сессия AI ассистента для анализа видео создана.",
				Time:    time.Now(),
			},
		},
	}

	sessionID, err := h.sessionRepo.Create(c.Context(), session)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	session.ID = sessionID
	return utils.Success(c, fiber.StatusCreated, session)
}

type SendMessageRequest struct {
	Message string `json:"message"`
}

func (h *AIHandlers) SendMessage(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	sessionID, err := h.getValidSessionID(c)
	if err != nil {
		return err
	}

	var req SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Message == "" {
		return utils.Error(c, fiber.StatusBadRequest, "Message cannot be empty")
	}

	session, err := h.validateSessionAccess(c, sessionID, userID)
	if err != nil {
		return err
	}

	userMessage := models.AIMessage{
		Role:    "user",
		Content: req.Message,
		Time:    time.Now(),
	}

	if err := h.sessionRepo.AddMessage(c.Context(), sessionID, userMessage); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to save user message")
	}

	aiResponse, err := h.aiService.SendMessage(c.Context(), session, req.Message)
	if err != nil {
		errorMessage := models.AIMessage{
			Role:    "assistant",
			Content: "Извините, произошла ошибка при обработке запроса. Пожалуйста, попробуйте позже.",
			Time:    time.Now(),
		}
		h.sessionRepo.AddMessage(c.Context(), sessionID, errorMessage)
		return utils.Error(c, fiber.StatusInternalServerError, "AI service error: "+err.Error())
	}

	aiMessage := models.AIMessage{
		Role:    "assistant",
		Content: aiResponse,
		Time:    time.Now(),
	}

	if err := h.sessionRepo.AddMessage(c.Context(), sessionID, aiMessage); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to save AI response")
	}

	response := models.AIResponse{
		Message:   aiResponse,
		SessionID: sessionID.Hex(),
		Time:      aiMessage.Time,
	}

	return utils.Success(c, fiber.StatusOK, response)
}

func (h *AIHandlers) GetSession(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	sessionID, err := h.getValidSessionID(c)
	if err != nil {
		return err
	}

	session, err := h.validateSessionAccess(c, sessionID, userID)
	if err != nil {
		return err
	}

	return utils.Success(c, fiber.StatusOK, session)
}

func (h *AIHandlers) GetUserSessions(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	sessions, err := h.sessionRepo.GetByUserID(c.Context(), userObjectID)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to get sessions")
	}

	return utils.Success(c, fiber.StatusOK, sessions)
}

func (h *AIHandlers) DeleteSession(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	sessionID, err := h.getValidSessionID(c)
	if err != nil {
		return err
	}

	_, err = h.validateSessionAccess(c, sessionID, userID)
	if err != nil {
		return err
	}

	if err := h.sessionRepo.Delete(c.Context(), sessionID); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to delete session")
	}

	return utils.Success(c, fiber.StatusOK, fiber.Map{
		"message": "Session deleted successfully",
	})
}

func (h *AIHandlers) getValidSessionID(c *fiber.Ctx) (primitive.ObjectID, error) {
	sessionID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return primitive.NilObjectID, utils.Error(c, fiber.StatusBadRequest, "Invalid session ID")
	}
	return sessionID, nil
}

func (h *AIHandlers) validateSessionAccess(c *fiber.Ctx, sessionID primitive.ObjectID, userID string) (*models.AISession, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, utils.Error(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	session, err := h.sessionRepo.GetByID(c.Context(), sessionID)
	if err != nil {
		return nil, utils.Error(c, fiber.StatusNotFound, "Session not found")
	}

	if session.UserID != userObjectID {
		return nil, utils.Error(c, fiber.StatusForbidden, "Access denied")
	}

	return session, nil
}
