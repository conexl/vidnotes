package handlers

import (
	"github.com/code-zt/vidnotes/internal/models"
	"github.com/code-zt/vidnotes/internal/services"
	"github.com/code-zt/vidnotes/pkg/auth"
	"github.com/code-zt/vidnotes/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserHandlers struct {
	userService services.UserService
	authManager *auth.JWTManager
}

func NewUserHandlers(userService services.UserService, authManager *auth.JWTManager) *UserHandlers {
	return &UserHandlers{
		userService: userService,
		authManager: authManager,
	}
}

func (h *UserHandlers) Register(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	userID, err := h.userService.Register(c.Context(), req)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	tokenPair, err := h.authManager.GenerateTokenPair(userID.Hex())
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to generate tokens")
	}

	return utils.Success(c, fiber.StatusCreated, fiber.Map{
		"user_id": userID,
		"tokens":  tokenPair,
	})
}

func (h *UserHandlers) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	userID, err := h.userService.Login(c.Context(), req)
	if err != nil {
		return utils.Error(c, fiber.StatusUnauthorized, err.Error())
	}

	tokenPair, err := h.authManager.GenerateTokenPair(userID.Hex())
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to generate tokens")
	}

	return utils.Success(c, fiber.StatusOK, fiber.Map{
		"user_id": userID,
		"tokens":  tokenPair,
	})
}

func (h *UserHandlers) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	user, err := h.userService.GetProfile(c.Context(), userObjectID)
	if err != nil {
		return utils.Error(c, fiber.StatusNotFound, "User not found")
	}

	return utils.Success(c, fiber.StatusOK, user)
}

func (h *UserHandlers) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.userService.UpdateProfile(c.Context(), userObjectID, req); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.Success(c, fiber.StatusOK, fiber.Map{
		"message": "Profile updated successfully",
	})
}

func (h *UserHandlers) ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return utils.Error(c, fiber.StatusBadRequest, "Old password and new password are required")
	}

	if len(req.NewPassword) < 6 {
		return utils.Error(c, fiber.StatusBadRequest, "New password must be at least 6 characters")
	}

	if err := h.userService.ChangePassword(c.Context(), userObjectID, req.OldPassword, req.NewPassword); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.Success(c, fiber.StatusOK, fiber.Map{
		"message": "Password changed successfully",
	})
}

func (h *UserHandlers) GetAnalytics(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid user ID")
	}

	analytics, err := h.userService.GetAnalyticsInfo(c.Context(), userObjectID)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.Success(c, fiber.StatusOK, analytics)
}

func (h *UserHandlers) RefreshToken(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	tokenPair, err := h.authManager.RefreshTokens(req.RefreshToken)
	if err != nil {
		return utils.Error(c, fiber.StatusUnauthorized, "Invalid refresh token")
	}

	return utils.Success(c, fiber.StatusOK, tokenPair)
}
